<script module>
	export { searchForData };

	function searchForData(searchValue: string, images: Image[]) {
		if (searchValue.trim() === "") {
			return [];
		}
		// eventually this should also look through keywords/tags
		// and labels idk. fuzzy search???
		return images.filter((i) => i.name.toLowerCase().includes(searchValue.toLowerCase()));
	}
</script>

<script lang="ts">
	import { invalidateAll } from "$app/navigation";
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
	import { onMount, type ComponentProps } from "svelte";
	import { sortCollectionImages } from "$lib/sort/sort.js";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import UploadManager from "$lib/upload/manager.svelte.js";
	import { addCollectionImages, getFullImagePath, type Image } from "$lib/api";
	import { thumbHashToDataURL } from "thumbhash";
	import { fade } from "svelte/transition";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte.js";
	import { handle } from "@oazapfts/runtime";

	let { data } = $props();
	// Keyboard events
	const permittedKeys: string[] = [];
	const selectKeys = ["Enter", "Space", " "];
	const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
	permittedKeys.push(...selectKeys, ...moveKeys);

	// Data
	let loadedData = $derived.by(() => ({
		...data,
		name: localName ?? data.name,
		description: localDescription ?? data.description
	}));
	// Track local edits separately to avoid clobbering them on refresh
	let localName: string | undefined = $state();
	let localDescription: string | undefined = $state();
	let loadedImages = $derived(loadedData.images?.items?.map((img) => img.image) ?? []);

	// Lightbox
	let lightboxImage: Image | undefined = $state();
	let currentImageEl: HTMLImageElement | undefined = $derived(lightboxImage ? document.createElement("img") : undefined);

	$effect(() => {
		if (lightboxImage) {
			lightbox.show = true;
		}
	});

	// Search stuff
	let searchValue = $state("");
	let searchData = $derived(searchForData(searchValue, loadedImages));

	// Pagination
	// NOTE: This might be moved to a settings thing and this could just be default
	const pagination = $state({
		limit: data.images.limit ?? 25,
		offset: data.images.offset ?? 0
	});

	// the searchValue hides the loading indicator when searching since we're
	// already searching through *all* the data that is available on the client
	let shouldUpdate = $derived(loadedImages.length > pagination.limit * pagination.offset && searchValue.trim() === "");

	// Selection
	let selectedAssets = $state<SvelteSet<Image>>(new SvelteSet());
	let singleSelectedAsset: Image | undefined = $state();

	let imageGridArray: AssetGridArray<Image> | undefined = $state();

	// Toolbar stuff
	let toolbarOpacity = $state(0);

	// Thumbhash placeholder
	let placeholderDataURL = $derived.by(() => {
		if (lightboxImage?.image_metadata?.thumbhash) {
			try {
				// Convert base64 thumbhash to Uint8Array
				const binaryString = atob(lightboxImage.image_metadata.thumbhash);
				const bytes = new Uint8Array(binaryString.length);
				for (let i = 0; i < binaryString.length; i++) {
					bytes[i] = binaryString.charCodeAt(i);
				}
				return thumbHashToDataURL(bytes);
			} catch (error) {
				console.warn("Failed to decode thumbhash:", error);
			}
		}
	});

	// Display Data
	let displayData = $derived(
		searchValue.trim()
			? sortCollectionImages(searchData, sort)
			: sortCollectionImages(loadedData.images?.items?.map((img) => img.image) ?? [], sort)
	);

	// Grid props
	let grid: ComponentProps<typeof AssetGrid<Image>> = $derived({
		assetSnippet: imageCard,
		assetGridArray: imageGridArray,
		selectedAssets,
		singleSelectedAsset,
		data: displayData,
		assetDblClick: (_, asset) => {
			lightboxImage = asset;
		}
	});

	async function handleCollectionUpload() {
		// allowed image types will come from the config but for now just hardcode
		const controller = new UploadManager([...SUPPORTED_RAW_FILES, ...SUPPORTED_IMAGE_TYPES] as SupportedImageTypes[]);
		controller.openFileHolder();
		const uploadedImages = await controller.uploadImage();

		const response = await addCollectionImages(loadedData.uid, {
			uids: uploadedImages.map((img) => img.uid)
		});

		if (response.data.added) {
			toastState.addToast({
				message: `Added photos to collection`,
				type: "success",
				timeout: 3000
			});

			await invalidateAll();
		}
	}

	hotkeys("esc", (e) => {
		lightboxImage = undefined;
	});
</script>

{#if lightboxImage}
	{@const imageToLoad = getFullImagePath(lightboxImage.image_paths?.preview) ?? ""}
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
			<img
				src={placeholderDataURL}
				class="lightbox-image"
				style="height: 90%; position: absolute;"
				out:fade={{ duration: 300 }}
				alt=""
				aria-hidden="true"
			/>
		{:then _}
			<img
				src={imageToLoad}
				class="lightbox-image"
				in:fade={{ duration: 300 }}
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

{#snippet imageCard(asset: Image)}
	<ImageCard {asset} />
{/snippet}

{#snippet searchInputSnippet()}
	<SearchInput style="margin: 0em 1em;" bind:value={searchValue} />
	<div id="coll-tools">
		<Button
			id="upload_to_collection"
			class="toolbar-button"
			style="font-size: 0.8rem; background-color: var(--imag-80);"
			title="Upload to Collection"
			aria-label="Upload to Collection"
			onclick={() => {
				handleCollectionUpload();
			}}
		>
			Upload
			<MaterialIcon iconName="upload" />
		</Button>
	</div>
{/snippet}

{#snippet noAssetsSnippet()}
	<div id="add_to_collection-container">
		<span style="margin: 1em; color: var(--imag-20); font-size: 1.2rem;">Add images to this collection</span>
		<Button
			id="add_to_collection-button"
			style="padding: 2em 8em; display: flex; align-items: center; justify-content: center;"
			title="Select Photos"
			aria-label="Select Photos"
			onclick={async () => handleCollectionUpload()}
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
	style="font-size: {page.url.pathname === '/' ? '0.9em' : 'inherit'};"
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
				<span id="coll-details"
					>{DateTime.fromJSDate(new Date(loadedData.created_at)).toFormat("dd.MM.yyyy")}
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
				oninput={(e) => (localName = e.currentTarget.value)}
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
					localDescription = e.currentTarget.value;
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
		min-height: 2rem;
		color: var(--imag-text-color);
		background-color: var(--imag-bg-color);
		outline: none;
		border: none;
		font-family: var(--imag-font-family);
		font-weight: bold;
		padding: 0.5rem 2rem;

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
		font-size: 3em;
		font-weight: bold;
	}

	#coll-description {
		font-size: 1.2em;
		resize: none;
		font-weight: 400;
		height: 2rem;
		padding: 0.3rem inherit;
	}

	#coll-tools {
		display: flex;
		align-items: center;
		height: 100%;
		position: absolute;
		right: 2rem;
	}

	:global(.lightbox-image) {
		max-width: 80%;
		max-height: 90%;
	}
</style>
