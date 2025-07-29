<script lang="ts">
	import { page } from "$app/state";
	import Lightbox from "$lib/components/Lightbox.svelte";
	import LoadingContainer from "$lib/components/LoadingContainer.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import { lightbox } from "$lib/states/index.svelte";
	import type { IImageObjectData } from "$lib/types/images";
	import { DateTime } from "luxon";
	import { SvelteSet } from "svelte/reactivity";
	import type { PageData } from "./$types.js";
	import hotkeys from "hotkeys-js";
	import { blurOnEsc, buildGridArray } from "$lib/utils/dom.js";
	import { dev } from "$app/environment";
	
	let { data = $bindable(page.data as PageData) } = $props();

	// Types
	type ImageGridArray = {
		image: IImageObjectData;
		row: number;
		rowSize: number;
		column: number;
		columnSize: number;
		size: number;
	}[][];

	// Keyboard events
	const permittedKeys: string[] = [];
	const selectKeys = ["Enter", "Space", " "];
	const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
	permittedKeys.push(...selectKeys, ...moveKeys);

	const pagination = $state({
		limit: 25,
		offset: 0
	});

	// HTML Elements
	let imagesGridEl: HTMLDivElement | undefined = $state();

	// Data
	let loadedData = $state(Object.assign({}, data.response));

	// Lightbox
	let lightboxImage: IImageObjectData | undefined = $state();
	let currentImageEl: HTMLImageElement | undefined = $derived(lightboxImage ? document.createElement("img") : undefined);
	let imageLoaded = $derived(false);

	// Pagination
	let imagesData = $derived(loadedData.images.slice(0, pagination.limit * (pagination.offset === 0 ? 1 : pagination.offset + 1))); // initialize with pagination limit
	let shouldUpdate = $derived(loadedData.images.length > pagination.limit * pagination.offset);

	// Selection
	let selectedImages = $state<SvelteSet<IImageObjectData>>(new SvelteSet());
	let singleSelectedImage: IImageObjectData | undefined = $state();
	let imageGridArray: ImageGridArray | undefined = $derived.by(() => {
		// depend on imagesData
		imagesGridEl;
		if (!imagesGridEl) {
			return;
		}

		const gridArray = buildImageGridArray(imagesGridEl);
		return gridArray;
	});

	// Inspecting/Debugging
	if (window.debug) {
		$inspect(
			"selected images",
			selectedImages.values().map((i) => i.id)
		);
	}

	$effect(() => {
		if (lightboxImage) {
			lightbox.show = true;
			loadImage(lightboxImage!.urls.original, currentImageEl!).then(() => {
				imageLoaded = true;
			});
		}
	});

	async function loadImage(url: string, element: HTMLImageElement): Promise<HTMLImageElement> {
		return new Promise((resolve, reject) => {
			element.src = url;
			element.onload = () => resolve(element);
			element.onerror = reject;
		});
	}

	function handleImageCardSelect(image: IImageObjectData, e: MouseEvent) {
		if (e.shiftKey && singleSelectedImage) {
			const ids = imagesData.map((i) => i.id);
			const startIndex = ids.indexOf(singleSelectedImage.id);
			const endIndex = ids.indexOf(image.id);

			const start = Math.min(startIndex, endIndex);
			const end = Math.max(startIndex, endIndex);

			for (let i = start; i <= end; i++) {
				selectedImages.add(imagesData[i]);
			}
		} else {
			selectedImages.clear();
			selectedImages.add(image);
			singleSelectedImage = image;
		}
	}

	function handleKeydownCardSelect(image: IImageObjectData, e: KeyboardEvent) {
		if (!imageGridArray) {
			return;
		}

		const imageInGridArray = imageGridArray
			.find((i) => i.find((j) => j.image.id === image.id))
			?.find((j) => j.image.id === image.id);

		if (!imageInGridArray) {
			if (dev) {
				console.warn(`Can't find image ${image.id} in grid array`);
			}

			return;
		}

		const columnCount = imageGridArray?.[0].length;
		const positionIndexInGrid = imageInGridArray.row * columnCount + imageInGridArray.column;
		const imageGridChildren = imagesGridEl?.children;

		// Mimic click since we already have a handler for that in `handleImageCardSelect()`
		const focusAndSelectElement = (element: HTMLElement | undefined) => {
			// out of bounds
			if (!element) {
				return;
			}

			// maybe unnessary to blur but i wanna make sure lmao
			(Array.from(imageGridChildren!)[positionIndexInGrid] as HTMLElement).blur();
			element.focus();
			element.click();
		};

		switch (e.key) {
			case "ArrowRight":
				const elementRight = imageGridChildren?.item(positionIndexInGrid + 1) as HTMLElement;
				focusAndSelectElement(elementRight);
				break;
			case "ArrowLeft":
				const elementLeft = imageGridChildren?.item(positionIndexInGrid - 1) as HTMLElement;
				focusAndSelectElement(elementLeft);
				break;
			case "ArrowUp":
				const elementUp = imageGridChildren?.item(positionIndexInGrid - columnCount) as HTMLElement;
				focusAndSelectElement(elementUp);

				break;
			case "ArrowDown":
				const elementDown = imageGridChildren?.item(positionIndexInGrid + columnCount) as HTMLElement;
				focusAndSelectElement(elementDown);
				break;
		}
	}

	function buildImageGridArray(element: HTMLElement) {
		const array = buildGridArray(element).map((i) => {
			return i.map((j) => {
				const imageId = j.element?.getAttribute("data-image-id")!;
				const image = imagesData.find((i) => i.id === imageId)!;

				return {
					image,
					row: j.row,
					column: j.column,
					columnSize: j.columnSize,
					size: j.size,
					rowSize: j.rowSize
				};
			});
		});

		return array;
	}

	function unselectImagesOnClickOutsideGrid(element: HTMLElement) {
		$effect(() => {
			document.addEventListener("click", (e) => {
				const target = e.target as HTMLElement;
				const imageCard = target.closest(".image-card") as HTMLElement | undefined;
				const isGridButNotImageCard = target === element && !imageCard;
				if (!element.contains(target) || isGridButNotImageCard) {
					selectedImages.clear();
				}
			});

			return () => {
				document.removeEventListener("click", (e) => {
					const target = e.target as HTMLElement;
					const imageCard = target.closest(".image-card") as HTMLElement | undefined;
					const isGridButNotImageCard = target === element && !imageCard;
					if (!element.contains(target) || isGridButNotImageCard) {
						selectedImages.clear();
					}
				});
			};
		});
	}

	function printGridAsTable() {
		console.log(
			`%cGrid Array at ${DateTime.now().toFormat("dd.MM.yyyy HH:mm:ss")}`,
			"font-weight: bold; color: var(--imag-100); font-size: 18px;"
		);
		console.table(imageGridArray?.map((i) => i.map((j) => j.image?.name)));
	}

	hotkeys("ctrl+a", (e) => {
		e.preventDefault();
		selectedImages.clear();
		imagesData.forEach((i) => selectedImages.add(i));
	});

	hotkeys("esc", (e) => {
		e.preventDefault();
		selectedImages.clear();
		lightboxImage = undefined;
		imageLoaded = false;
	});
