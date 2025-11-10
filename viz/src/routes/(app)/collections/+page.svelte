<script module lang="ts">
	export { collectionCard };
</script>

<script lang="ts">
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import type { PageProps } from "./$types";
	import CollectionCard, { openCollection } from "$lib/components/CollectionCard.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import Button from "$lib/components/Button.svelte";
	import { createCollection, type Collection } from "$lib/api";
	import ModalOverlay from "$lib/components/modal/ModalOverlay.svelte";
	import { modal, sort } from "$lib/states/index.svelte";
	import { goto } from "$app/navigation";
	import SliderToggle from "$lib/components/SliderToggle.svelte";
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

	let { data }: PageProps = $props();

	const pagination = $state({
		limit: data?.limit ?? 10,
		offset: data?.offset ?? 0
	});

	let listOfCollectionsData = $state(data?.items);

	let shouldUpdate = $derived(listOfCollectionsData?.length > pagination.limit * pagination.offset);
	let displayData = $derived(sortCollections(listOfCollectionsData, sort));

	let fadeOpacity = false;
	let toolbarOpacity = $state(1);

	// Selection
	let selectedAssets = $state<SvelteSet<any>>(new SvelteSet());
	let singleSelectedAsset: any | undefined = $state();

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
</script>

<CollectionModal
	heading="Create Collection"
	buttonText="Create"
	data={undefined as any}
	modalAction={async (event) => {
		const formData = new FormData(event.currentTarget);
		const name = formData.get("name") as string;
		const description = formData.get("description") as string;
		const isPrivate = formData.get("isPrivate") === "on";

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
	}}
/>

{#snippet collectionCard(collectionData: Collection)}
	{#if page.url.pathname !== "/"}
		<a
			data-sveltekit-preload-data
			data-asset-id={collectionData.uid}
			class="collection-card-link"
			href="/collections/{collectionData.uid}"
		>
			<CollectionCard collection={collectionData} />
		</a>
	{:else}
		<CollectionCard style={page.url.pathname === "/" ? "font-size: 0.9rem;" : ""} collection={collectionData} />
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
		pagination.offset++;
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
