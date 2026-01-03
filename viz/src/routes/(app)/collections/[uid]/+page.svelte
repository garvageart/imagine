<script module>
	export { searchForData };

	function searchForData(searchValue: string, images: Image[]) {
		if (searchValue.trim() === "") {
			return [];
		}
		// eventually this should also look through keywords/tags
		// and labels idk. fuzzy search???
		return images.filter((i) =>
			i.name.toLowerCase().includes(searchValue.toLowerCase())
		);
	}
</script>

<script lang="ts">
	import { goto, invalidateAll } from "$app/navigation";
	import { page } from "$app/state";
	import PhotoAssetGrid from "$lib/components/PhotoAssetGrid.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import {
		debugMode,
		isLayoutPage,
		modal,
		sort,
		viewSettings
	} from "$lib/states/index.svelte";
	import {
		selectionManager,
		SelectionScopeNames
	} from "$lib/states/selection.svelte";
	import type { AssetGridArray, AssetGridView } from "$lib/types/asset.js";
	import {
		SUPPORTED_IMAGE_TYPES,
		SUPPORTED_RAW_FILES,
		type SupportedImageTypes
	} from "$lib/types/images";
	import hotkeys from "hotkeys-js";
	import { DateTime } from "luxon";
	import { onDestroy, type ComponentProps } from "svelte";
	import { sortCollectionImages } from "$lib/sort/sort.js";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import UploadManager from "$lib/upload/manager.svelte.js";
	import {
		addCollectionImages,
		updateCollection,
		deleteCollection,
		deleteCollectionImages,
		type CollectionUpdate,
		type Image,
		listCollectionImages,
		updateImage
	} from "$lib/api";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte.js";
	import CollectionModal from "$lib/components/modals/CollectionModal.svelte";
	import InputText from "$lib/components/dom/InputText.svelte";
	import Dropdown from "$lib/components/Dropdown.svelte";
	import {
		signDownload,
		downloadImagesZipBlob,
		createCollection
	} from "$lib/api";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import ImageLightbox from "$lib/components/ImageLightbox.svelte";
	import IconButton from "$lib/components/IconButton.svelte";
	import type VizView from "$lib/views/views.svelte";
	import type { PageProps } from "./$types";
	import { filterManager } from "$lib/states/filter.svelte";
	import { untrack } from "svelte";
	import FilterModal from "$lib/components/modals/FilterModal.svelte";
	import type { MenuItem } from "$lib/context-menu/types";
	import { createCollectionImageMenu } from "$lib/context-menu/menus/images";
	import { ImagePaginationState } from "$lib/images/state.svelte";
	import { performImageDownloads } from "$lib/utils/http";
	import LabelSelector from "$lib/components/LabelSelector.svelte";
	import { LabelColours, type ImageLabel } from "$lib/images/constants";
	import { getImageLabel } from "$lib/utils/images";
	import { invalidateViz } from "$lib/views/views.svelte";

	// Context menu state
	let ctxShowMenu = $state(false);
	let ctxItems = $state([] as MenuItem[]);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(
		null as any
	);

	let { data, view }: PageProps & { view?: VizView } = $props();

	let showFilterModal = $state(false);
	let showCollectionModal = $state(false);

	$effect(() => {
		if (debugMode) {
			console.log(
				`[CollectionPage] Mount/Update. View ID: ${view?.id}, View Name: ${view?.name}, Data Name: ${data.name}`
			);
		}
	});

	$effect(() => {
		untrack(() => {
			filterManager.setActiveScopeType("images");
			if (!filterManager.keepFilters) {
				filterManager.resetActiveScope();
			}
		});
	});

	// Keyboard events
	const permittedKeys: string[] = [];
	const selectKeys = ["Enter", "Space", " "];
	const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
	permittedKeys.push(...selectKeys, ...moveKeys);

	// Data
	let localDataUpdates = $state({
		name: untrack(() => data.name),
		description: untrack(() => data.description ?? ""),
		private: untrack(() => data.private ?? false)
	});

	$effect(() => {
		localDataUpdates.name = data.name;
		localDataUpdates.description = data.description ?? "";
		localDataUpdates.private = data.private ?? false;
	});

	let collectionState = $derived(new ImagePaginationState(data.images));
	let isPaginating = $state(false);

	async function paginate() {
		if (isPaginating || !collectionState.hasMore) {
			return;
		}

		isPaginating = true;
		const nextPage = collectionState.pagination.page + 1;
		const res = await listCollectionImages(data.uid, {
			limit: collectionState.pagination.limit,
			offset: nextPage * collectionState.pagination.limit
		});

		if (res.status === 200) {
			const nextItems = res.data.items?.map((i) => i.image) ?? [];
			collectionState.images.push(...nextItems);

			// Update pagination state from response
			collectionState.pagination.page = res.data.page ?? nextPage;
			collectionState.totalCount = res.data.count ?? collectionState.totalCount;
			collectionState.hasMore = !!res.data.next;
		} else {
			// Avoid infinite loop on failure
			toastState.addToast({
				type: "error",
				title: `Image Load Failure: ${res.status}`,
				message: `Failed to load more images for collection: ${res.data?.error ?? "Unknown error"}`
			});
			collectionState.hasMore = false;
		}

		isPaginating = false;
	}

	// Sync tab name with collection name directly on the passed view instance
	$effect(() => {
		if (view && localDataUpdates.name) {
			// Ensure we aren't applying stale data to a new view
			if (view.path && !view.path.includes(data.uid)) {
				return;
			}

			if (debugMode) {
				console.log(
					`Syncing tab name to "${localDataUpdates.name}" for view ${view.id}. Data Name: ${data.name}`
				);
			}
			view.name = localDataUpdates.name;
		}
	});

	// Lightbox
	let lightboxImage: Image | undefined = $state();
	let show = $derived(lightboxImage !== undefined);

	// Search stuff
	let searchValue = $state("");
	let searchData = $derived(searchForData(searchValue, collectionState.images));

	// Selection
	const scopeId = $derived(SelectionScopeNames.COLLECTION_PREFIX + data.uid);
	const selectionScope = $derived(selectionManager.getScope<Image>(scopeId));
	let selectionFirstImage = $derived(Array.from(selectionScope.selected)[0]);

	onDestroy(() => {
		selectionManager.removeScope(scopeId);
	});

	let imageGridArray: AssetGridArray<Image> | undefined = $state();

	// Toolbar stuff
	let toolbarOpacity = $state(0);

	// Display Data
	let displayData = $derived(
		searchValue.trim()
			? sortCollectionImages(searchData, sort)
			: sortCollectionImages(filterManager.apply(collectionState.images), sort)
	);

	// Grid props
	let grid: ComponentProps<typeof PhotoAssetGrid> = $derived({
		photoCardSnippet: imageCard,
		view: viewSettings.current,
		assetGridArray: imageGridArray,
		data: displayData,
		scopeId: scopeId,
		assetGridDisplayProps: {
			style: `padding: 0em ${isLayoutPage() ? "1em" : "2em"};`
		},
		assetDblClick: (_e: MouseEvent, asset: Image) => {
			lightboxImage = asset;
		},
		// Context menu event from PhotoAssetGrid: { asset, anchor }
		onassetcontext: (detail: {
			asset: Image;
			anchor: { x: number; y: number } | HTMLElement;
		}) => {
			const { asset, anchor } = detail;
			// Make sure this asset is the only selected one for context actions
			if (!selectionScope.has(asset) || selectionScope.selected.size <= 1) {
				selectionScope.select(asset);
			}

			console.log("asset", $state.snapshot(asset));
			ctxItems = createCollectionImageMenu(asset, data, {
				downloadImages() {
					performImageDownloads([asset]);
				}
			});

			ctxAnchor = anchor as any;
			ctxShowMenu = true;
		}
	});

	async function handleCollectionUpload() {
		// allowed image types will come from the config but for now just hardcode
		const manager = new UploadManager([
			...SUPPORTED_RAW_FILES,
			...SUPPORTED_IMAGE_TYPES
		] as SupportedImageTypes[]);

		const files = await manager.openPicker();

		if (files.length === 0) {
			return;
		}

		const uploadedTasks = manager.addFiles(files);
		manager.start();

		const uploadedSuccesses = await Promise.all(
			uploadedTasks.map((task) => task.imageData)
		);

		const uids = uploadedSuccesses
			.map((success) => success?.uid)
			.filter((v) => v !== undefined);

		const response = await addCollectionImages(data.uid, {
			uids: uids
		});

		if (response.data.added) {
			toastState.addToast({
				message: `Added ${uids.length} photo(s) to collection`,
				type: "success",
				timeout: 3000
			});

			await invalidateViz();
		}
	}

	async function updateCollectionDetails(updateData?: CollectionUpdate) {
		const response = await updateCollection(
			data.uid,
			updateData ?? {
				...localDataUpdates
			}
		);

		modal.show = false;
		if (response.status !== 200) {
			toastState.addToast({
				type: "error",
				message: `Failed to update collection: ${response.data || "Unknown error"}`
			});

			return;
		}

		await invalidateViz();

		toastState.addToast({
			title: response.data.name,
			type: "success",
			message: `Successfully updated collection`
		});
	}

	async function handleDeleteCollection() {
		const ok = confirm(
			`Delete collection "${data.name}"? This will remove the collection record. This action cannot be undone.`
		);

		if (!ok) {
			return;
		}

		try {
			const res = await deleteCollection(data.uid);
			if (res.status === 204) {
				toastState.addToast({
					type: "success",
					message: `Deleted collection ${data.name}`,
					timeout: 3000
				});
				goto("/collections");
			} else {
				const errMsg = res.data.error ?? "Unknown error";
				toastState.addToast({
					type: "error",
					message: `Failed to delete collection: ${errMsg}`
				});
			}
		} catch (err) {
			console.error("deleteCollection error", err);
			toastState.addToast({
				type: "error",
				message: `Failed to delete collection: ${err}`
			});
		}
	}

	async function handleDeleteSelected() {
		// Delete selected images from this collection (client-side selection)
		const items = Array.from(selectionScope.selected ?? []);
		if (!items || items.length === 0) {
			toastState.addToast({ type: "info", message: "No images selected" });
			return;
		}

		const ok = confirm(
			`Remove ${items.length} selected image(s) from collection "${data.name}"?`
		);
		if (!ok) {
			return;
		}

		const uids = items.map((i: Image) => i.uid);
		try {
			const res = await deleteCollectionImages(data.uid, { uids });
			if (res.status === 200 && (res.data?.deleted ?? true)) {
				toastState.addToast({
					type: "success",
					message: `Removed ${uids.length} image(s) from collection`,
					timeout: 2500
				});
				// Clear selection and refresh data
				selectionScope.clear();
				await invalidateViz();
			} else {
				const errMsg = (res as any).data?.error ?? "Failed to remove images";
				toastState.addToast({ type: "error", message: errMsg });
			}
		} catch (err) {
			console.error("deleteCollectionImages error", err);
			toastState.addToast({
				type: "error",
				message: `Failed to remove images: ${err}`
			});
		}
	}

	async function handleDuplicateCollection() {
		try {
			const res = await createCollection({
				name: `Copy of ${data.name}`,
				description: data.description ?? undefined,
				private: data.private ?? false
			});

			if (res.status === 201) {
				const newCollectionUid = res.data.uid;
				const uidsToCopy = collectionState.images.map((img) => img.uid);

				if (uidsToCopy.length > 0) {
					const addRes = await addCollectionImages(newCollectionUid, {
						uids: uidsToCopy
					});
					if (addRes.status === 200) {
						toastState.addToast({
							message: "Collection duplicated with images",
							type: "success"
						});
						await invalidateViz();
						goto(`/collections/${newCollectionUid}`);
					} else {
						toastState.addToast({
							message: `Collection duplicated but failed to copy images (${addRes.status})`,
							type: "warning"
						});
						goto(`/collections/${newCollectionUid}`); // Still navigate to the new collection
					}
				} else {
					toastState.addToast({
						message: "Collection duplicated (no images to copy)",
						type: "success"
					});
					await invalidateViz();
					goto(`/collections/${newCollectionUid}`);
				}
			} else {
				toastState.addToast({
					title: "Duplicate Collection Failed",
					message:
						res.data.error ?? "Unknown error occurred during duplication",
					type: "error"
				});
			}
		} catch (err) {
			toastState.addToast({
				title: "Duplicate Collection Failed",
				message:
					(err as Error).message ?? "Unknown error occurred during duplication",
				type: "error"
			});
		}
	}

	function getDisplayArray(): Image[] {
		return Array.isArray(displayData) ? displayData : (displayData ?? []);
	}

	function prevLightboxImage() {
		if (!lightboxImage) {
			return;
		}

		const arr = getDisplayArray();
		if (!arr.length) {
			return;
		}

		const idx = arr.findIndex((i) => i.uid === lightboxImage!.uid);
		if (idx === -1) {
			return;
		}

		const next = (idx - 1 + arr.length) % arr.length;
		lightboxImage = arr[next];
	}

	function nextLightboxImage() {
		if (!lightboxImage) {
			return;
		}

		const arr = getDisplayArray();
		if (!arr.length) {
			return;
		}

		const idx = arr.findIndex((i) => i.uid === lightboxImage!.uid);
		if (idx === -1) {
			return;
		}

		const next = (idx + 1) % arr.length;
		lightboxImage = arr[next];
	}

	hotkeys("left,right", (e, handler) => {
		if (!show) {
			return;
		}

		e.preventDefault();
		if (handler.key === "left") {
			prevLightboxImage();
		} else if (handler.key === "right") {
			nextLightboxImage();
		}
	});

	hotkeys("escape", (e) => {
		if (!show || !lightboxImage || selectionScope.selected.size === 0) {
			return;
		}

		e.preventDefault();
		selectionScope.clear();

		lightboxImage = undefined;
	});

	// Menu items for collection actions
	let collectionMenuItems: MenuItem[] = $derived(ctxItems);

	// Display options as MenuItem[] for Dropdown
	let displayMenuItems: MenuItem[] = $derived(
		viewSettings.displayOptions.map((o, idx) => ({
			id: `display-${idx}`,
			label: o.label,
			icon: o.icon,
			action: () => viewSettings.setView(o.label.toLowerCase() as AssetGridView)
		}))
	);

	let displaySelectedId: string | undefined = $derived.by(() => {
		const idx = viewSettings.displayOptions.findIndex(
			(o) => o.label.toLowerCase() === viewSettings.current
		);
		return idx !== -1 ? `display-${idx}` : undefined;
	});
