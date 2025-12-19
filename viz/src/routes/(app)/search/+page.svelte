<script lang="ts">
	import { goto } from "$app/navigation";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import AssetToolbar from "$lib/components/AssetToolbar.svelte";
	import LoadingSpinner from "$lib/components/LoadingSpinner.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { performSearch } from "$lib/search/execute";
	import { modal, search, sort } from "$lib/states/index.svelte";
	import { onMount, type ComponentProps } from "svelte";
	import { SvelteSet } from "svelte/reactivity";
	import CollectionCard from "$lib/components/CollectionCard.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import PhotoAssetGrid, {
		type AssetGridView
	} from "$lib/components/PhotoAssetGrid.svelte";
	import ImageLightbox from "$lib/components/ImageLightbox.svelte";
	import type { Collection, Image } from "$lib/api";
	import hotkeys from "hotkeys-js";
	import {
		getConsolidatedGroups,
		groupImagesByDate,
		type ConsolidatedGroup,
		type DateGroup,
		type ImageWithDateLabel
	} from "$lib/photo-layout";
	import { DateTime } from "luxon";
	import { downloadOriginalImageFile } from "$lib/utils/http";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import Dropdown, {
		type DropdownOption
	} from "$lib/components/Dropdown.svelte";

	let collections = $derived(search.data.collections.data);
	let images = $derived(search.data.images.data);
	let totalResults = $derived(collections.length + images.length);

	let timeFound = $state(0);

	// Lightbox
	let lightboxImage: Image | undefined = $state();

	// Selection
	let selectedAssets = $state(new SvelteSet<Image>()); // For images
	let singleSelectedAsset: Image | undefined = $state();

	let selectedCollections = $state(new SvelteSet<Collection>()); // For collections
	let singleSelectedCollection: Collection | undefined = $state();

	// Display state
	let currentAssetGridView: AssetGridView = $state("grid");
	const displayOptions: DropdownOption[] = [
		{ title: "Grid" },
		{ title: "List" },
		{ title: "Card" }
	];

	// Grouping Logic for Images (Reused from photos/+page.svelte)
	let groups: DateGroup[] = $derived(groupImagesByDate(images) ?? []);

	let consolidatedGroups: ConsolidatedGroup[] = $derived.by(() => {
		return getConsolidatedGroups(groups);
	});

	let allImagesFlat = $derived(consolidatedGroups.flatMap((g) => g.allImages));

	// Grid props for Collections (using AssetsShell/AssetGrid)
	let collectionsGrid: ComponentProps<typeof AssetGrid<Collection>> = $derived({
		data: collections,
		assetSnippet: collectionCard,
		selectedAssets: selectedCollections,
		singleSelectedAsset: singleSelectedCollection,
		searchValue: search.value
	});

	hotkeys("escape", (e) => {
		e.preventDefault();
		selectedAssets.clear();
		selectedCollections.clear();

		singleSelectedAsset = undefined;
		singleSelectedCollection = undefined;
		lightboxImage = undefined;

		if (modal.show) {
			modal.show = false;
		}
	});

	function openLightbox(asset: Image) {
		lightboxImage = asset;
	}

	function nextLightboxImage() {
		if (!lightboxImage || allImagesFlat.length === 0) {
			return;
		}

		const idx = allImagesFlat.findIndex((i) => i.uid === lightboxImage!.uid);
		if (idx === -1) {
			return;
		}

		const nextIdx = (idx - 1 + allImagesFlat.length) % allImagesFlat.length;
		lightboxImage = allImagesFlat[nextIdx];
	}

	function prevLightboxImage() {
		if (!lightboxImage || allImagesFlat.length === 0) {
			return;
		}

		const idx = allImagesFlat.findIndex((i) => i.uid === lightboxImage!.uid);
		if (idx === -1) {
			return;
		}

		const nextIdx = (idx + 1) % allImagesFlat.length;
		lightboxImage = allImagesFlat[nextIdx];
	}

	async function performImageDownloads(uids: string[]) {
		// Reusing logic from photos page (simplified for single download for now, bulk needs token endpoint logic)
		// For search page, let's just support single download via helper if bulk not fully set up here
		if (uids.length === 1) {
			const img = images.find((i) => i.uid === uids[0]);
			if (img) await downloadOriginalImageFile(img);
		} else {
			toastState.addToast({
				type: "info",
				message: "Bulk download not fully implemented on search page yet",
				timeout: 3000
			});
		}
	}

	onMount(() => {
		if (search.value) {
			const startTime = performance.now();
			performSearch().then(() => {
				const endTime = performance.now();
				timeFound = endTime - startTime;
			});
		}
	});
</script>

<svelte:head>
	<title>Search{search.value ? ` - ${search.value}` : ""}</title>
</svelte:head>

