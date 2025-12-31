<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import {
		createCollection,
		deleteCollection,
		updateCollection,
		type Collection
	} from "$lib/api";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import Button from "$lib/components/Button.svelte";
	import CollectionCard, {
		openCollection
	} from "$lib/components/CollectionCard.svelte";
	import IconButton from "$lib/components/IconButton.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import CollectionModal from "$lib/components/modals/CollectionModal.svelte";
	import FilterModal from "$lib/components/modals/FilterModal.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import { createCollectionMenu } from "$lib/context-menu/menus/collections";
	import { TabGroup } from "$lib/layouts/model.svelte";
	import { sortCollections } from "$lib/sort/sort";
	import { filterManager } from "$lib/states/filter.svelte";
	import { isLayoutPage, modal, sort } from "$lib/states/index.svelte";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import type { AssetGridArray } from "$lib/types/asset";
	import { getContext, untrack, type ComponentProps } from "svelte";
	import type { PageProps } from "./$types";
	import {
		selectionManager,
		SelectionScopeNames
	} from "$lib/states/selection.svelte";

	let { data }: PageProps = $props();

	let showFilterModal = $state(false);
	let showCollectionModal = $state(false);

	$effect(() => {
		untrack(() => {
			filterManager.setActiveScopeType("collections");
			if (!filterManager.keepFilters) {
				filterManager.resetActiveScope();
			}
		});
	});

	const pagination = $derived({
		limit: data.limit ?? 50,
		page: data.page ?? 0
	});

	let listOfCollectionsData = $derived(data?.items ?? []);

	let shouldUpdate = $derived(!!data.next);
	let displayData = $derived(sortCollections(listOfCollectionsData, sort));

	let fadeOpacity = false;
	let toolbarOpacity = $state(1);

	// Selection
	const scopeId = $derived(SelectionScopeNames.COLLECTIONS_MAIN);
	const selectionScope = $derived(
		selectionManager.getScope<Collection>(scopeId)
	);

	let collectionGridArray: AssetGridArray<Collection> | undefined = $state();
	let grid: ComponentProps<typeof AssetGrid<Collection>> = $derived({
		assetDblClick: (_, asset) => {
			openCollection(asset, currentPanelContent);
		},
		assetSnippet: collectionCard,
		data: displayData,
		assetGridArray: collectionGridArray,
		scopeId,
		assetGridDisplayProps: {
			style: `padding: 0em ${isLayoutPage() ? "1em" : "2em"};`
		},
		onassetcontext: (detail) => {
			openCollectionContext(detail.asset, detail.anchor);
		}
	});

	const currentPanelContent = getContext<TabGroup>("content");

	// Context menu state for right-click on collections
	let ctxShowMenu = $state(false);
	let ctxItems = $state([] as any[]);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(
		null as any
	);

	// Modal data for create/edit
	let modalData: Collection | undefined = $state();
	let modalMode: "create" | "edit" = $state("create");

	function openCollectionContext(
		collection: Collection,
		anchor: { x: number; y: number } | HTMLElement
	) {
		ctxItems = createCollectionMenu(collection, {
			editCollection: (col) => {
				modalMode = "edit";
				modalData = { ...col };
				modal.show = true;
			},
			onCollectionDuplicated: (newCol) => {
				listOfCollectionsData = [newCol, ...listOfCollectionsData];
				toastState.addToast({
					message: `Duplicated collection ${newCol.name}`,
					type: "success"
				});
			},
			onCollectionDeleted: (deletedCol) => {
				listOfCollectionsData = listOfCollectionsData.filter(
					(c) => c.uid !== deletedCol.uid
				);
				toastState.addToast({
					message: `Deleted collection ${deletedCol.name}`,
					type: "success"
				});
			}
		});

		ctxAnchor = anchor;
		ctxShowMenu = true;
	}

	async function handleDeleteSelected() {
		const items = Array.from(selectionScope.selected ?? []);
		if (items.length === 0) {
			return;
		}

		const ok = confirm(
			`Remove ${items.length} selected collections(s)? This cannot be undone!`
		);
		if (!ok) {
			return;
		}

		const uids = items.map((i: Collection) => i.uid);
		try {
			const deletePromises = uids.map((uid) => deleteCollection(uid));
			const res = await Promise.all(deletePromises);

			const successful = res.filter((r) => r.status === 204);
			const failed = res.filter((r) => r.status !== 204);

			toastState.addToast({
				type: "success",
				message: `Deleted ${successful.length} collection(s)`
			});

			if (failed.length > 0) {
				toastState.addToast({
					type: "error",
					message: `Failed to delete ${failed.length} collection(s)`
				});
			}
		} catch (err) {
			toastState.addToast({
				type: "error",
				message: `Failed to remove images: ${err}`
			});
		}
	}
</script>

