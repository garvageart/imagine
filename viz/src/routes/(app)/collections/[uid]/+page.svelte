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
	import { goto, invalidateAll } from "$app/navigation";
	import { page } from "$app/state";
	import PhotoAssetGrid from "$lib/components/PhotoAssetGrid.svelte";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import SearchInput from "$lib/components/SearchInput.svelte";
	import { debugMode, modal, sort } from "$lib/states/index.svelte";
	import type { AssetGridArray } from "$lib/types/asset.js";
	import { SUPPORTED_IMAGE_TYPES, SUPPORTED_RAW_FILES, type SupportedImageTypes } from "$lib/types/images";
	import hotkeys from "hotkeys-js";
	import { DateTime } from "luxon";
	import { SvelteSet } from "svelte/reactivity";
	import { type ComponentProps } from "svelte";
	import { sortCollectionImages } from "$lib/sort/sort.js";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import UploadManager from "$lib/upload/manager.svelte.js";
	import {
		addCollectionImages,
		getFullImagePath,
		updateCollection,
		deleteCollection,
		deleteCollectionImages,
		type CollectionUpdate,
		type Image
	} from "$lib/api";
	import { thumbHashToDataURL } from "thumbhash";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte.js";
	import CollectionModal from "$lib/components/CollectionModal.svelte";
	import InputText from "$lib/components/dom/InputText.svelte";
	import { layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import Dropdown, { type DropdownOption } from "$lib/components/Dropdown.svelte";
	import { signDownload, downloadImagesBlob, defaults, API_BASE_URL } from "$lib/api/client";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import ImageLightbox from "$lib/components/ImageLightbox.svelte";

	// Context menu state
	let ctxShowMenu = $state(false);
	let ctxItems = $state([] as any[]);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(null as any);

	let { data } = $props();
	// Keyboard events
	const permittedKeys: string[] = [];
	const selectKeys = ["Enter", "Space", " "];
	const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
	permittedKeys.push(...selectKeys, ...moveKeys);

	// Data
	let loadedData = $derived(data);
	// Track local edits separately to avoid clobbering them on refresh
	let localDataUpdates = $state({
		name: data.name,
		description: data.description,
		private: data.private as boolean | undefined
	});

	let loadedImages = $derived(loadedData.images?.items?.map((img) => img.image) ?? []);

	// Sync tab name with collection name
	$effect(() => {
		const currentPath = `/collections/${loadedData.uid}`;
		const newName = localDataUpdates.name;

		if (debugMode) {
			console.log("Syncing tab name:", newName, "for path:", currentPath);
		}

		// Iterate through all panels and their content to find matching view
		for (const panel of layoutState.tree) {
			// Check panel's direct views
			for (const view of panel.views) {
				if (view.path === currentPath) {
					if (debugMode) {
						console.log("Found matching view in panel.views, updating from:", view.name, "to:", newName);
					}

					view.name = newName;
				}
			}

			// Check content views
			if (panel.childs?.content) {
				for (const content of panel.childs.content) {
					for (const view of content.views) {
						if (view.path === currentPath) {
							if (debugMode) {
								console.log("Found matching view in content.views, updating from:", view.name, "to:", newName);
							}

							view.name = newName;
						}
					}
				}
			}
		}
	});

	// Lightbox
	let lightboxImage: Image | undefined = $state();
	let show = $derived(lightboxImage !== undefined);

	// Search stuff
	let searchValue = $state("");
	let searchData = $derived(searchForData(searchValue, loadedImages));

	// Pagination
	// NOTE: This might be moved to a settings thing and this could just be default
	const pagination = $derived({
		limit: loadedData.images?.limit ?? 25,
		page: loadedData.images?.page ?? 0
	});

	// Server tells us if there are more pages via the next field
	// When searching, we're filtering client-side so no more server pages needed
	let shouldUpdate = $derived(!!loadedData.images?.next && searchValue.trim() === "");

	// Selection
	let selectedAssets = $state<SvelteSet<Image>>(new SvelteSet());
	let singleSelectedAsset: Image | undefined = $state();

	let imageGridArray: AssetGridArray<Image> | undefined = $state();

	// Toolbar stuff
	let toolbarOpacity = $state(0);

	// Thumbhash placeholder
	let thumbhashURL = $derived.by(() => {
		if (lightboxImage?.image_metadata?.thumbhash) {
			try {
				const binaryString = atob(lightboxImage.image_metadata.thumbhash);
				const bytes = new Uint8Array(binaryString.length);
				for (let i = 0; i < binaryString.length; i++) {
					bytes[i] = binaryString.charCodeAt(i);
				}
				return thumbHashToDataURL(bytes);
			} catch (error) {
				console.warn("Failed to decode thumbhash:", error);
				return null;
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
	let grid: ComponentProps<typeof PhotoAssetGrid> = $derived({
		photoCardSnippet: imageCard,
		assetGridArray: imageGridArray,
		selectedAssets,
		singleSelectedAsset,
		data: displayData,
		assetGridDisplayProps: {
			style: `padding: 0em ${page.url.pathname === "/" ? "1em" : "2em"};`
		},
		assetDblClick: (_e: MouseEvent, asset: Image) => {
			lightboxImage = asset;
		},
		// Context menu event from PhotoAssetGrid: { asset, anchor }
		onassetcontext: (detail: { asset: Image; anchor: { x: number; y: number } | HTMLElement }) => {
			const { asset, anchor } = detail;
			// Make sure this asset is the only selected one for context actions
			if (!selectedAssets.has(asset) || selectedAssets.size <= 1) {
				singleSelectedAsset = asset;
				selectedAssets.clear();
				selectedAssets.add(asset);
			}

			// Build context menu items for this asset
			ctxItems = [
				{
					id: `download-${asset.uid}`,
					label: "Download",
					icon: "download",
					action: async () => {
						try {
							// Use the server download route so the server can create a short-lived
							// token and redirect the request to the actual file endpoint. The
							// browser will follow the redirect and return the file blob.
							// Build the download URL using the OpenAPI-generated servers value
							// and open it in a new tab so the browser follows redirects and
							// receives the file from the server with proper Content-Disposition.
							const base = API_BASE_URL;
							const dlUrl = new URL(`/images/${asset.uid}/download`, base);

							const a = document.createElement("a");
							a.href = dlUrl.toString();
							a.target = "_blank";
							document.body.appendChild(a);
							a.click();
							a.remove();
						} catch (err) {
							console.error("Context menu download error", err);
							toastState.addToast({ type: "error", message: `Download failed: ${err}` });
						}
					}
				},
				{
					id: `remove-${asset.uid}`,
					label: "Remove from collection",
					icon: "remove_circle",
					action: async () => {
						if (!confirm(`Remove "${asset.name || asset.uid}" from collection "${loadedData.name}"?`)) return;
						try {
							const r = await deleteCollectionImages(loadedData.uid, { uids: [asset.uid] });
							if (r.status === 200) {
								toastState.addToast({ type: "success", message: `Removed from collection` });
								selectedAssets.clear();
								await invalidateAll();
							} else {
								toastState.addToast({ type: "error", message: r.data?.error ?? "Failed to remove" });
							}
						} catch (err) {
							console.error("remove from collection error", err);
							toastState.addToast({ type: "error", message: `Failed to remove: ${err}` });
						}
					}
				},
				{
					id: `copy-${asset.uid}`,
					label: "Copy link",
					icon: "link",
					action: async () => {
						try {
							const url = getFullImagePath(asset.image_paths?.original) ?? "";
							if (url) {
								await navigator.clipboard.writeText(url);
								toastState.addToast({ type: "success", message: "Link copied to clipboard" });
							} else {
								toastState.addToast({ type: "error", message: "No URL available" });
							}
						} catch (err) {
							console.error("copy link error", err);
							toastState.addToast({ type: "error", message: "Failed to copy link" });
						}
					}
				},
				{
					id: `share-${asset.uid}`,
					label: "Share",
					icon: "share",
					action: () => {
						// Placeholder - open share dialog or implement later
						console.log("Share not implemented");
						toastState.addToast({ type: "info", message: "Share not implemented" });
					}
				}
			];

			ctxAnchor = anchor as any;
			ctxShowMenu = true;
		}
	});

	async function handleCollectionUpload() {
		// allowed image types will come from the config but for now just hardcode
		const manager = new UploadManager([...SUPPORTED_RAW_FILES, ...SUPPORTED_IMAGE_TYPES] as SupportedImageTypes[]);

		// Use new API: open picker and upload
		const uploadedImages = await manager.openPickerAndUpload();

		if (uploadedImages.length === 0) {
			return;
		}

		const response = await addCollectionImages(loadedData.uid, {
			uids: uploadedImages.map((img) => img.uid)
		});

		if (response.data.added) {
			toastState.addToast({
				message: `Added ${uploadedImages.length} photo(s) to collection`,
				type: "success",
				timeout: 3000
			});

			await invalidateAll();
		}
	}

	async function updateCollectionDetails(data?: CollectionUpdate) {
		const response = await updateCollection(
			loadedData.uid,
			data ?? {
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

		loadedData = {
			...loadedData,
			updated_at: response.data.updated_at,
			description: response.data.description,
			name: response.data.name,
			private: response.data.private,
			image_count: response.data.image_count,
			created_by: response.data.created_by
		};

		toastState.addToast({
			type: "success",
			message: `Succesfully updated collection ${response.data.name}`
		});
	}

	async function handleDeleteCollection() {
		const ok = confirm(
			`Delete collection "${loadedData.name}"? This will remove the collection record. This action cannot be undone.`
		);

		if (!ok) {
			return;
		}

		try {
			const res = await deleteCollection(loadedData.uid);
			if (res.status === 204) {
				toastState.addToast({
					type: "success",
					message: `Deleted collection ${loadedData.name}`,
					timeout: 3000
				});
				goto("/collections");
			} else {
				const errMsg = (res as any).data?.error ?? (res as any).data?.message ?? "Unknown error";
				toastState.addToast({ type: "error", message: `Failed to delete collection: ${errMsg}` });
			}
		} catch (err) {
			console.error("deleteCollection error", err);
			toastState.addToast({ type: "error", message: `Failed to delete collection: ${err}` });
		}
	}

	async function handleExportPhotos() {
		// Gather all UIDs from the collection and create a download token
		try {
			const uids = loadedImages.map((img) => img.uid);

			if (uids.length === 0) {
				toastState.addToast({ type: "info", message: "No images to export" });
				return;
			}

			// Create a download token (5 minute expiry)
			const signRes = await signDownload({
				uids,
				expires_in: 300,
				allow_download: true,
				allow_embed: false,
				show_metadata: true
			});

			if (signRes.status !== 200) {
				const errMsg = signRes.data?.error ?? "Failed to create download token";
				toastState.addToast({ type: "error", message: errMsg });
				return;
			}

			const token = signRes.data.uid;
			const collectionNameClean = loadedData.name.replace(/[^a-z0-9]/gi, "_").toLowerCase();

			const filename = `${collectionNameClean}-${DateTime.now().toFormat("ddMMyyyy_HHmmss")}.zip`;

			// Use custom downloadImagesBlob function (properly handles binary responses)
			const res = await downloadImagesBlob(token, { uids, filename });

			if (res.status !== 200) {
				const errMsg = res.data?.error ?? "Failed to download images";
				toastState.addToast({ type: "error", message: errMsg });
				return;
			}

			const blob = res.data;
			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);

			toastState.addToast({ type: "success", message: `Exporting ${loadedData.name}`, timeout: 3000 });
		} catch (err) {
			console.error("Export collection error", err);
			toastState.addToast({ type: "error", message: `Failed to export collection: ${err}` });
		}
	}

	async function handleDeleteSelected() {
		// Delete selected images from this collection (client-side selection)
		const items = Array.from(selectedAssets ?? []);
		if (!items || items.length === 0) {
			toastState.addToast({ type: "info", message: "No images selected" });
			return;
		}

		const ok = confirm(`Remove ${items.length} selected image(s) from collection "${loadedData.name}"?`);
		if (!ok) {
			return;
		}

		const uids = items.map((i: Image) => i.uid);
		try {
			const res = await deleteCollectionImages(loadedData.uid, { uids });
			if (res.status === 200 && (res.data?.deleted ?? true)) {
				toastState.addToast({ type: "success", message: `Removed ${uids.length} image(s) from collection`, timeout: 2500 });
				// Clear selection and refresh data
				selectedAssets.clear();
				await invalidateAll();
			} else {
				const errMsg = (res as any).data?.error ?? "Failed to remove images";
				toastState.addToast({ type: "error", message: errMsg });
			}
		} catch (err) {
			console.error("deleteCollectionImages error", err);
			toastState.addToast({ type: "error", message: `Failed to remove images: ${err}` });
		}
	}

	function getDisplayArray(): Image[] {
		return Array.isArray(displayData) ? displayData : (displayData ?? []);
	}

	function prevLightboxImage() {
		if (!lightboxImage) return;
		const arr = getDisplayArray();
		if (!arr.length) return;
		const idx = arr.findIndex((i) => i.uid === lightboxImage!.uid);
		if (idx === -1) return;
		const next = (idx - 1 + arr.length) % arr.length;
		lightboxImage = arr[next];
	}

	function nextLightboxImage() {
		if (!lightboxImage) return;
		const arr = getDisplayArray();
		if (!arr.length) return;
		const idx = arr.findIndex((i) => i.uid === lightboxImage!.uid);
		if (idx === -1) return;
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
		if (!show || !lightboxImage || selectedAssets.size === 0) {
			return;
		}

		e.preventDefault();
		selectedAssets.clear();

		singleSelectedAsset = undefined;
		lightboxImage = undefined;
	});

	const dropdownOptions: DropdownOption[] = [
		{ title: "Export Photos", icon: "download" },
		{ title: "Share Collection", icon: "share" },
		{ title: "Duplicate Collection", icon: "content_copy" },
		{ title: "Delete Collection", icon: "delete" }
	];
</script>

<CollectionModal
	bind:data={localDataUpdates}
	heading="Edit Collection"
	buttonText="Save"
	modalAction={() => {
		updateCollectionDetails();
	}}
/>

<ImageLightbox bind:lightboxImage {prevLightboxImage} {nextLightboxImage} />

{#snippet imageCard(asset: Image)}
	<ImageCard {asset} />
{/snippet}

{#snippet searchInputSnippet()}
	<SearchInput inputId="collection-search" style="margin: 0em 1em; height: 2em;" bind:value={searchValue} />
	<div id="coll-tools">
		<Button
			id="upload_to_collection"
			class="toolbar-button"
			style="padding: 0.1em 0.5em; background-color: var(--imag-100);"
			title="Upload to Collection"
			aria-label="Upload to Collection"
			onclick={() => {
				handleCollectionUpload();
			}}
		>
			Upload
			<MaterialIcon iconName="upload" />
		</Button>
		<Button
			id="upload_to_collection"
			class="toolbar-button"
			style="padding: 0.1em 0.5em; background-color: var(--imag-100);"
			title="Edit Collection"
			aria-label="Edit Collection"
			onclick={() => {
				modal.show = true;
			}}
		>
			Edit
			<MaterialIcon iconName="edit" />
		</Button>
		<Dropdown
			class="toolbar-button"
			icon="more_horiz"
			showSelectionIndicator={false}
			options={dropdownOptions}
			onSelect={(o) => {
				switch (o.title) {
					case "Delete Collection":
						handleDeleteCollection();
						break;
					case "Export Photos":
						handleExportPhotos();
						break;
					case "Share Collection":
						alert("Not implemented yet");
						break;
					case "Duplicate Collection":
						alert("Maybe won't implement");
						break;
				}
			}}
		></Dropdown>
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

{#snippet selectionToolbarSnippet()}
	<Button
		title="Delete Selected"
		style="position: absolute; right: 1em; background-color: var(--imag-100);"
		onclick={handleDeleteSelected}
	>
		<MaterialIcon iconName="delete" />
	</Button>
{/snippet}

<VizViewContainer
	bind:data={displayData}
	bind:hasMore={shouldUpdate}
	name="{localDataUpdates.name} - Collection"
	style="font-size: {page.url.pathname === '/' ? '0.9em' : 'inherit'};"
	paginate={() => {
		// Note: This collection detail page doesn't actually fetch more pages yet
		// The backend loads all images at once. This is just updating the UI state.
		pagination.page++;
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
		gridComponent={PhotoAssetGrid}
		{pagination}
		{noAssetsSnippet}
		{selectionToolbarSnippet}
		toolbarSnippet={searchInputSnippet}
		toolbarProps={{
			style: "justify-content: center; "
		}}
	>
		<div id="viz-info-container">
			<div id="coll-metadata">
				<span id="coll-name">
					<InputText
						id="coll-name-input"
						style="padding: 0% 0.5rem;"
						bind:value={localDataUpdates.name}
						onblur={() => {
							if (localDataUpdates.name !== loadedData.name) {
								updateCollectionDetails({
									name: localDataUpdates.name
								});
							}
						}}
					></InputText>
				</span>
				<span
					id="coll-details"
					style="padding: 0% 0.5rem;"
					title="Updated at: {DateTime.fromJSDate(new Date(loadedData.updated_at)).toFormat('dd.MM.yyyy HH:mm')}"
					>{DateTime.fromJSDate(new Date(loadedData.created_at)).toFormat("dd.MM.yyyy")}
					•
					{#if searchValue.trim()}
						{searchData.length}
						{searchData.length === 1 ? "image" : "images"} of {loadedData.image_count}
					{:else}
						{loadedData.image_count}
						{loadedData.image_count === 1 ? "image" : "images"}
					{/if}
				</span>
			</div>
		</div>
	</AssetsShell>
	<!-- Context menu for right-click on assets -->
	<ContextMenu bind:showMenu={ctxShowMenu} items={ctxItems} anchor={ctxAnchor} offsetY={4} />
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
	}

	#coll-metadata {
		padding: 0.5rem 2rem;
		display: flex;
		flex-direction: column;
		width: 25%;
		overflow: hidden;
		color: var(--imag-60);
		font-family: var(--imag-code-font);
	}

	#coll-tools {
		display: flex;
		align-items: center;
		font-size: inherit;
		height: 100%;
		position: absolute;
		gap: 0.5rem;
		right: 2rem;
	}
</style>
