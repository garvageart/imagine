<script module>
	import { ImageObjectData } from "$lib/entities/image.js";

	export { searchForData };

	function searchForData(searchValue: string, images: ImageObjectData[]) {
		if (searchValue.trim() === "") {
			return [];
		}
		// eventually this should also look through keywords/tags
		// and labels idk. fuzzy search???
		return images.filter((i) => i.name.toLowerCase().includes(searchValue.toLowerCase()));
	}
</script>

<script lang="ts">
	import { invalidate } from "$app/navigation";
	import { page } from "$app/state";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import Lightbox from "$lib/components/Lightbox.svelte";
	import LoadingContainer from "$lib/components/LoadingContainer.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import SearchInput from "$lib/components/SearchInput.svelte";
	import { lightbox, sort } from "$lib/states/index.svelte";
	import type { AssetGridArray } from "$lib/types/asset.js";
	import { SUPPORTED_IMAGE_TYPES, SUPPORTED_RAW_FILES, type SupportedImageTypes } from "$lib/types/images";
	import { blurOnEsc, loadImage } from "$lib/utils/dom.js";
	import hotkeys from "hotkeys-js";
	import { DateTime } from "luxon";
	import { SvelteSet } from "svelte/reactivity";
	import type { ComponentProps } from "svelte";
	import { sortCollectionImages } from "$lib/sort/sort.js";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import UploadManager from "$lib/upload/manager.svelte.js";
	import { sendAPIRequest } from "$lib/utils/http";

	let { data } = $props();
	// Keyboard events
	const permittedKeys: string[] = [];
	const selectKeys = ["Enter", "Space", " "];
	const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
	permittedKeys.push(...selectKeys, ...moveKeys);

	// Data
	let loadedData = $state(data.response);

	// Lightbox
	let lightboxImage: ImageObjectData | undefined = $state();
	let currentImageEl: HTMLImageElement | undefined = $derived(lightboxImage ? document.createElement("img") : undefined);

	$effect(() => {
		if (lightboxImage) {
			lightbox.show = true;
		}
	});

	// Search stuff
	let searchValue = $state("");
	let searchData = $derived(searchForData(searchValue, loadedData.images));

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
	let selectedAssets = $state<SvelteSet<ImageObjectData>>(new SvelteSet());
	let singleSelectedAsset: ImageObjectData | undefined = $state();

	let imageGridArray: AssetGridArray<ImageObjectData> | undefined = $state();

	// Toolbar stuff
	let toolbarOpacity = $state(0);

	// Display Data
	let displayData = $derived(
		searchValue.trim() ? sortCollectionImages(searchData, sort) : sortCollectionImages(loadedData.images, sort)
	);

	// Grid props
	let grid: ComponentProps<typeof AssetGrid<ImageObjectData>> = $derived({
		assetSnippet: imageCard,
		assetGridArray: imageGridArray,
		selectedAssets,
		singleSelectedAsset,
		data: displayData,
		assetDblClick: (_, asset) => {
			lightboxImage = asset;
		}
	});

	hotkeys("esc", (e) => {
		lightboxImage = undefined;
	});
</script>

{#if lightboxImage}
	{@const imageToLoad = lightboxImage.image_paths?.original_path ?? ''}
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
				alt={lightboxImage.name}
				title={lightboxImage.name}
				loading="eager"
				data-image-uid={lightboxImage.uid}
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
	{#if displayData.length > 0}
		<SearchInput style="margin: 0em 1em;" bind:value={searchValue} />
	{/if}
{/snippet}

{#snippet noAssetsSnippet()}
	<div id="add_to_collection-container">
		<span style="margin: 1em; color: var(--imag-20); font-size: 1.2rem;">Add images to this collection</span>
		<Button
			id="add_to_collection-button"
			style="padding: 2em 8em; display: flex; align-items: center; justify-content: center;"
			title="Select Photos"
			aria-label="Select Photos"
			onclick={async () => {
				// allowed image types will come from the config but for now just hardcode
				const controller = new UploadManager([...SUPPORTED_RAW_FILES, ...SUPPORTED_IMAGE_TYPES] as SupportedImageTypes[]);
				controller.openFileHolder();
				const uploadedImages = await controller.uploadImage();

				const response = await sendAPIRequest<{ added: true }>(`collection/${loadedData.uid}/images`, {
					method: "PUT",
					body: JSON.stringify({
						uids: uploadedImages.map((img) => img.uid)
					})
				});

				if (response.added) {
					await invalidate(page.url.pathname);
				}
			}}
		>
			Select Photos
			<MaterialIcon iconName="add" style="font-size: 2em;" />
		</Button>
	</div>
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
		bind:grid
		{pagination}
		{noAssetsSnippet}
		toolbarSnippet={searchInputSnippet}
		toolbarProps={{
			style: "justify-content: center; "
		}}
	>
		<div id="viz-info-container">
			<div id="coll-metadata">
				<span id="coll-details">
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

	#add_to_collection-container {
		display: flex;
		flex-direction: column;
		justify-content: left;
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

	:global(.lightbox-image) {
		max-width: 70%;
		max-height: 70%;
	}
</style>