</script>

{#if showCollectionModal && modal.show}
	<CollectionModal
		bind:data={localDataUpdates}
		heading="Edit Collection"
		buttonText="Save"
		modalAction={() => {
			updateCollectionDetails();
		}}
	/>
{/if}

{#if showFilterModal && modal.show}
	<FilterModal />
{/if}

<ImageLightbox bind:lightboxImage {prevLightboxImage} {nextLightboxImage} />

{#snippet imageCard(asset: Image)}
	<ImageCard {asset} />
{/snippet}

{#snippet toolbarSnippet()}
	<!-- This looks like ass -->
	<!-- <SearchInput
		inputId="collection-search"
		bind:value={searchValue}
		placeholder="Search images"
		style="font-size: 1.1em;"
	/> -->
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
			iconName="upload"
			id="upload_to_collection"
			class="toolbar-button"
			title="Upload to Collection"
			aria-label="Upload to Collection"
			onclick={() => {
				handleCollectionUpload();
			}}
		>
			Upload
		</IconButton>
		<IconButton
			iconName="edit"
			id="upload_to_collection"
			class="toolbar-button"
			title="Edit Collection"
			aria-label="Edit Collection"
			onclick={() => {
				showCollectionModal = true;
				modal.show = true;
			}}
		>
			Edit
		</IconButton>
		<Dropdown
			title="Display"
			class="toolbar-button"
			icon="list_alt"
			items={displayMenuItems}
			selectedItemId={displaySelectedId}
		/>
		<Dropdown
			class="toolbar-button"
			icon="more_horiz"
			showSelectionIndicator={false}
			items={collectionMenuItems}
		/>
	</div>
{/snippet}

{#snippet noAssetsSnippet()}
	<div id="add_to_collection-container">
		<span style="margin: 1em; color: var(--imag-20); font-size: 1.2rem;"
			>Add images to this collection</span
		>
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

{#snippet selectionToolbarSnippet()}
	<div class="selection-actions">
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

				console.log("labelToSend", labelToSend);

				const updatePromises = Array.from(selectionScope.selected).map((img) =>
					updateImage(img.uid, {
						image_metadata: { label: labelToSend }
					})
				);

				const res = await Promise.all(updatePromises);

				const successCount = res.filter((r) => r.status === 200).length;
				if (successCount > 0) {
					res.forEach((r) => {
						if (r.status === 200) {
							const updatedImage = r.data;
							const imageIndex = collectionState.images.findIndex(
								(img) => img.uid === updatedImage.uid
							);
							if (imageIndex !== -1) {
								collectionState.images[imageIndex] = updatedImage;
							}
						}
					});
				}
			}}
		/>
	</div>

	<IconButton
		iconName="delete"
		title="Delete Selected"
		style="position: absolute; right: 1em; background-color: var(--imag-100);"
		onclick={handleDeleteSelected}
	/>
{/snippet}

<VizViewContainer
	bind:data={displayData}
	hasMore={collectionState.hasMore}
	name="{localDataUpdates.name} - Collection"
	style="font-size: {page.url.pathname === '/' ? '0.9em' : 'inherit'};"
	{paginate}
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
		gridComponent={PhotoAssetGrid}
		pagination={collectionState.pagination}
		{noAssetsSnippet}
		{selectionToolbarSnippet}
		{toolbarSnippet}
		toolbarProps={{
			style: "justify-content: space-between; gap: 0.5rem;"
		}}
	>
		<div id="viz-info-container">
			<div id="coll-metadata">
				<span id="coll-name">
					<InputText
						autocorrect="off"
						spellcheck="false"
						id="coll-name-input"
						style="padding: 0% 0.5rem;"
						title={localDataUpdates.name}
						bind:value={localDataUpdates.name}
					/>
					{#if localDataUpdates.name.trim() === ""}
						<MaterialIcon
							iconName="warning"
							style="font-size: 0.9rem;"
							title="Name cannot be empty"
						/>
					{:else}
						<div
							id="confirm-icons"
							style:visibility={localDataUpdates.name.trim() ===
							data.name.trim()
								? "hidden"
								: "visible"}
						>
							<IconButton
								title="Cancel"
								class="name-confirm-btn"
								onclick={() => {
									localDataUpdates.name = data.name;
								}}
								iconName="close"
							/>
							<IconButton
								title="Confirm"
								class="name-confirm-btn"
								onclick={() => {
									updateCollectionDetails({
										name: localDataUpdates.name
									});
								}}
								iconName="check"
							/>
						</div>
					{/if}
				</span>
				<span
					id="coll-details"
					style="padding: 0% 0.5rem;"
					title="Updated at: {DateTime.fromJSDate(
						new Date(data.updated_at)
					).toFormat('dd.MM.yyyy HH:mm')}"
					>{DateTime.fromJSDate(new Date(data.created_at)).toFormat(
						"dd.MM.yyyy"
					)}
					•
					{#if searchValue.trim()}
						{searchData.length}
						{searchData.length === 1 ? "image" : "images"} of {data.image_count}
					{:else}
						{data.image_count}
						{data.image_count === 1 ? "image" : "images"}
					{/if}
				</span>
			</div>
		</div>
	</AssetsShell>
	<!-- Context menu for right-click on assets -->
	<ContextMenu
		bind:showMenu={ctxShowMenu}
		items={ctxItems}
		anchor={ctxAnchor}
		offsetY={4}
	/>
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

	#coll-name {
		color: var(--imag-20);
		font-weight: bold;
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;

		#confirm-icons {
			display: flex;
			flex-direction: row;
		}

		:global(.name-confirm-btn) {
			font-size: 0.75rem;
		}
	}

	#coll-metadata {
		padding: 0.5rem 1rem;
		display: flex;
		flex-direction: column;
		width: 30%;
		overflow: hidden;
		color: var(--imag-60);
		font-family: var(--imag-code-font);
		gap: 1rem;
	}

	.selection-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin: auto 1rem;
	}

	#coll-tools {
		display: flex;
		align-items: center;
		font-size: inherit;
		height: 100%;
		gap: 0.75rem;
	}
</style>
