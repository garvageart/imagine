<script lang="ts">
	import { page } from "$app/state";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import Lightbox from "$lib/components/Lightbox.svelte";
	import LoadingContainer from "$lib/components/LoadingContainer.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import SearchInput from "$lib/components/SearchInput.svelte";
	import { lightbox, sort } from "$lib/states/index.svelte";
	import type { AssetGridArray } from "$lib/types/asset.js";
	import type { IImageObjectData } from "$lib/types/images";
	import { blurOnEsc, loadImage } from "$lib/utils/dom.js";
	import hotkeys from "hotkeys-js";
	import { DateTime } from "luxon";
	import { SvelteSet } from "svelte/reactivity";
	import type { ComponentProps } from "svelte";
	import { sortCollectionImages } from "$lib/sort/sort.js";

	let { data } = $props();
	// Keyboard events
	const permittedKeys: string[] = [];
	const selectKeys = ["Enter", "Space", " "];
	const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
	permittedKeys.push(...selectKeys, ...moveKeys);

	// Data
	let loadedData = $state(data.response);

	// Lightbox
	let lightboxImage: IImageObjectData | undefined = $state();
	let currentImageEl: HTMLImageElement | undefined = $derived(lightboxImage ? document.createElement("img") : undefined);

	// Search stuff
	let searchValue = $state("");
	let searchData = $derived(searchForData());

	// Pagination
	// NOTE: This might be moved to a settings thing and this could just be default
	const pagination = $state({
		limit: 25,
		offset: 0
	});

	// the searchValue hides the loading indicator when searching since we're
	// already searching through *all* the data that is available on the client
	let shouldUpdate = $derived(loadedData.images.length > pagination.limit * pagination.offset && searchValue.trim() === "");

	// Selection
	let selectedAssets = $state<SvelteSet<IImageObjectData>>(new SvelteSet());
	let singleSelectedAsset: IImageObjectData | undefined = $state();

	let imageGridArray: AssetGridArray<IImageObjectData> | undefined = $state();

	// Toolbar stuff
	let toolbarOpacity = $state(0);

	function searchForData() {
		if (searchValue.trim() === "") {
			return [];
		}
		// eventually this should also look through keywords/tags
		// and labels idk. fuzzy search???
		return loadedData.images.filter((i) => i.name.toLowerCase().includes(searchValue.toLowerCase()));
	}

	// Display Data
	let displayData = $derived(
		searchValue.trim() ? sortCollectionImages(searchData, sort) : sortCollectionImages(loadedData.images, sort)
	);

	// Grid props
	let grid: ComponentProps<typeof AssetGrid<IImageObjectData>> = $derived({
		assetSnippet: imageCard,
		assetGridArray: imageGridArray,
		selectedAssets,
		singleSelectedAsset,
		data: displayData,
		assetDblClick: (_, asset) => {
			lightboxImage = asset;
		}
	});

	$effect(() => {
		if (lightboxImage) {
			lightbox.show = true;
		}
	});

	hotkeys("esc", (e) => {
		lightboxImage = undefined;
	});
</script>

