<script lang="ts">
	import { goto } from "$app/navigation";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import AssetToolbar from "$lib/components/AssetToolbar.svelte";
	import Lightbox from "$lib/components/Lightbox.svelte";
	import LoadingContainer from "$lib/components/LoadingContainer.svelte";
	import LoadingSpinner from "$lib/components/LoadingSpinner.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { performSearch } from "$lib/search/execute";
	import { lightbox, search, sort } from "$lib/states/index.svelte";
	import { loadImage } from "$lib/utils/dom";
	import { onMount, type ComponentProps } from "svelte";
	import { SvelteSet } from "svelte/reactivity";
	import { collectionCard } from "../collections/+page.svelte";
	import { ImageObjectData } from "$lib/entities/image";
	import type { Collection, IImageObjectData } from "$lib/types/images";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import SearchInput from "$lib/components/SearchInput.svelte";
	import { sortCollectionImages } from "$lib/sort/sort";
	import type { AssetGridArray } from "$lib/types/asset";
	import { searchForData } from "../collections/[id]/+page.svelte";

	export class SelectedAssets<T> {
		selectedAssets = $state<SvelteSet<T>>(new SvelteSet());
		singleSelectedAsset: T | undefined = $state();
	}

	let collections = $derived(search.data.collections.data);
	let images = $derived(search.data.images.data);
	let totalResults = $derived(collections.length + images.length);

	let timeFound = $state(0);

	// Lightbox
	let lightboxImage: ImageObjectData | undefined = $state();
	let currentImageEl: HTMLImageElement | undefined = $derived(lightboxImage ? document.createElement("img") : undefined);

	$effect(() => {
		if (lightboxImage) {
			lightbox.show = true;
		}
	});

	// Selection
	let collectionSelectedAssets = $state(new SelectedAssets<Collection>());
	let imageSelectedAssets = $state(new SelectedAssets<ImageObjectData>());

	// Images Stuff

	// - Search stuff
	let collectionSearchValue = $state("");
	let searchData = $derived(searchForData(collectionSearchValue, images));

	// - Display Data
	let imageGridArray: AssetGridArray<IImageObjectData> | undefined = $state();
	let imageDisplayData = $derived(
		collectionSearchValue.trim() ? sortCollectionImages(searchData, sort) : sortCollectionImages(images, sort)
	);

	let imagesGrid: ComponentProps<AssetGrid<ImageObjectData>> = $derived({
		selectedAssets: imageSelectedAssets.selectedAssets,
		singleSelectedAsset: imageSelectedAssets.singleSelectedAsset,
		assetGridArray: imageGridArray,
		data: imageDisplayData,
		assetDblClick(_, asset) {
			lightboxImage = asset;
		},
		assetSnippet: imageCard
	});

	let collectionsGrid: ComponentProps<AssetGrid<Collection>> = $derived({
		selectedAssets: collectionSelectedAssets.selectedAssets,
		singleSelectedAsset: collectionSelectedAssets.singleSelectedAsset,
		data: collections,
		assetDblClick(_, asset) {
			goto(`/collections/${asset.id}`);
		},
		assetSnippet: collectionCard
	});

	let selectedAssets = $derived.by(() => {
		return collectionSelectedAssets.selectedAssets.union(imageSelectedAssets.selectedAssets);
	});

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
	{@const imageToLoad = lightboxImage.urls.original}
	<Lightbox
		onclick={() => {
			lightboxImage = undefined;
		}}
	>
		{#await loadImage(imageToLoad, currentImageEl!)}
			<div style="width: 3em; height: 3em">
				<LoadingContainer />
			</div>
		{:then _}
			<img
				src={imageToLoad}
				class="lightbox-image"
				alt="{lightboxImage.name} by {lightboxImage.uploaded_by.username}"
				title="{lightboxImage.name} by {lightboxImage.uploaded_by.username}"
				loading="eager"
				data-image-id={lightboxImage.id}
			/>
		{:catch error}
			<p>Failed to load image</p>
			<p>{error}</p>
		{/await}
	</Lightbox>
{/if}

{#snippet imageCard(asset: ImageObjectData)}
	<ImageCard {asset} />
{/snippet}

{#snippet searchInputSnippet()}
	<SearchInput style="margin: 0em 1em;" bind:value={collectionSearchValue} />
{/snippet}

<!-- Container to show search information and context to help the user
 better understand the results and further help them find what they're looking for 
-->
<div id="search-info-container" class="selection-container">
	{#if selectedAssets.size > 1}
		<AssetToolbar class="selection-toolbar">
			<button
				id="coll-clear-selection"
				title="Clear selection"
				aria-label="Clear selection"
				style="margin-right: 0.5em;"
				class="toolbar-button"
				onclick={() => selectedAssets.clear()}
			>
				<MaterialIcon iconName="close" />
			</button>
			<span style="font-weight: 600;">{selectedAssets.size} selected</span>
		</AssetToolbar>
	{:else if !search.loading}
		<div id="search-info">
			<p>{totalResults} results found in {(timeFound / 1000).toFixed(2)} seconds</p>
			<div class="search-info-details">
				<p>
					{collections.length} collection{collections.length === 1 ? "" : "s"}, {images.length} image{images.length === 1
						? ""
						: "s"}
				</p>
			</div>
		</div>
	{/if}
</div>

<div class="search-container no-select">
	<h1 style="margin-top: 2rem;">Search</h1>
	{#if search.loading}
		<div class="loading-container">
			<p id="search-loading-text">Searching for "{search.value}"...</p>
			<LoadingSpinner />
		</div>
	{:else if search.executed}
		<div class="results">
			{#if totalResults === 0}
				<!-- TODO: Create a suggestions component to show other collections
			 something like a closest match to what was searched			 
			-->
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
						<h2>Images ({images.length})</h2>
						<AssetsShell
							toolbarSnippet={searchInputSnippet}
							toolbarProps={{
								style: "justify-content: center;"
							}}
							bind:grid={imagesGrid}
						/>
					</section>
				{/if}
			{/if}
		</div>
	{/if}
</div>

<style>
	#search-info-container {
		z-index: 1;
	}

	#search-info {
		padding: 1rem 1rem;
		border-bottom: 1px solid var(--imag-60);
		display: flex;
		align-items: center;
		font-size: 0.9rem;
		color: var(--imag-40);
	}

	#search-info p {
		margin: 0;
		padding: 0rem 0.5rem;
	}

	.search-container {
		white-space: wrap;
		display: flex;
		align-items: center;
		flex-direction: column;
		overflow-y: auto;
	}

	h1 {
		margin-bottom: 1.5rem;
		font-weight: 800;
		text-align: center;
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
		margin-bottom: 1rem;
	}

	.results {
		width: 100%;
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
		border-bottom: 1px solid var(--imag-20);
		font-weight: 400;
	}

	/* .collection-card,
	.image-card {
		border-radius: 0.5rem;
		overflow: hidden;
	}

	.collections-grid,
	.images-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.5rem;
		width: 100%;
	} */

	/* @media (max-width: 768px) {
		.collections-grid,
		.images-grid {
			grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		}
	}

	@media (max-width: 480px) {
		.collections-grid,
		.images-grid {
			grid-template-columns: 1fr;
		}
	} */
</style>
