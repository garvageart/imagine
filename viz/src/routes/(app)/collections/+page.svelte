<script lang="ts">
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import type { PageProps } from "./$types";
	import CollectionCard, { openCollection } from "$lib/components/CollectionCard.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import Button from "$lib/components/Button.svelte";
	import { createCollection, type Collection, deleteCollection, updateCollection } from "$lib/api";
	import { modal, sort } from "$lib/states/index.svelte";
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { SvelteSet } from "svelte/reactivity";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import type { Content } from "$lib/layouts/subpanel.svelte";
	import { getContext, type ComponentProps } from "svelte";
	import type { AssetGridArray } from "$lib/types/asset";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import { sortCollections } from "$lib/sort/sort";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import CollectionModal from "$lib/components/CollectionModal.svelte";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";

	let { data }: PageProps = $props();

	const pagination = $state({
		limit: data?.limit ?? 10,
		page: data?.page ?? 0
	});

	let listOfCollectionsData = $state(data?.items);

	let shouldUpdate = $derived(!!data?.next);
	let displayData = $derived(sortCollections(listOfCollectionsData, sort));

	let fadeOpacity = false;
	let toolbarOpacity = $state(1);

	// Selection
	let selectedAssets = $state<SvelteSet<Collection>>(new SvelteSet());
	let singleSelectedAsset: Collection | undefined = $state();

	let collectionGridArray: AssetGridArray<Collection> | undefined = $state();
	let grid: ComponentProps<typeof AssetGrid<Collection>> = $derived({
		assetDblClick: (_, asset) => {
			openCollection(asset, currentPanelContent);
		},
		assetSnippet: collectionCard,
		data: displayData,
		assetGridArray: collectionGridArray,
		selectedAssets,
		singleSelectedAsset
	});

	const currentPanelContent = getContext<Content>("content");

	// Context menu state for right-click on collections
	let ctxShowMenu = $state(false);
	let ctxItems = $state([] as any[]);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(null as any);

	// Modal data for create/edit
	let modalData: any = $state(undefined);
	let modalMode: "create" | "edit" = $state("create");

	function openCollectionContext(collection: Collection, anchor: { x: number; y: number } | HTMLElement) {
		// Build context menu items for a collection
		ctxItems = [
			{
				id: `open-${collection.uid}`,
				label: "Open",
				icon: "open_in_new",
				action: () => goto(`/collections/${collection.uid}`)
			},
			{
				id: `edit-${collection.uid}`,
				label: "Edit",
				icon: "edit",
				action: () => {
					modalMode = "edit";
					modalData = { ...collection };
					modal.show = true;
				}
			},
			{
				id: `duplicate-${collection.uid}`,
				label: "Duplicate",
				icon: "content_copy",
				action: async () => {
					try {
						const res = await createCollection({
							name: `Copy of ${collection.name}`,
							description: collection.description ?? undefined,
							private: collection.private ?? false
						});

						if (res.status === 201) {
							listOfCollectionsData = [res.data as Collection, ...listOfCollectionsData];
							toastState.addToast({ message: "Collection duplicated", type: "success" });
						} else {
							toastState.addToast({ message: (res as any).data?.error ?? `Duplicate failed (${res.status})`, type: "error" });
						}
					} catch (err) {
						toastState.addToast({ message: "Duplicate failed: " + (err as Error).message, type: "error" });
					}
				}
			},
			{ separator: true, id: `sep-${collection.uid}`, label: "" },
			{
				id: `copylink-${collection.uid}`,
				label: "Copy link",
				icon: "link",
				action: async () => {
					try {
						const url = `${location.origin}/collections/${collection.uid}`;
						await navigator.clipboard.writeText(url);
						toastState.addToast({ message: "Link copied to clipboard", type: "success" });
					} catch (err) {
						toastState.addToast({ message: "Failed to copy link", type: "error" });
					}
				}
			},
			{
				id: `delete-${collection.uid}`,
				label: "Delete",
				icon: "delete",
				danger: true,
				action: async () => {
					if (!confirm(`Delete collection "${collection.name}"? This cannot be undone.`)) return;
					try {
						const res = await deleteCollection(collection.uid);
						if (res.status === 204) {
							listOfCollectionsData = listOfCollectionsData.filter((c) => c.uid !== collection.uid);
							toastState.addToast({ message: `Deleted collection ${collection.name}`, type: "success" });
						} else {
							toastState.addToast({ message: res.data?.error ?? "Failed to delete", type: "error" });
						}
					} catch (err) {
						toastState.addToast({ message: `Failed to delete: ${err}`, type: "error" });
					}
				}
			}
		];

		ctxAnchor = anchor as any;
		ctxShowMenu = true;
	}
</script>