</script>

{#if lightboxImage}
	<Lightbox
		onclick={() => {
			lightboxImage = undefined;
			imageLoaded = false;
		}}
	>
		{#if !imageLoaded}
			<div style="width: 3em; height: 3em">
				<LoadingContainer />
			</div>
		{:else}
			<img
				bind:this={currentImageEl}
				class="lightbox-image"
				alt="{lightboxImage.name} by {lightboxImage.uploaded_by.username}"
				title="{lightboxImage.name} by {lightboxImage.uploaded_by.username}"
				loading="eager"
				data-image-id={lightboxImage.id}
			/>
		{/if}
	</Lightbox>
{/if}

<!-- TODO: Create a floating toolbar component that will be shown on selection :)
 Ideally, this would be the default but would allow users to be able to pin it somewhere
 or let it be shown in a fixed place, like the header.
 When user clicks on a card, it'll just select that one card and they can do actions 
 on that card but if there is more than one card, a floating toolbar can be shown as a shortcut instead
 of using the context menu actions for each card
-->

<VizViewContainer
	bind:data={imagesData}
	bind:hasMore={shouldUpdate}
	name="{loadedData.name} - Collection"
	style="padding: 0em {page.url.pathname === '/' ? '1em' : '0em'};"
	paginate={() => {
		pagination.offset++;
	}}
>
	<div id="viz-toolbar-container" style="top: {0}px;">
		<!-- TODO: Selection options will show in this toolbar -->
		<input
			name="name"
			id="coll-name-floating"
			type="text"
			placeholder="Add a title"
			autocomplete="off"
			autocorrect="off"
			title={loadedData.name}
			value={loadedData.name}
			oninput={(e) => (loadedData.name = e.currentTarget.value)}
			onkeydown={blurOnEsc}
		/>
		<span id="coll-details-floating"
			>{DateTime.fromJSDate(loadedData.created_on).toFormat("dd.MM.yyyy")} - {loadedData.image_count}
			{loadedData.image_count === 1 ? "image" : "images"}</span
		>

		<div id="coll-tools">
			{#if dev}
				<button
					id="print-collection-button"
					class="toolbar-button"
					title="Print Collection as Table"
					aria-label="Print Collection as Table"
					onclick={() => {
						printGridAsTable();
					}}
				>
					<MaterialIcon showHoverBG={false} iconName="print" />
				</button>
			{/if}
		</div>
	</div>
	<div id="viz-info-container">
		<div id="coll-metadata">
			<span id="coll-details"
				>{DateTime.fromJSDate(loadedData.created_on).toFormat("dd.MM.yyyy")} - {loadedData.image_count}
				{loadedData.image_count === 1 ? "image" : "images"}</span
			>
		</div>
		<input
			name="name"
			id="coll-name"
			type="text"
			placeholder="Add a title"
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
	<div
		bind:this={imagesGridEl}
		class="images-grid"
		style="padding: 0em {page.url.pathname === '/' ? '1em' : '2em'};"
		data-collection-id={loadedData.id}
		use:unselectImagesOnClickOutsideGrid
	>
		{#each imagesData as image, i}
			{@const imageDate = DateTime.fromJSDate(image.uploaded_on)}
			<div
				class="image-card {selectedImages
					.values()
					.map((i) => i.id)
					.some((i) => i === image.id) || singleSelectedImage?.id === image.id
					? 'selected-card'
					: ''}"
				role="button"
				tabindex="0"
				data-image-id={image.id}
				onclick={(e) => {
					e.preventDefault();
					handleImageCardSelect(image, e);
				}}
				ondblclick={(e) => {
					e.preventDefault();
					lightboxImage = image;
				}}
				onkeydown={(e) => {
					e.preventDefault();
					handleKeydownCardSelect(image, e);
				}}
			>
				<div class="image-container">
					<img
						class="image-card-image"
						src={image.urls.preview}
						alt="{image.name} by {image.uploaded_by.username}"
						title="{image.name} by {image.uploaded_by.username}"
						loading="lazy"
					/>
				</div>
				<div class="image-card-meta">
					<span class="image-card-name" title={image.image_data.file_name}>{image.image_data.file_name}</span>
					<div class="image-card-date_time" title="Photo taken at {imageDate.toFormat('dd/MM/yyyy - HH:mm')}">
						<span class="image-card-date">{imageDate.toFormat("dd/MM/yyyy")}</span>
						<span class="image-card-divider">•</span>
						<span class="image-card-time">{imageDate.toFormat("HH:mm")}</span>
					</div>
				</div>
			</div>
		{/each}
	</div>
</VizViewContainer>

<style lang="scss">
	#viz-toolbar-container {
		position: sticky;
		z-index: 1;
		background-color: rgba(39, 51, 74, 0.9);
		backdrop-filter: blur(5px);
		border-bottom: 1px solid var(--imag-60);
		width: 100%;
		max-width: 100%;
		display: flex;
		align-items: center;
		flex-direction: row;
	}

	#coll-tools {
		display: flex;
		align-items: center;
		position: absolute;
		right: 2rem;
	}

	.toolbar-button {
		border-radius: 10em;
		width: 2em;
		height: 2em;
		padding: 1em;
		font-size: 1.3em;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	#coll-name-floating {
		font-weight: bold;
		font-size: 1.2rem;
		color: var(--imag-text-color);
		background-color: transparent;
		padding: 0.2rem 0rem;
		margin: 0rem 2rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 30%;
	}

	#coll-details-floating {
		color: var(--imag-60);
		background-color: transparent;
		padding: 0.5rem 0rem;
		margin: 0rem -1rem;
		font-family: var(--imag-code-font);
	}

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

	.images-grid {
		margin: 1em 0em;
		margin-bottom: 2em;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(15em, 1fr));
		gap: 1rem;
		width: calc(100% - (1em * 2) - 2rem);
		max-width: 100%;
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

		&:hover {
			background-color: var(--imag-90);
		}
	}

	.selected-card {
		outline: 2px solid var(--imag-primary);
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
