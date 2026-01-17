<script lang="ts">
	import {
		addCollectionImages,
		getImage,
		listImages,
		updateImage,
		type Collection,
		type Image
	} from "$lib/api";
	import AssetToolbar from "$lib/components/AssetToolbar.svelte";
	import Button from "$lib/components/Button.svelte";
	import DragAndDropUpload from "$lib/components/DragAndDropUpload.svelte";
	import Dropdown from "$lib/components/Dropdown.svelte";
	import IconButton from "$lib/components/IconButton.svelte";
	import ImageLightbox from "$lib/components/ImageLightbox.svelte";
	import LabelSelector from "$lib/components/LabelSelector.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import PhotoAssetGrid from "$lib/components/PhotoAssetGrid.svelte";
	import CollectionSelectionModal from "$lib/components/modals/CollectionSelectionModal.svelte";
	import FilterModal from "$lib/components/modals/FilterModal.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import { VizMimeTypes } from "$lib/constants.js";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import { createImageMenu } from "$lib/context-menu/menus/images.js";
	import type { MenuItem } from "$lib/context-menu/types";
	import { DragData } from "$lib/drag-drop/data.js";
	import { LabelColours, type ImageLabel } from "$lib/images/constants.js";
	import { ImagePaginationState } from "$lib/images/state.svelte.js";
	import {
		getConsolidatedGroups,
		groupImagesByDate,
		type ConsolidatedGroup,
		type DateGroup
	} from "$lib/photo-layout/index.js";
	import { filterManager } from "$lib/states/filter.svelte";
	import { modal, viewSettings } from "$lib/states/index.svelte";
	import {
		selectionManager,
		SelectionScopeNames
	} from "$lib/states/selection.svelte";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import {
		SUPPORTED_IMAGE_TYPES,
		SUPPORTED_RAW_FILES,
		type SupportedImageTypes
	} from "$lib/types/images";
	import UploadManager, {
		type ImageUploadSuccess
	} from "$lib/upload/manager.svelte";
	import { performImageDownloads } from "$lib/utils/http.js";
	import { getImageLabel } from "$lib/utils/images.js";
	import StarRating from "$lib/components/StarRating.svelte";
	import hotkeys from "hotkeys-js";
	import { onDestroy, untrack } from "svelte";
	import { goto } from "$app/navigation";

	// Display options as MenuItem[] for Dropdown
	const displayMenuItems: MenuItem[] = [
		{ id: "display-grid", label: "Grid" },
		{ id: "display-list", label: "List" },
		{ id: "display-cards", label: "Thumbnails" }
	];

	function getDisplaySelectedId(): string | undefined {
		const map: Record<string, string> = {
			grid: "display-grid",
			list: "display-list",
			thumbnails: "display-cards"
		};
		return map[(viewSettings.current as string) ?? ""];
	}

	let { data } = $props();

	$effect(() => {
		untrack(() => {
			filterManager.setActiveScopeType("images");
			if (!filterManager.keepFilters) {
				filterManager.resetActiveScope();
			}
		});
	});

	$effect(() => {
		if (!modal.show) {
			showFilterModal = false;
			showCollectionSelectionModal = false;
		}
	});

	let galleryState = $derived(new ImagePaginationState(data));
	let isPaginating = $state(false);

	// Page state
	let groups: DateGroup[] = $derived(
		groupImagesByDate(filterManager.apply(galleryState.images)) ?? []
	);

	let consolidatedGroups: ConsolidatedGroup[] = $derived.by(() => {
		return getConsolidatedGroups(groups);
	});

	// Lightbox
	let lightboxImage: Image | undefined = $state();

	// Selection (shared across groups)
	const scopeId = SelectionScopeNames.PHOTOS_MAIN;
	const selectionScope = selectionManager.getScope<Image>(scopeId);
	let selectionFirstImage = $derived(Array.from(selectionScope.selected)[0]);

	onDestroy(() => {
		selectionManager.removeScope(scopeId);
	});

	// Flat list of all images for cross-group range selection
	let allImagesFlat = $derived(consolidatedGroups.flatMap((g) => g.allImages));

	// Modal state for collection selection
	let showCollectionSelectionModal = $state(false);
	let imageUidsForCollection = $state<string[]>([]);

	// Action menu items for selected images
	let actionMenuItems: MenuItem[] = $derived.by(() => {
		const baseMenuItems = createImageMenu(galleryState.images, selectionScope, {
			onDelete: (deletedUIDs) => {
				galleryState.images = galleryState.images.filter(
					(img) => !deletedUIDs.includes(img.uid)
				);
				galleryState.totalCount -= deletedUIDs.length;
			}
		});
		const pageMenuItems: MenuItem[] = [
			{
				id: "act-add-to-collection",
				label: "Add to Collection",
				icon: "collections_bookmark",
				action: () => {
					imageUidsForCollection = Array.from(selectionScope.selected).map(
						(img) => img.uid
					);
					showCollectionSelectionModal = true;
					modal.show = true;
				}
			}
		];

		return [...pageMenuItems, ...baseMenuItems];
	});

	// Context menu state for right-click on assets
	let ctxShowMenu = $state(false);
	let ctxItems: MenuItem[] = $derived(actionMenuItems);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(null);

	async function paginate() {
		if (isPaginating || !galleryState.hasMore) {
			return;
		}

		isPaginating = true;
		const nextPage = galleryState.pagination.page + 1;
		const res = await listImages({
			limit: galleryState.pagination.limit,
			page: nextPage
		});

		if (res.status === 200) {
			const nextItems = res.data.items?.map((i) => i.image) ?? [];
			galleryState.images.push(...nextItems);

			// Update pagination state from response
			galleryState.pagination.page = res.data.page ?? nextPage;
			galleryState.totalCount = res.data.count ?? galleryState.totalCount;
			galleryState.hasMore = !!res.data.next;
		} else {
			// On error, avoid tight loops; allow retry on next scroll
			console.error("paginate: request failed", res);
			galleryState.hasMore = false;
		}

		isPaginating = false;
	}

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

	// Upload confirmation state
	let showUploadConfirm = $state(false);

	let showFilterModal = $state(false);

	let pendingNewRaw: ImageUploadSuccess[] = [];
	let addImagesDebounceTimer: number | undefined;
	const ADD_IMAGES_DEBOUNCE_MS = 550;

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
				const skippedCount =
					imageUidsForCollection.length - newImageUids.length;
				let message = `Added ${newImageUids.length} image(s) to collection **${collection.name}**`;
				if (skippedCount > 0) {
					message += `. Skipped ${skippedCount} existing image(s).`;
				}
				toastState.addToast({
					type: "success",
					message: message,
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
			selectionScope.clear();
			showCollectionSelectionModal = false;
			modal.show = false;
			imageUidsForCollection = [];
		}
	}

	async function resolveRawToImages(
		items: ImageUploadSuccess[]
	): Promise<Image[]> {
		if (!items || items.length === 0) {
			return [];
		}

		const results: Image[] = [];
		const fetchPromises = items.map(async (it) => {
			if (!it) {
				return null;
			}

			const uid = it.uid;
			if (!uid) {
				return null;
			}

			try {
				const res = await getImage(uid);

				if (res.status === 200) {
					return res.data;
				}
				console.warn(
					`Failed to fetch image metadata for ${uid}: ${res.data.error}`
				);
				return null;
			} catch (err) {
				console.warn("Failed to fetch image metadata for", uid, err);
				return null;
			}
		});

		const fetched = await Promise.all(fetchPromises);
		for (const f of fetched) {
			if (f) {
				results.push(f);
			}
		}

		return results;
	}

	function scheduleAddImages(newRaw: ImageUploadSuccess[]) {
		if (!newRaw || newRaw.length === 0) {
			return;
		}

		pendingNewRaw.push(...newRaw);

		if (addImagesDebounceTimer) {
			clearTimeout(addImagesDebounceTimer);
		}

		addImagesDebounceTimer = window.setTimeout(async () => {
			const batch = pendingNewRaw.slice();
			pendingNewRaw = [];
			addImagesDebounceTimer = undefined;

			const imagesToAdd = await resolveRawToImages(batch);
			if (imagesToAdd.length > 0) {
				galleryState.images.push(...imagesToAdd);
			}
		}, ADD_IMAGES_DEBOUNCE_MS) as unknown as number;
	}

	async function addImagesToImagine() {
		const manager = new UploadManager([
			...SUPPORTED_RAW_FILES,
			...SUPPORTED_IMAGE_TYPES
		] as SupportedImageTypes[]);
		const uploadedImages = await manager.openPickerAndUpload();

		if (uploadedImages.length === 0) {
			return;
		}

		scheduleAddImages(uploadedImages);
	}

	hotkeys("escape", (e) => {
		e.preventDefault();
		selectionScope.clear();

		lightboxImage = undefined;

		if (modal.show) {
			modal.show = false;
			showUploadConfirm = false;
			showCollectionSelectionModal = false;
		}
	});