<CollectionModal
	heading={modalMode === "create" ? "Create Collection" : "Edit Collection"}
	buttonText={modalMode === "create" ? "Create" : "Save"}
	data={modalData}
	modalAction={async (event) => {
		const formData = new FormData(event.currentTarget);
		const name = formData.get("name") as string;
		const description = formData.get("description") as string;
		const isPrivate = formData.get("isPrivate") === "on";

		if (modalMode === "create") {
			try {
				const res = await createCollection({ name, description, private: isPrivate });
				if (res.status === 201) {
					listOfCollectionsData = [res.data as Collection, ...listOfCollectionsData];
					modal.show = false;
					toastState.addToast({ message: "Collection created", type: "success" });
					goto(`/collections/${(res.data as Collection).uid}`);
				} else {
					toastState.addToast({ message: (res as any).data?.error ?? `Creation failed (${res.status})`, type: "error" });
				}
			} catch (e) {
				toastState.addToast({ message: "Creation failed: " + (e as Error).message, type: "error" });
			}
		} else {
			// edit
			try {
				if (!modalData || !modalData.uid) return;
				const res = await updateCollection(modalData.uid, { name, description, private: isPrivate });
				if (res.status === 200) {
					// update local list
					listOfCollectionsData = listOfCollectionsData.map((c) => (c.uid === modalData.uid ? res.data : c));
					modal.show = false;
					toastState.addToast({ message: "Collection updated", type: "success" });
				} else {
					toastState.addToast({ message: res.data?.error ?? `Update failed (${res.status})`, type: "error" });
				}
			} catch (e) {
				toastState.addToast({ message: "Update failed: " + (e as Error).message, type: "error" });
			}
		}
	}}
/>

{#snippet collectionCard(collectionData: Collection)}
	{#if page.url.pathname !== "/"}
		<a
			data-sveltekit-preload-data
			data-asset-id={collectionData.uid}
			class="collection-card-link"
			href="/collections/{collectionData.uid}"
			oncontextmenu={(e) => {
				e.preventDefault();
				openCollectionContext(collectionData, { x: e.clientX, y: e.clientY });
			}}
		>
			<CollectionCard collection={collectionData} />
		</a>
	{:else}
		<CollectionCard
			style={page.url.pathname === "/" ? "font-size: 0.9rem;" : ""}
			collection={collectionData}
			oncontextmenu={(e) => {
				e.preventDefault();
				openCollectionContext(collectionData, e.currentTarget as HTMLElement);
			}}
		/>
	{/if}
{/snippet}

{#snippet toolbarSnippet()}
	<div id="coll-details-toolbar">
		<div id="coll-tools">
			<Button
				id="create-collection"
				class="toolbar-button"
				style="font-size: 0.8rem; background-color: var(--imag-80);"
				title="Create Collection"
				aria-label="Create Collection"
				onclick={() => {
					modalMode = "create";
					modalData = undefined;
					modal.show = true;
				}}
			>
				Create
				<MaterialIcon iconName="add" />
			</Button>
		</div>
		<span id="coll-details-floating"
			>{listOfCollectionsData.length}
			{listOfCollectionsData.length === 1 ? "collection" : "collections"}</span
		>
	</div>
{/snippet}

{#snippet noAssetsSnippet()}
	<div id="no-collections-container">
		<div id="no-collections-text">
			<MaterialIcon iconName="folder_open" style="font-size: 2rem; margin: 0rem 0.5rem; color: var(--imag-20);" />
			<span style="color: var(--imag-20); font-size: 1.2rem;">Add your first collection</span>
		</div>

		<Button
			id="create_collection-button"
			style="padding: 2em 8em; display: flex; align-items: center; justify-content: center;"
			title="Create Collection"
			aria-label="Create Collection"
			onclick={() => {
				modalMode = "create";
				modalData = undefined;
				modal.show = true;
			}}
		>
			Create a New Collection
			<MaterialIcon iconName="add" style="font-size: 2em;" />
		</Button>
	</div>
{/snippet}

<VizViewContainer
	bind:data={displayData}
	bind:hasMore={shouldUpdate}
	name="Collections"
	paginate={() => {
		pagination.page++;
	}}
	onscroll={(e) => {
		if (!fadeOpacity) {
			return;
		}

		const assetGrid = document.querySelector(".viz-asset-grid-container")! as HTMLElement;

		const top = Math.max(assetGrid.offsetTop, 1);
		const current = e.currentTarget.scrollTop;
		if (current >= top) {
			toolbarOpacity = 0;
		} else {
			toolbarOpacity = Math.max(0, Math.min(1, 1 - current / top));
		}
	}}
>
	<AssetsShell
		bind:grid
		{pagination}
		{toolbarSnippet}
		{noAssetsSnippet}
		toolbarProps={{
			style: `justify-content: space-between;` + (fadeOpacity ? `opacity: ${toolbarOpacity};` : "")
		}}
	/>
</VizViewContainer>

<!-- Context menu for right-click on collections -->
<ContextMenu bind:showMenu={ctxShowMenu} items={ctxItems} anchor={ctxAnchor} offsetY={4} />

<style lang="scss">
	@use "sass:color";

	.collection-card-link {
		cursor: context-menu;
	}

	:global(.toolbar-button) {
		border-radius: 10em;
		margin: 0.5em 0em;
		/* Use rem so button font-size is consistent site-wide (not relative to local container) */
		font-size: 0.9rem;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--imag-text-color);
		cursor: pointer;
	}

	#coll-details-toolbar {
		width: 100%;
		display: flex;
		justify-content: left;
	}

	#coll-details-floating {
		color: var(--imag-40);
		background-color: transparent;
		margin: 0.5rem 0rem;
		font-family: var(--imag-code-font);
	}

	#coll-tools {
		display: flex;
		align-items: center;
		height: 100%;
	}

	#no-collections-text {
		display: flex;
		flex-direction: row;
		justify-content: center;
		align-items: center;
		margin-bottom: 1rem;
	}
</style>