{#if lightboxImage}
	{@const imageToLoad = lightboxImage.urls.original}
	<Lightbox
		onclick={() => {
			lightboxImage = undefined;
		}}
	>
		<!-- Awaitng like this is better inline but `currentImageEl` is kinda
	 being created/allocated unncessarily and is never removed or freed until the component is destroyed
	 It's small but annoying enough where I want to find a different way to load an image
	  -->
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

{#snippet imageCard(asset: IImageObjectData)}
	{@const imageDate = DateTime.fromJSDate(asset.uploaded_on)}

	<div class="image-card" data-asset-id={asset.id}>
		<div class="image-container">
			<img
				class="image-card-image"
				src={asset.urls.preview}
				alt="{asset.name} by {asset.uploaded_by.username}"
				title="{asset.name} by {asset.uploaded_by.username}"
				loading="lazy"
			/>
		</div>
		<div class="image-card-meta">
			<span class="image-card-name" title={asset.image_data.file_name}>{asset.image_data.file_name}</span>
			<div class="image-card-date_time" title="Photo taken at {imageDate.toFormat('dd/MM/yyyy - HH:mm')}">
				<span class="image-card-date">{imageDate.toFormat("dd/MM/yyyy")}</span>
				<span class="image-card-divider">•</span>
				<span class="image-card-time">{imageDate.toFormat("HH:mm")}</span>
			</div>
		</div>
	</div>
{/snippet}

{#snippet toolbarSnippet()}
	<SearchInput style="margin: 0em 1em;" bind:value={searchValue} />
{/snippet}

<VizViewContainer
	bind:data={displayData}
	bind:hasMore={shouldUpdate}
	name="{loadedData.name} - Collection"
	style="padding: 0em {page.url.pathname === '/' ? '1em' : '0em'};"
	paginate={() => {
		pagination.offset++;
	}}
	onscroll={(e) => {
		const info = document.getElementById("viz-info-container")!;
		const bottom = info.scrollHeight;

		if (e.currentTarget.scrollTop < bottom) {
			toolbarOpacity = e.currentTarget.scrollTop / bottom;
		} else {
			toolbarOpacity = 1;
		}
	}}
>
	<AssetsShell
		toolbarProps={{
			style: "justify-content: center;"
		}}
		bind:grid
		{toolbarSnippet}
		{pagination}
	>
		<div id="viz-info-container">
			<div id="coll-metadata">
				<span id="coll-details"
					>{DateTime.fromJSDate(loadedData.created_on).toFormat("dd.MM.yyyy")}
					•
					{#if searchValue.trim()}
						{searchData.length} {searchData.length === 1 ? "image" : "images"} of {loadedData.image_count}
					{:else}
						{loadedData.image_count} {loadedData.image_count === 1 ? "image" : "images"}
					{/if}
				</span>
			</div>
			<input
				name="name"
				id="coll-name"
				type="text"
				placeholder="Add a title"
				autocomplete="off"
				autocorrect="off"
				spellcheck="false"
				value={loadedData.name}
				oninput={(e) => (loadedData.name = e.currentTarget.value)}
				onkeydown={blurOnEsc}
			/>
			<textarea
				name="description"
				id="coll-description"
				placeholder="Add a description"
				spellcheck="false"
				rows="1"
				value={loadedData.description}
				oninput={(e) => {
					loadedData.description = e.currentTarget.value;
				}}
				onkeydown={blurOnEsc}
			></textarea>
		</div>
	</AssetsShell>
</VizViewContainer>

<style lang="scss">
	:global(#create-collection) {
		margin: 0em 1rem;
	}

	#viz-info-container {
		width: 100%;
		max-width: 100%;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		margin: 1em 0em;
	}

	#coll-metadata {
		padding: 0.5rem 2rem;
		display: flex;
		color: var(--imag-60);
		font-family: var(--imag-code-font);
	}

	input:not([type="submit"]),
	textarea {
		max-width: 100%;
		min-height: 3rem;
		color: var(--imag-text-color);
		background-color: var(--imag-bg-color);
		outline: none;
		border: none;
		font-family: var(--imag-font-family);
		font-weight: bold;
		padding: 0.75rem 2rem;

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
		}

		&:focus::placeholder {
			color: var(--imag-60);
		}

		&:focus {
			box-shadow: 0 -2px 0 var(--imag-primary) inset;
		}

		&:-webkit-autofill,
		&:-webkit-autofill:focus {
			-webkit-text-fill-color: var(--imag-text-color);
			-webkit-box-shadow: 0 0 0px 1000px var(--imag-100) inset;
			-webkit-box-shadow: 0 -5px 0 var(--imag-primary) inset;
			transition:
				background-color 0s 600000s,
				color 0s 600000s !important;
		}
	}

	#coll-name {
		font-size: 3rem;
		font-weight: bold;
	}

	#coll-description {
		font-size: 1.2rem;
		resize: none;
		font-weight: 400;
		height: 2rem;
		padding: 0.3rem inherit;
	}

	.image-card {
		max-height: 20em;
		background-color: var(--imag-100);
		padding: 0.8em;
		border-radius: 0.5em;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;

		&:focus {
			outline: none;
		}

		&:hover {
			background-color: var(--imag-90);
		}
	}

	.image-container {
		height: 13em;
		background-color: var(--imag-80);
	}

	.image-card img {
		max-width: 100%;
		min-height: 100%;
		height: auto;
		object-fit: contain;
		display: block;
		pointer-events: none; // prevent clicks on image (right clicking should show the to be made context menu)
	}

	.image-card-meta {
		margin-top: 0.5rem;
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-direction: column;
		font-size: 0.9rem;
	}

	.image-card-name {
		font-weight: bold;
		margin-bottom: 0.2em;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 100%;
	}

	.image-card-date_time {
		color: var(--imag-20);
	}

	.image-card-divider {
		color: var(--imag-40);
	}

	.image-card-time {
		font-size: 0.9rem;
	}

	.lightbox-image {
		max-width: 70%;
		max-height: 70%;
	}
</style>
