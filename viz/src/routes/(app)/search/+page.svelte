<script lang="ts">
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import AssetToolbar from "$lib/components/AssetToolbar.svelte";
	import LoadingSpinner from "$lib/components/LoadingSpinner.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { performSearch } from "$lib/search/execute";
	import {
		isLayoutPage,
		modal,
		search,
		viewSettings
	} from "$lib/states/index.svelte";
	import { onMount, type ComponentProps } from "svelte";
	import CollectionCard from "$lib/components/CollectionCard.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import PhotoAssetGrid from "$lib/components/PhotoAssetGrid.svelte";
	import ImageLightbox from "$lib/components/ImageLightbox.svelte";
	import type { Collection, ImageAsset } from "$lib/api";
	import hotkeys from "hotkeys-js";
	import {
		getConsolidatedGroups,
		groupImagesByDate,
		type ConsolidatedGroup,
		type DateGroup
	} from "$lib/photo-layout";
	import { downloadOriginalImageFile } from "$lib/utils/http";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import Dropdown from "$lib/components/Dropdown.svelte";
	import type { MenuItem } from "$lib/context-menu/types";
	import {
		selectionManager,
		SelectionScopeNames
	} from "$lib/states/selection.svelte";
	import type { AssetGridView } from "$lib/types/asset";
	import { createImageMenu } from "$lib/context-menu/menus/images";
	import { createCollectionMenu } from "$lib/context-menu/menus/collections";
	import IconButton from "$lib/components/IconButton.svelte";
	import LabelSelector from "$lib/components/LabelSelector.svelte";
	import StarRating from "$lib/components/StarRating.svelte";
	import { getImageLabel } from "$lib/utils/images";
	import { LabelColours, type ImageLabel } from "$lib/images/constants";
	import { updateImage } from "$lib/api";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import { DragData } from "$lib/drag-drop/data";
	import { VizMimeTypes } from "$lib/constants";
	import CollectionSelectionModal from "$lib/components/modals/CollectionSelectionModal.svelte";
	import { goto } from "$app/navigation";
	import { addCollectionImages } from "$lib/api";

	let collections = $derived(search.data.collections.data);
	let images = $derived(search.data.images.data);
	let totalResults = $derived(collections.length + images.length);

	let timeFound = $state(0);

	// Lightbox
	let lightboxImage: ImageAsset | undefined = $state();

	// Selection Scopes
	const imageScopeId = SelectionScopeNames.SEARCH_IMAGES;
	const collectionScopeId = SelectionScopeNames.SEARCH_COLLECTIONS;

	const imageSelection = selectionManager.getScope<ImageAsset>(imageScopeId);
	const collectionSelection =
		selectionManager.getScope<Collection>(collectionScopeId);
	let disableOutsideUnselect = $derived(isLayoutPage());

	// Modal state for collection selection
	let showCollectionSelectionModal = $state(false);
	let imageUidsForCollection = $state<string[]>([]);

	// Derived selection helpers
	let firstSelectedImage = $derived(Array.from(imageSelection.selected)[0]);
	let firstSelectedCollection = $derived(
		Array.from(collectionSelection.selected)[0]
	);

	// Context Menu State
	let ctxShowMenu = $state(false);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(null);
	let ctxItems: MenuItem[] = $state([]);

	// Action Menus
	let imageActionMenuItems = $derived.by(() => {
		const baseMenuItems = createImageMenu(images, imageSelection, {
			onDelete: (deletedUIDs) => {
				// For search, we might just want to remove them from the view or re-search
				// Ideally, we'd remove them from the local 'images' array if it were mutable in this context
				// Since 'images' is derived from search.data, we might need to trigger a re-search or just let it be.
				// For now, we'll just toast.
				toastState.addToast({
					type: "success",
					message: `${deletedUIDs.length} image(s) deleted.`
				});
				performSearch();
			}
		});

		const pageMenuItems: MenuItem[] = [
			{
				id: "act-add-to-collection",
				label: "Add to Collection",
				icon: "collections_bookmark",
				action: () => {
					imageUidsForCollection = Array.from(imageSelection.selected).map(
						(img) => img.uid
					);
					showCollectionSelectionModal = true;
					modal.show = true;
				}
			}
		];

		return [...pageMenuItems, ...baseMenuItems];
	});

	let collectionActionMenuItems = $derived.by(() => {
		return createCollectionMenu(firstSelectedCollection, {
			onCollectionDeleted: () => performSearch(),
			onCollectionDuplicated: () => performSearch(),
			onCollectionUpdated: () => performSearch()
		});
	});

	// View Modes
	let imageViewMode = $state<AssetGridView>(viewSettings.current);
	// Collections in AssetGrid only support "list" and "thumbnails"
	let collectionViewMode = $state<"list" | "thumbnails">(
		viewSettings.current === "grid"
			? "thumbnails"
			: (viewSettings.current as "list" | "thumbnails")
	);

	// Display options for Images
	const imageDisplayMenuItems: MenuItem[] = [
		{
			id: "img-display-grid",
			label: "Grid",
			action: () => (imageViewMode = "grid")
		},
		{
			id: "img-display-list",
			label: "List",
			action: () => (imageViewMode = "list")
		},
		{
			id: "img-display-thumbnails",
			label: "Thumbnails",
			action: () => (imageViewMode = "thumbnails")
		}
	];

	let imageDisplaySelectedId = $derived.by(() => {
		switch (imageViewMode) {
			case "grid":
				return "img-display-grid";
			case "list":
				return "img-display-list";
			case "thumbnails":
				return "img-display-thumbnails";
		}
	});

	// Display options for Collections
	const collectionDisplayMenuItems: MenuItem[] = [
		{
			id: "col-display-list",
			label: "List",
			action: () => (collectionViewMode = "list")
		},
		{
			id: "col-display-thumbnails",
			label: "Thumbnails",
			action: () => (collectionViewMode = "thumbnails")
		}
	];

	let collectionDisplaySelectedId = $derived.by(() => {
		switch (collectionViewMode) {
			case "list":
				return "col-display-list";
			case "thumbnails":
				return "col-display-thumbnails";
		}
	});

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
		searchValue: search.value,
		view: collectionViewMode,
		scopeId: collectionScopeId,
		onassetcontext: (detail) => {
			const { asset } = detail;
			if (
				!collectionSelection.has(asset) ||
				collectionSelection.selected.size <= 1
			) {
				collectionSelection.select(asset);
			}
			ctxAnchor = detail.anchor;
			ctxItems = collectionActionMenuItems;
			ctxShowMenu = true;
		}
	});

	hotkeys("escape", (e) => {
		e.preventDefault();
		imageSelection.clear();
		collectionSelection.clear();

		lightboxImage = undefined;

		if (modal.show) {
			modal.show = false;
		}
	});

	function openLightbox(asset: ImageAsset) {
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
		if (uids.length === 0) return;

		// If single image, use direct download helper
		if (uids.length === 1) {
			const img = images.find((i) => i.uid === uids[0]);
			if (img) {
				await downloadOriginalImageFile(img);
			}
			return;
		}

		// Bulk download logic (placeholder if API doesn't support generic bulk download yet,
		// but typically we'd use a token or loop. Reusing logic from photos page is best practice.)
		// For now, we'll loop single downloads to ensure functionality
		// A better approach would be to use the proper bulk endpoint if available.
		for (const uid of uids) {
			const img = images.find((i) => i.uid === uid);
			if (img) {
				await downloadOriginalImageFile(img);
			}
		}
	}

	async function handleCollectionSelect(
		collection: Collection,
		newImageUids: string[]
	) {
		if (newImageUids.length === 0) {
			toastState.addToast({
				type: "info",
				message: "No new images to add.",
				timeout: 3000
			});
			return;
		}

		try {
			const res = await addCollectionImages(collection.uid, {
				uids: newImageUids
			});

			if (res.status === 200) {
				toastState.addToast({
					type: "success",
					message: `Added ${newImageUids.length} image(s) to collection **${collection.name}**`,
					timeout: 3000,
					actions: [
						{
							label: "Open Collection",
							onClick: () => {
								goto(`/collections/${collection.uid}`);
							}
						}
					]
				});
			} else {
				toastState.addToast({
					type: "error",
					message: res.data?.error ?? "Failed to add images to collection",
					timeout: 3000
				});
			}
		} catch (error) {
			toastState.addToast({
				type: "error",
				message: `Failed to add images to collection: ${(error as Error).message}`,
				timeout: 3000
			});
		} finally {
			imageSelection.clear();
			showCollectionSelectionModal = false;
			modal.show = false;
			imageUidsForCollection = [];
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

{#if showCollectionSelectionModal}
	<CollectionSelectionModal
		bind:showModal={showCollectionSelectionModal}
		onSelect={handleCollectionSelect}
		imageUidsToAdd={imageUidsForCollection}
	/>
{/if}

{#if lightboxImage}
	<ImageLightbox
		bind:lightboxImage
		{nextLightboxImage}
		{prevLightboxImage}
		onImageUpdated={(image) => imageSelection.updateItem(image, images)}
	/>
{/if}

{#snippet imageCard(asset: ImageAsset)}
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
		{#if imageSelection.selected.size > 0}
			<AssetToolbar class="asset-toolbar">
				<div class="toolbar-content">
					<div class="selection-info">
						<IconButton
							iconName="close"
							class="toolbar-button"
							title="Clear selection"
							aria-label="Clear selection"
							style="margin-right: 1em;"
							onclick={() => imageSelection.clear()}
						/>
						<span style="font-weight: 600;"
							>{imageSelection.selected.size} selected</span
						>
					</div>
					<div class="selection-actions">
						<IconButton
							iconName="collections_bookmark"
							class="action"
							role="tooltip"
							title="Add to Collection"
							onclick={() => {
								imageUidsForCollection = Array.from(
									imageSelection.selected
								).map((img) => img.uid);
								showCollectionSelectionModal = true;
								modal.show = true;
							}}
							ondragenter={(e) => {
								e.currentTarget.classList.add("on-enter");
							}}
							ondragleave={(e) => {
								e.currentTarget.classList.remove("on-enter");
							}}
							ondragover={(e) => {
								e.preventDefault();
							}}
							ondrop={(e) => {
								if (
									!e.dataTransfer?.types.includes(VizMimeTypes.IMAGE_UIDS)
								) {
									return;
								}

								const uidsData = DragData.getData<string[]>(
									e.dataTransfer!,
									VizMimeTypes.IMAGE_UIDS
								)?.payload;

								if (!uidsData) {
									return;
								}

								e.currentTarget.classList.remove("on-enter");

								imageUidsForCollection = uidsData;
								showCollectionSelectionModal = true;
								modal.show = true;
							}}
						>
							Add to Collection
						</IconButton>
						{#if firstSelectedImage}
							<LabelSelector
								variant="expanded"
								label={getImageLabel(firstSelectedImage)}
								onSelect={async (selectedLabel) => {
									if (!firstSelectedImage) {
										return;
									}

									// Reverse lookup: find the key (Name) for the selected color (Value)
									const entry = Object.entries(LabelColours).find(
										([_, colour]) => colour === selectedLabel
									);
									const labelName = entry ? entry[0] : null;
									const labelToSend = (
										labelName === "None" || !labelName
											? null
											: labelName
									) as ImageLabel | null;

									const updatePromises = Array.from(
										imageSelection.selected
									).map((img) =>
										updateImage(img.uid, {
											image_metadata: { label: labelToSend }
										})
									);

									await Promise.all(updatePromises);
								}}
							/>
							<StarRating
								value={firstSelectedImage.image_metadata?.rating ?? 0}
								onChange={async (rating) => {
									if (!firstSelectedImage) {
										return;
									}

									const updatePromises = Array.from(
										imageSelection.selected
									).map((img) =>
										updateImage(img.uid, {
											image_metadata: { rating }
										})
									);

									await Promise.all(updatePromises);
								}}
							/>
						{/if}
					</div>
					<div class="toolbar-right">
						<Dropdown
							class="toolbar-button"
							icon="more_horiz"
							items={imageActionMenuItems}
							showSelectionIndicator={false}
							align="right"
						/>
					</div>
				</div>
			</AssetToolbar>
		{:else if collectionSelection.selected.size > 0}
			<AssetToolbar class="asset-toolbar">
				<div class="toolbar-content">
					<div class="selection-info">
						<IconButton
							iconName="close"
							class="toolbar-button"
							title="Clear selection"
							aria-label="Clear selection"
							style="margin-right: 1em;"
							onclick={() => collectionSelection.clear()}
						/>
						<span style="font-weight: 600;"
							>{collectionSelection.selected.size} selected</span
						>
					</div>
					<div class="toolbar-right">
						<Dropdown
							class="toolbar-button"
							icon="more_horiz"
							items={collectionActionMenuItems}
							showSelectionIndicator={false}
							align="right"
						/>
					</div>
				</div>
			</AssetToolbar>
		{:else if !search.loading}
			<AssetToolbar
				style="position: sticky; top: 0px;"
				class="asset-toolbar"
			>
				<div class="toolbar-content">
					<div class="selection-info">
						<span>
							{totalResults} results found in
							<strong>{(timeFound / 1000).toFixed(2)} seconds</strong>
						</span>
						<div class="search-info-details" style="margin-left: 1rem;">
							<span>
								{collections.length} collection{collections.length === 1
									? ""
									: "s"}, {images.length}
								image{images.length === 1 ? "" : "s"}
							</span>
						</div>
					</div>
					<div class="toolbar-right">
						<IconButton
							iconName="filter_list"
							class="toolbar-button"
							title="Filter"
							aria-label="Filter"
							onclick={() => {
								// Placeholder for now as filter modal isn't fully integrated into search logic
								// But we keep the button for consistency and future wiring
								toastState.addToast({
									type: "info",
									message: "Filtering search results is coming soon",
									timeout: 3000
								});
							}}
						>
							Filter
						</IconButton>
					</div>
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
							<div
								style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--imag-20); padding-right: 1rem;"
							>
								<h2>Collections ({collections.length})</h2>
								{#if collectionSelection.selected.size <= 1}
									<div style="display: flex; align-items: center; gap: 0.5rem;">
										<Dropdown
											title="Display"
											class="toolbar-button"
											icon="list_alt"
											items={collectionDisplayMenuItems}
											selectedItemId={collectionDisplaySelectedId}
										/>
									</div>
								{/if}
							</div>
							<div class="collection-group-container">
								<AssetGrid
									{...collectionsGrid}
									onassetcontext={(detail: {
										asset: Collection;
										anchor: { x: number; y: number } | HTMLElement;
									}) => {
										const { asset } = detail;
										if (
											!collectionSelection.has(asset) ||
											collectionSelection.selected.size <= 1
										) {
											collectionSelection.select(asset);
										}
										ctxAnchor = detail.anchor;
										ctxItems = collectionActionMenuItems;
										ctxShowMenu = true;
									}}
								/>
							</div>
						</section>
					{/if}
					{#if images.length > 0}
						<section class="images-section">
							<div
								style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--imag-20); padding-right: 1rem;"
							>
								<h2>Images ({images.length})</h2>
								{#if imageSelection.selected.size <= 1}
									<div style="display: flex; align-items: center; gap: 0.5rem;">
										<Dropdown
											title="Display"
											class="toolbar-button"
											icon="list_alt"
											items={imageDisplayMenuItems}
											selectedItemId={imageDisplaySelectedId}
										/>
									</div>
								{/if}
							</div>

							<div class="photo-group-container">
								<PhotoAssetGrid
									bind:allData={allImagesFlat}
									view={imageViewMode}
									data={images}
									groupedData={consolidatedGroups}
									showDateHeaders={true}
									scopeId={imageScopeId}
									assetDblClick={(_e, asset) => {
										openLightbox(asset);
									}}
									onassetcontext={(detail: {
										asset: ImageAsset;
										anchor: { x: number; y: number } | HTMLElement;
									}) => {
										const { asset } = detail;
										if (
											!imageSelection.has(asset) ||
											imageSelection.selected.size <= 1
										) {
											imageSelection.select(asset);
										}

										ctxAnchor = detail.anchor;
										ctxItems = imageActionMenuItems;
										ctxShowMenu = true;
									}}
								/>
							</div>
						</section>
					{/if}
				{/if}
			</div>
		{/if}
	</div>
</div>

<!-- Context menu for right-click on assets -->
<ContextMenu
	bind:showMenu={ctxShowMenu}
	items={ctxItems}
	anchor={ctxAnchor}
	offsetY={0}
/>

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

	.selection-info {
		display: flex;
		align-items: center;
	}

	.selection-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.toolbar-content {
		display: flex;
		align-items: center;
		width: 100%;
		justify-content: space-between;
	}

	.toolbar-right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	:global(.on-enter) {
		background-color: var(--imag-80);
		outline: 2px solid var(--imag-primary);
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

	.collection-group-container {
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	.photo-group-container {
		display: flex;
		flex-direction: column;
		padding: 0; /* Remove padding that might push content out of bounds */
		box-sizing: border-box;
		width: 100%;
	}
</style>