{#if showCollectionModal && modal.show}
	<CollectionModal
		heading={modalMode === "create" ? "Create Collection" : "Edit Collection"}
		buttonText={modalMode === "create" ? "Create" : "Save"}
		bind:data={modalData}
		modalAction={async (event) => {
			const formData = new FormData(event.currentTarget);
			const name = formData.get("name") as string;
			const description = formData.get("description") as string;
			const isPrivate = formData.get("isPrivate") === "on";

			if (modalMode === "create") {
				try {
					const res = await createCollection({
						name,
						description,
						private: isPrivate
					});
					if (res.status === 201) {
						listOfCollectionsData = [res.data, ...listOfCollectionsData];
						modal.show = false;
						toastState.addToast({
							message: "Collection created",
							type: "success"
						});
						goto(`/collections/${res.data.uid}`);
					} else {
						toastState.addToast({
							message:
								(res as any).data?.error ?? `Creation failed (${res.status})`,
							type: "error"
						});
					}
				} catch (e) {
					toastState.addToast({
						message: "Creation failed: " + (e as Error).message,
						type: "error"
					});
				}
			} else {
				// edit
				try {
					if (!modalData || !modalData.uid) {
						return;
					}
					const res = await updateCollection(modalData.uid, {
						name,
						description,
						private: isPrivate
					});
					if (res.status === 200) {
						// update local list
						listOfCollectionsData = listOfCollectionsData.map((c) =>
							c.uid === modalData!.uid ? res.data : c
						);
						modal.show = false;
						toastState.addToast({
							message: "Collection updated",
							type: "success"
						});
					} else {
						toastState.addToast({
							message: res.data?.error ?? `Update failed (${res.status})`,
							type: "error"
						});
					}
				} catch (e) {
					toastState.addToast({
						message: "Update failed: " + (e as Error).message,
						type: "error"
					});
				}
			}
		}}
	/>
{/if}

{#if showFilterModal && modal.show}
	<FilterModal />
{/if}

{#snippet collectionCard(collectionData: Collection)}
	{#if page.url.pathname !== "/"}
		<a
			title="{collectionData.name} | Last Updated: {new Date(
				collectionData.updated_at
			).toLocaleDateString()}"
			data-sveltekit-preload-data
			data-asset-id={collectionData.uid}
			class="collection-card-link"
			href="/collections/{collectionData.uid}"
		>
			<CollectionCard collection={collectionData} />
		</a>
	{:else}
		<CollectionCard
			style={isLayoutPage() ? "font-size: 0.9rem;" : ""}
			collection={collectionData}
		/>
	{/if}
{/snippet}

{#snippet toolbarSnippet()}
	<div id="coll-details-toolbar">
		<div id="coll-tools">
			{#if !isLayoutPage()}
				<IconButton
					iconName="filter_list"
					class="toolbar-button"
					title="Filter"
					aria-label="Filter"
					onclick={() => {
						showFilterModal = true;
						modal.show = true;
					}}
				>
					Filter
				</IconButton>
			{/if}
			<IconButton
				iconName="add"
				id="create-collection"
				title="Create Collection"
				aria-label="Create Collection"
				onclick={() => {
					modalMode = "create";
					modalData = undefined;
					showCollectionModal = true;
					modal.show = true;
				}}
			>
				Create
			</IconButton>
			<span id="coll-details-floating"
				>{#if listOfCollectionsData}{listOfCollectionsData.length}{/if}
				{listOfCollectionsData?.length === 1
					? "collection"
					: "collections"}</span
			>
		</div>
	</div>
{/snippet}

{#snippet selectionToolbarSnippet()}
	<IconButton
		iconName="delete"
		title="Delete Selected"
		style="position: absolute; right: 1em; background-color: var(--imag-100);"
		onclick={handleDeleteSelected}
	/>
{/snippet}

{#snippet noAssetsSnippet()}
	<div id="no-collections-container">
		<div id="no-collections-text">
			<MaterialIcon
				iconName="folder_open"
				style="font-size: 2rem; margin: 0rem 0.5rem; color: var(--imag-20);"
			/>
			<span style="color: var(--imag-20); font-size: 1.2rem;"
				>Add your first collection</span
			>
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
	style="padding: 0em ${page.url.pathname === '/' ? '1em' : '2em'};"
	name="Collections"
	paginate={() => {
		pagination.page++;
	}}
	onscroll={(e) => {
		if (!fadeOpacity) {
			return;
		}

		const assetGrid = document.querySelector(
			".viz-asset-grid-container"
		)! as HTMLElement;

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
		gridComponent={AssetGrid}
		{pagination}
		{selectionToolbarSnippet}
		{toolbarSnippet}
		{noAssetsSnippet}
		toolbarProps={{
			style:
				`justify-content: space-between;` +
				(fadeOpacity ? `opacity: ${toolbarOpacity};` : "")
		}}
	/>
</VizViewContainer>

<!-- Context menu for right-click on collections -->
<ContextMenu
	bind:showMenu={ctxShowMenu}
	items={ctxItems}
	anchor={ctxAnchor}
	offsetY={4}
/>

<style lang="scss">
	@use "sass:color";

	.collection-card-link {
		cursor: context-menu;
	}

	#coll-details-toolbar {
		height: 100%;
		display: flex;
	}

	#coll-details-floating {
		color: var(--imag-40);
		background-color: transparent;
		font-family: var(--imag-code-font);
	}

	#coll-tools {
		display: flex;
		align-items: center;
	}

	#no-collections-text {
		display: flex;
		flex-direction: row;
		justify-content: center;
		align-items: center;
		margin-bottom: 1rem;
	}
</style>