{#if lightboxImage}
	<ImageLightbox bind:lightboxImage {nextLightboxImage} {prevLightboxImage} />
{/if}

{#snippet imageCard(asset: Image)}
	<ImageCard {asset} />
{/snippet}

{#snippet collectionCard(collectionData: Collection)}
	<a
		data-sveltekit-preload-data
		data-asset-id={collectionData.uid}
		class="collection-card-link"
		href="/collections/{collectionData.uid}"
	>
		<CollectionCard collection={collectionData} />
	</a>
{/snippet}

<div id="search">
	<div id="search-info-container" class="selection-container">
		{#if selectedAssets.size > 1 || selectedCollections.size > 1}
			<AssetToolbar class="asset-toolbar">
				<button
					id="coll-clear-selection"
					title="Clear selection"
					aria-label="Clear selection"
					style="margin-right: 0.5em;"
					class="toolbar-button"
					onclick={() => {
						selectedAssets.clear();
						selectedCollections.clear();
					}}
				>
					<MaterialIcon iconName="close" />
				</button>
				<span style="font-weight: 600;"
					>{selectedAssets.size + selectedCollections.size} selected</span
				>
			</AssetToolbar>
		{:else if !search.loading}
			<AssetToolbar
				style="position: sticky; top: 0px; display: flex; justify-content: flex-start;"
				class="asset-toolbar"
			>
				<p>
					{totalResults} results found in {(timeFound / 1000).toFixed(2)} seconds
				</p>
				<div class="search-info-details">
					<p>
						{collections.length} collection{collections.length === 1
							? ""
							: "s"}, {images.length}
						image{images.length === 1 ? "" : "s"}
					</p>
				</div>
			</AssetToolbar>
		{/if}
	</div>

	<div class="search-container no-select">
		{#if search.loading}
			<div class="loading-container">
				<p id="search-loading-text">Searching for "{search.value}"...</p>
				<LoadingSpinner />
			</div>
		{:else if search.executed}
			<div class="results">
				{#if totalResults === 0}
					<div class="no-results">
						<p>No results found for "{search.value}"</p>
					</div>
				{:else}
					{#if collections.length > 0}
						<section class="collections-section">
							<h2>Collections ({collections.length})</h2>
							<AssetsShell
								toolbarProps={{
									style: "justify-content: right;"
								}}
								bind:grid={collectionsGrid}
							/>
						</section>
					{/if}
					{#if images.length > 0}
						<section class="images-section">
							<div
								style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--imag-20); padding-right: 1rem;"
							>
								<h2>Images ({images.length})</h2>
								{#if selectedAssets.size <= 1}
									<div style="display: flex; align-items: center; gap: 0.5rem;">
										<Dropdown
											title="Display"
											class="toolbar-button"
											icon="list_alt"
											options={displayOptions}
											selectedOption={displayOptions.find(
												(o) => o.title.toLowerCase() === currentAssetGridView
											)}
											onSelect={(opt) => {
												currentAssetGridView = opt.title as AssetGridView;
											}}
										/>
									</div>
								{/if}
							</div>

							<div class="photo-group-container">
								{#each consolidatedGroups as consolidatedGroup}
									<section class="photo-group">
										<h2 class="photo-group-label">{consolidatedGroup.label}</h2>
										<PhotoAssetGrid
											{selectedAssets}
											bind:singleSelectedAsset
											bind:allData={allImagesFlat}
											bind:view={currentAssetGridView}
											data={consolidatedGroup.allImages}
											assetDblClick={(_e, asset) => openLightbox(asset)}
											onassetcontext={(detail: {
												asset: Image;
												anchor: { x: number; y: number };
											}) => {
												const { asset } = detail;
												if (
													!selectedAssets.has(asset) ||
													selectedAssets.size <= 1
												) {
													singleSelectedAsset = asset;
													selectedAssets.clear();
													selectedAssets.add(asset);
												}
												// Context menu not fully implemented in search page yet,
												// but selection works for basic actions.
											}}
										/>
									</section>
								{/each}
							</div>
						</section>
					{/if}
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	#search {
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		align-items: center;
	}

	#search-info-container {
		z-index: 1;
		width: 100%;
		position: sticky;
		top: 0;
		background-color: var(--imag-100);
	}

	:global(.asset-toolbar) {
		gap: 1rem;
	}

	.search-container {
		white-space: wrap;
		width: 100%;

		/* h1 {
			margin: 0.5rem 0rem;
			font-weight: 800;
			text-align: center;
		} */
	}

	.loading-container,
	.no-results {
		text-align: center;
		padding: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-direction: column;
	}

	#search-loading-text {
		font-size: 1em;
	}

	.results {
		width: 100%;
		flex-grow: 1;
	}

	.collections-section,
	.images-section {
		position: relative;
		margin-bottom: 3rem;
		width: 100%;
	}

	h2 {
		padding: 0.5rem 1rem;
		padding-bottom: 0.5rem;
		font-weight: 400;
		margin: 0;
		font-size: 1.2rem;
	}

	.photo-group-container {
		display: flex;
		flex-direction: column;
		padding: 0; /* Remove padding that might push content out of bounds */
		box-sizing: border-box;
		width: 100%;
	}

	.photo-group {
		display: flex;
		flex-direction: column;
		box-sizing: border-box;
		padding: 1rem 0rem;
		width: 100%;
	}

	.photo-group-label {
		padding: 0.5rem 2rem;
		font-weight: 400;
		font-size: 1.2rem;
		color: var(--imag-10);
		width: fit-content;
		border-bottom: none;
	}
</style>