</script>

<svelte:head>
	<title>Photos</title>
</svelte:head>

<DragAndDropUpload {scopeId} {selectionScope} showCollectionCreateBox={true} />

{#if lightboxImage}
	<ImageLightbox
		bind:lightboxImage
		{prevLightboxImage}
		{nextLightboxImage}
		onImageUpdated={(image) =>
			selectionScope.updateItem(image, galleryState.images)}
	/>
{/if}

{#if showFilterModal && modal.show}
	<FilterModal />
{/if}

{#if showCollectionSelectionModal}
	<CollectionSelectionModal
		bind:showModal={showCollectionSelectionModal}
		onSelect={handleCollectionSelect}
		imageUidsToAdd={imageUidsForCollection}
	/>
{/if}

{#snippet noAssetsSnippet()}
	<div id="add_to_imagine-container">
		<span style="margin: 1em; color: var(--imag-20); font-size: 1.2rem;"
			>Add your first images</span
		>
		<Button
			id="add_to_collection-button"
			style="padding: 2em 8em; display: flex; align-items: center; justify-content: center;"
			title="Select Photos"
			aria-label="Select Photos"
			onclick={async () => addImagesToImagine()}
		>
			Select Photos
			<MaterialIcon iconName="add" style="font-size: 2em;" />
		</Button>
	</div>
{/snippet}

<VizViewContainer
	name="Photos"
	bind:data={galleryState.images}
	hasMore={galleryState.hasMore}
	paginate={() => paginate()}
>
	{#if galleryState.images.length > 0}
		{#if selectionScope.selected.size > 0}
			<AssetToolbar class="selection-toolbar" stickyToolbar={true}>
				<div class="selection-info">
					<IconButton
						iconName="close"
						class="toolbar-button"
						title="Clear selection"
						aria-label="Clear selection"
						style="margin-right: 1em;"
						onclick={() => selectionScope.clear()}
					/>
					<span style="font-weight: 600;"
						>{selectionScope.selected.size} selected</span
					>
				</div>
				<div class="selection-actions">
					<IconButton
						iconName={actionMenuItems.find(
							(it) => it.id === "act-add-to-collection"
						)?.icon ?? "collections_bookmark"}
						class="action"
						role="tooltip"
						title="Add to Collection"
						onclick={() => {
							imageUidsForCollection = Array.from(selectionScope.selected).map(
								(img) => img.uid
							);
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
							if (!e.dataTransfer?.types.includes(VizMimeTypes.IMAGE_UIDS)) {
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
					<LabelSelector
						variant="expanded"
						label={getImageLabel(selectionFirstImage)}
						onSelect={async (selectedLabel) => {
							if (!selectionFirstImage) {
								return;
							}

							// Reverse lookup: find the key (Name) for the selected color (Value)
							const entry = Object.entries(LabelColours).find(
								([_, colour]) => colour === selectedLabel
							);
							const labelName = entry ? entry[0] : null;
							// If "None" is selected, send null to clear the label
							const labelToSend = (
								labelName === "None" || !labelName ? null : labelName
							) as ImageLabel | null;

							const updatePromises = Array.from(selectionScope.selected).map(
								(img) =>
									updateImage(img.uid, {
										image_metadata: { label: labelToSend }
									})
							);

							const res = await Promise.all(updatePromises);

							const successCount = res.filter((r) => r.status === 200).length;
							if (successCount > 0) {
								res.forEach((r) => {
									if (r.status === 200) {
										selectionScope.updateItem(r.data, galleryState.images);
									}
								});
							}
						}}
					/>
					<StarRating
						value={selectionFirstImage?.image_metadata?.rating ?? 0}
						onChange={async (rating) => {
							if (!selectionFirstImage) {
								return;
							}

							const updatePromises = Array.from(selectionScope.selected).map(
								(img) =>
									updateImage(img.uid, {
										image_metadata: { rating }
									})
							);

							const res = await Promise.all(updatePromises);

							const successCount = res.filter((r) => r.status === 200).length;
							if (successCount > 0) {
								res.forEach((r) => {
									if (r.status === 200) {
										selectionScope.updateItem(r.data, galleryState.images);
									}
								});
							}
						}}
					/>
				</div>
				<div
					style="margin-left: auto; display: flex; gap: 0.5rem; align-items: center;"
				>
					<Dropdown
						class="toolbar-button"
						icon="more_horiz"
						items={actionMenuItems}
						showSelectionIndicator={false}
						align="right"
					/>
				</div>
			</AssetToolbar>
		{:else}
			<AssetToolbar
				style="position: sticky; top: 0px; display: flex; justify-content: flex-end;"
				stickyToolbar={true}
			>
				<div style="display: flex; align-items: center; gap: 0.5rem;">
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
					<Dropdown
						title="Display"
						class="toolbar-button"
						icon="list_alt"
						items={displayMenuItems}
						selectedItemId={getDisplaySelectedId()}
					/>
				</div>
			</AssetToolbar>
		{/if}
	{/if}
	{#if groups.length === 0}
		<div id="viz-no_assets">
			{@render noAssetsSnippet()}
		</div>
	{:else}
		<div class="photo-group-container">
			{#each consolidatedGroups as consolidatedGroup}
				<section class="photo-group">
					<h2 class="photo-group-label">{consolidatedGroup.label}</h2>

					<PhotoAssetGrid
						bind:allData={allImagesFlat}
						bind:view={viewSettings.current}
						data={consolidatedGroup.allImages}
						{scopeId}
						assetDblClick={(_e, asset) => {
							openLightbox(asset);
						}}
						onassetcontext={(detail) => {
							ctxAnchor = detail.anchor;
							ctxShowMenu = true;
						}}
					/>
				</section>
			{/each}
		</div>
	{/if}
</VizViewContainer>

<!-- Context menu for right-click on assets -->
<ContextMenu
	bind:showMenu={ctxShowMenu}
	items={ctxItems}
	anchor={ctxAnchor}
	offsetY={0}
/>

<style lang="scss">
	.photo-group-container {
		display: flex;
		flex-direction: column;
		padding: 2rem 2rem;
		box-sizing: border-box;
		width: 100%;
	}

	.selection-info {
		display: flex;
		align-items: center;
	}

	:global(.selection-toolbar) {
		gap: 1rem;
	}

	:global(.on-enter) {
		background-color: var(--imag-80);
		outline: 2px solid var(--imag-primary);
	}

	.selection-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.photo-group {
		display: flex;
		flex-direction: column;
		box-sizing: border-box;
		margin-bottom: 1rem;
		width: 100%;

		h2 {
			font-weight: 400;
			font-size: 1.2rem;
			color: var(--imag-10);
			width: fit-content;
		}
	}

	#add_to_imagine-container {
		display: flex;
		flex-direction: column;
		justify-content: left;
	}

	#viz-no_assets {
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}
</style>
