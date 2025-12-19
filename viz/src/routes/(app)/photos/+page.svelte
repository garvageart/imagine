<script lang="ts">
	import PhotoAssetGrid, {
		type AssetGridView
	} from "$lib/components/PhotoAssetGrid.svelte";
	import LoadingContainer from "$lib/components/LoadingContainer.svelte";
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import { type Image } from "$lib/api";
	import { DateTime } from "luxon";
	import { getFullImagePath } from "$lib/api";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { SvelteSet } from "svelte/reactivity";
	import {
		listImages,
		deleteImagesBulk,
		signDownload,
		downloadImagesZipBlob,
		getImage
	} from "$lib/api";
	import AssetToolbar from "$lib/components/AssetToolbar.svelte";
	import Dropdown, {
		type DropdownOption
	} from "$lib/components/Dropdown.svelte";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import { fade } from "svelte/transition";
	import hotkeys from "hotkeys-js";
	import UploadManager, {
		type ImageUploadSuccess,
		waitForUploadCompletion
	} from "$lib/upload/manager.svelte";
	import { UploadState } from "$lib/upload/asset.svelte";
	import { createCollection, addCollectionImages } from "$lib/api";
	import {
		SUPPORTED_IMAGE_TYPES,
		SUPPORTED_RAW_FILES,
		type SupportedImageTypes
	} from "$lib/types/images";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import { goto, invalidateAll } from "$app/navigation";
	import ImageLightbox from "$lib/components/ImageLightbox.svelte";
	import Button from "$lib/components/Button.svelte";
	import { modal } from "$lib/states/index.svelte";
	import { copyToClipboard } from "$lib/utils/misc.js";
	import CollectionModal from "$lib/components/CollectionModal.svelte";
	import ConfirmationModal from "$lib/components/modals/ConfirmationModal.svelte";
	import { downloadOriginalImageFile } from "$lib/utils/http.js";
	import {
		getConsolidatedGroups,
		groupImagesByDate,
		type ConsolidatedGroup,
		type DateGroup,
		type ImageWithDateLabel
	} from "$lib/photo-layout/index.js";

	let { data } = $props();

	// Pagination
	const pagination = $state({
		limit: data.limit,
		page: data.page
	});

	let images: Image[] = $state(data.images ?? []);
	let totalCount = $state<number>(data.count ?? 0);
	let hasMore = $state(!!data.next);
	let isPaginating = $state(false);

	// Page state
	let groups: DateGroup[] = $derived(groupImagesByDate(images) ?? []);

	let consolidatedGroups: ConsolidatedGroup[] = $derived.by(() => {
		return getConsolidatedGroups(groups);
	});

	// Display state (local to photos page)
	let currentAssetGridView: AssetGridView = $state("grid");
	const displayOptions: DropdownOption[] = [
		{ title: "Grid" },
		{ title: "List" },
		{ title: "Card" }
	];

	// Lightbox
	let lightboxImage: Image | undefined = $state();
	let show = $derived(!!lightboxImage);
	// Selection (shared across groups)
	let selectedAssets = $state(new SvelteSet<Image>());
	let singleSelectedAsset: Image | undefined = $state();

	// Flat list of all images for cross-group range selection
	let allImagesFlat = $derived(consolidatedGroups.flatMap((g) => g.allImages));

	// UI state: show a small spinner while a download is in progress
	let downloadInProgress = $state(false);

	// Context menu state for right-click on assets
	let ctxShowMenu = $state(false);
	let ctxItems: any[] = $state([]);
	let ctxAnchor: { x: number; y: number } | HTMLElement | null = $state(null);

	// Action menu options for selected images
	const actionMenuOptions: DropdownOption[] = [
		{ title: "Download", icon: "download" },
		{ title: "Add to Collection", icon: "collections_bookmark" },
		{ title: "Share", icon: "share" },
		{ title: "Copy Link", icon: "link" },
		{ title: "Edit Metadata", icon: "edit" },
		{ title: "Move to Trash", icon: "delete" },
		{ title: "Force Delete", icon: "delete_forever" }
	];

	async function handleActionMenu(option: DropdownOption) {
		const items = Array.from(selectedAssets);
		if (items.length === 0) {
			toastState.addToast({
				type: "warning",
				message: "No images selected",
				timeout: 3000
			});
			return;
		}

		switch (option.title) {
			case "Download":
				downloadInProgress = true;
				try {
					const uids = items.map((i) => i.uid);
					await performImageDownloads(uids);
				} catch (err) {
					console.error("Download error", err);
					toastState.addToast({
						type: "error",
						message: `Download failed: ${err}`,
						timeout: 5000
					});
				} finally {
					downloadInProgress = false;
				}
				break;

			case "Add to Collection":
				// TODO: Open collection picker modal
				toastState.addToast({
					type: "info",
					message: `Add ${items.length} image(s) to collection - Not yet implemented`,
					timeout: 3000
				});
				break;

			case "Share":
				// TODO: Open share dialog
				toastState.addToast({
					type: "info",
					message: `Share ${items.length} image(s) - Not yet implemented`,
					timeout: 3000
				});
				break;

			case "Copy Link":
				if (items.length === 1) {
					const url = getFullImagePath(items[0].image_paths?.original);
					copyToClipboard(url);
					toastState.addToast({
						type: "success",
						message: "Link copied to clipboard",
						timeout: 3000
					});
				} else {
					toastState.addToast({
						type: "warning",
						message: "Can only copy link for a single image",
						timeout: 3000
					});
				}
				break;

			case "Edit Metadata":
				// TODO: Open metadata editor
				toastState.addToast({
					type: "info",
					message: `Edit metadata for ${items.length} image(s) - Not yet implemented`,
					timeout: 3000
				});
				break;

			case "Move to Trash":
				const okTrash = confirm(
					`Move ${items.length} selected image(s) to trash?`
				);
				if (!okTrash) {
					return;
				}

				try {
					const res = await deleteImagesBulk({
						uids: items.map((i) => i.uid),
						force: false
					});
					if (res.status === 200 || res.status === 207) {
						const deletedUIDs = (res.data.results ?? [])
							.filter((r) => r.deleted)
							.map((r) => r.uid);
						images = images.filter((img) => !deletedUIDs.includes(img.uid));
						selectedAssets.clear();
					} else {
						toastState.addToast({
							type: "error",
							message: (res as any).data?.error ?? "Failed to delete images",
							timeout: 4000
						});
					}
				} catch (err) {
					toastState.addToast({
						type: "error",
						message: `Delete failed: ${err}`,
						timeout: 5000
					});
				}
				break;

			case "Force Delete":
				const okForce = confirm(
					`Permanently delete ${items.length} image(s)? This action cannot be undone!`
				);
				if (!okForce) {
					return;
				}

				try {
					const res = await deleteImagesBulk({
						uids: items.map((i) => i.uid),
						force: true
					});
					if (res.status === 200 || res.status === 207) {
						const deletedUIDs = (res.data.results ?? [])
							.filter((r) => r.deleted)
							.map((r) => r.uid);
						images = images.filter((img) => !deletedUIDs.includes(img.uid));
						selectedAssets.clear();
					} else {
						toastState.addToast({
							type: "error",
							message: (res as any).data?.error ?? "Failed to delete images",
							timeout: 4000
						});
					}
				} catch (err) {
					toastState.addToast({
						type: "error",
						message: `Delete failed: ${err}`,
						timeout: 5000
					});
				}
				break;
		}
	}

	async function paginate() {
		if (isPaginating || !hasMore) {
			return;
		}

		isPaginating = true;
		const nextPage = pagination.page + 1;
		const res = await listImages({ limit: pagination.limit, page: nextPage });

		if (res.status === 200) {
			const nextItems = res.data.items?.map((i) => i.image) ?? [];
			images.push(...nextItems);

			// Update pagination state from response
			pagination.page = res.data.page ?? nextPage;
			totalCount = res.data.count ?? totalCount;
			hasMore = !!res.data.next;
		} else {
			// On error, avoid tight loops; allow retry on next scroll
			console.error("paginate: request failed", res);
			hasMore = false;
		}

		isPaginating = false;
	}

	// Helper to perform token-based bulk download given a list of UIDs.
	// The server will create a download token and use it for authentication.
	async function performImageDownloads(uids: string[]) {
		if (!uids || uids.length === 0) {
			toastState.addToast({
				type: "warning",
				message: "No images selected for download",
				timeout: 3000
			});
			return;
		}

		try {
			if (uids.length === 1) {
				const uid = uids[0];
				const img = images.find((i) => i.uid === uid)!;

				return downloadOriginalImageFile(img);
			}

			const signRes = await signDownload({
				uids,
				expires_in: 300,
				allow_download: true,
				allow_embed: false,
				show_metadata: true
			});

			if (signRes.status !== 200) {
				toastState.addToast({
					type: "error",
					message: signRes.data?.error ?? "Failed to create download token",
					timeout: 4000
				});
				return;
			}

			const token = signRes.data.uid; // The token is stored in the uid field
			const dlRes = await downloadImagesZipBlob(token, { uids });

			if (dlRes.status !== 200) {
				throw new Error(dlRes.data?.error ?? "Failed to download archive");
			}

			// Extract filename from Content-Disposition header if available
			// Note: The custom function returns the blob directly, so we need to handle filename separately
			const filename = `images-${Date.now()}.zip`;

			const blob = dlRes.data;
			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);
		} catch (err) {
			console.error("Download error", err);
			toastState.addToast({
				type: "error",
				message: `Download failed: ${err}`,
				timeout: 5000
			});
		}
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

	// Drag and drop upload state
	let isDragging = $state(false);
	let dragCounter = $state(0);

	// Upload confirmation state
	let showUploadConfirm = $state(false);
	let uploadCandidates: File[] = $state([]);
	let suggestedCollectionName = $state("");

	// Collection creation state
	let showCollectionCreate = $state(false);
	let collectionCreateData = $state({
		name: "",
		description: "",
		private: false
	});
	let collectionCreatePending = $state(false);

	// Small drop-target state for 'Add to Collection' box
	let addBoxHover = $state(false);

	/**
	 * Recursively traverse file system entries to collect all files,
	 * including those in nested folders.
	 */
	async function traverseFileTree(item: FileSystemEntry): Promise<File[]> {
		const files: File[] = [];

		if (item.isFile) {
			const fileEntry = item as FileSystemFileEntry;
			const file = await new Promise<File>((resolve, reject) => {
				fileEntry.file(resolve, reject);
			});
			files.push(file);
		} else if (item.isDirectory) {
			const dirEntry = item as FileSystemDirectoryEntry;
			const reader = dirEntry.createReader();

			const entries = await new Promise<FileSystemEntry[]>(
				(resolve, reject) => {
					reader.readEntries(resolve, reject);
				}
			);

			for (const entry of entries) {
				const nestedFiles = await traverseFileTree(entry);
				files.push(...nestedFiles);
			}
		}

		return files;
	}

	/**
	 * Handle dropped files and folders.
	 * Supports single file, multiple files, and entire folders.
	 */
	async function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		dragCounter = 0;

		if (!e.dataTransfer) {
			return;
		}

		try {
			// Ignore internal image drops on the background - they must be dropped on the specific box
			// checking types is enough, getData works too but let's just skip if we see the key
			if (e.dataTransfer.types.includes("application/x-imagine-ids")) {
				return;
			}

			const allFiles: File[] = [];
			let detectedFolderName = "";

			// Use DataTransferItemList for folder support
			if (e.dataTransfer.items) {
				const items = Array.from(e.dataTransfer.items);

				// Note: Extract all entries synchronously FIRST before any async operations
				// DataTransferItem entries become invalid after the first async operation
				const entries: FileSystemEntry[] = [];
				for (const item of items) {
					if (item.kind === "file") {
						const entry = item.webkitGetAsEntry?.();
						if (entry) {
							entries.push(entry);
							if (entry.isDirectory && !detectedFolderName) {
								detectedFolderName = entry.name;
							}
						} else {
							// Fallback for browsers that don't support webkitGetAsEntry
							const file = item.getAsFile();
							if (file) {
								allFiles.push(file);
							}
						}
					}
				}

				// Now process all entries asynchronously
				for (const entry of entries) {
					const files = await traverseFileTree(entry);
					allFiles.push(...files);
				}
			} else {
				// Fallback to files list (doesn't support folders)
				const files = Array.from(e.dataTransfer.files);
				allFiles.push(...files);
			}

			if (allFiles.length === 0) {
				toastState.addToast({
					type: "info",
					message: "No files to upload",
					timeout: 3000
				});
				return;
			}

			// Filter valid files here to avoid processing invalid ones later
			const supportedExtensions = [
				...SUPPORTED_IMAGE_TYPES,
				...SUPPORTED_RAW_FILES
			];
			const validFiles = allFiles.filter((file) => {
				const ext = file.type.split("/")[1];
				return supportedExtensions.includes(ext as any);
			});

			if (validFiles.length === 0) {
				toastState.addToast({
					type: "error",
					message: "No supported image files found",
					timeout: 4000
				});
				return;
			}

			uploadCandidates = validFiles;
			suggestedCollectionName =
				detectedFolderName ||
				`New Collection ${new Date().toLocaleDateString()}`;

			// If we detected a folder or simply want to offer the choice always:
			showUploadConfirm = true;
			modal.show = true;
		} catch (err) {
			console.error("Drop upload error:", err);
			toastState.addToast({
				type: "error",
				message: `Upload failed: ${err}`,
				timeout: 5000
			});
		}
	}

	async function processUploads(files: File[]) {
		if (files.length < uploadCandidates.length) {
			toastState.addToast({
				type: "warning",
				message: `${uploadCandidates.length - files.length} file(s) skipped (unsupported format)`,
				timeout: 4000
			});
		}

		const manager = new UploadManager([
			...SUPPORTED_RAW_FILES,
			...SUPPORTED_IMAGE_TYPES
		] as SupportedImageTypes[]);
		const tasks = manager.addFiles(files);

		toastState.addToast({
			type: "success",
			message: `Starting upload of ${tasks.length} file(s)...`,
			timeout: 2500
		});

		// Start uploads with concurrency control
		await manager.start(tasks);

		// Wait for completion
		await waitForUploadCompletion(tasks);

		const uploadedImages = tasks
			.filter(
				(t) => t.state === UploadState.DONE || t.state === UploadState.DUPLICATE
			)
			.map((t) => t.imageData)
			.filter((img): img is ImageUploadSuccess => !!img);

		if (uploadedImages.length > 0) {
			toastState.addToast({
				type: "success",
				message: `Successfully uploaded ${uploadedImages.length} file(s)`,
				timeout: 3000
			});

			try {
				await invalidateAll();
			} catch (err) {
				console.error("Failed to fetch uploaded images:", err);
			}
		}

		return uploadedImages;
	}

	/**
	 * Handle drop specifically onto the "Add to Collection" box.
	 * This will upload any dropped files and create a new collection containing
	 * the resulting uploaded images.
	 */
	async function handleDropCreateCollection(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDragging = false;
		dragCounter = 0;

		if (!e.dataTransfer) return;

		try {
			// Check for internal drag of images first
			const dt = e.dataTransfer;
			const json = dt.getData("application/x-imagine-ids");
			if (json) {
				try {
					const uids: string[] = JSON.parse(json);
					if (uids.length === 0) {
						toastState.addToast({
							type: "info",
							message: "No images to add to collection",
							timeout: 3000
						});
						return;
					}

					const createRes = await createCollection({
						name: `New collection ${new Date().toLocaleString()}`,
						description: "Created from dropped images",
						private: false
					});

					if (createRes.status !== 201) {
						toastState.addToast({
							type: "error",
							message: `Failed to create collection (${createRes.status})`,
							timeout: 4000
						});
						return;
					}

					const collectionUid = createRes.data.uid;
					const addRes = await addCollectionImages(collectionUid, { uids });
					if (addRes.status === 200) {
						toastState.addToast({
							type: "success",
							message: `Collection created with ${uids.length} image(s)`,
							timeout: 4000
						});
						await invalidateAll();
						goto(`/collections/${collectionUid}`);
						return;
					} else {
						toastState.addToast({
							type: "warning",
							message: `Collection created but failed to add images (${addRes.status})`,
							timeout: 4000
						});
						return;
					}
				} catch (err) {
					console.warn("Failed to parse dragged image UIDs", err);
					return;
				}
			}

			const allFiles: File[] = [];

			if (e.dataTransfer.items) {
				const items = Array.from(e.dataTransfer.items);
				const entries: FileSystemEntry[] = [];
				for (const item of items) {
					if (item.kind === "file") {
						const entry = item.webkitGetAsEntry?.();
						if (entry) entries.push(entry);
						else {
							const file = item.getAsFile();
							if (file) allFiles.push(file);
						}
					}
				}

				for (const entry of entries) {
					const files = await traverseFileTree(entry);
					allFiles.push(...files);
				}
			} else {
				allFiles.push(...Array.from(e.dataTransfer.files));
			}

			if (allFiles.length === 0) {
				toastState.addToast({
					type: "info",
					message: "No files to add to collection",
					timeout: 3000
				});
				return;
			}

			const supportedExtensions = [
				...SUPPORTED_IMAGE_TYPES,
				...SUPPORTED_RAW_FILES
			];
			const validFiles = allFiles.filter((file) => {
				const ext = file.type.split("/")[1];
				return supportedExtensions.includes(ext as any);
			});

			if (validFiles.length === 0) {
				toastState.addToast({
					type: "error",
					message: "No supported image files found to add to collection",
					timeout: 4000
				});
				return;
			}

			const manager = new UploadManager([
				...SUPPORTED_RAW_FILES,
				...SUPPORTED_IMAGE_TYPES
			] as SupportedImageTypes[]);
			const tasks = manager.addFiles(validFiles);

			toastState.addToast({
				type: "success",
				message: `Uploading ${tasks.length} file(s) to create collection...`,
				timeout: 2500
			});

			// Start uploads with concurrency control
			await manager.start(tasks);

			// Wait for completion
			await waitForUploadCompletion(tasks);

			const uploadedImages = tasks
				.filter(
					(t) =>
						t.state === UploadState.DONE || t.state === UploadState.DUPLICATE
				)
				.map((t) => t.imageData)
				.filter((img): img is ImageUploadSuccess => !!img);

			if (uploadedImages.length > 0) {
				toastState.addToast({
					type: "success",
					message: `Successfully uploaded ${uploadedImages.length} file(s)`,
					timeout: 3000
				});

				try {
					await invalidateAll();
				} catch (err) {
					console.error("Failed to fetch uploaded images:", err);
				}
			}

			if (!uploadedImages || uploadedImages.length === 0) {
				toastState.addToast({
					type: "error",
					message: "Upload failed, no images available to add to collection",
					timeout: 4000
				});
				return;
			}

			// Create collection
			const createRes = await createCollection({
				name: `New collection ${new Date().toLocaleString()}`,
				description: "Created from dropped images",
				private: false
			});

			if (createRes.status !== 201) {
				toastState.addToast({
					type: "error",
					message: `Failed to create collection (${createRes.status})`,
					timeout: 4000
				});
				return;
			}

			const collectionUid = createRes.data.uid;
			const uids = uploadedImages.map((i: any) => i.uid).filter(Boolean);

			if (uids.length > 0) {
				const addRes = await addCollectionImages(collectionUid, { uids });
				if (addRes.status === 200) {
					toastState.addToast({
						type: "success",
						message: `Collection created with ${uids.length} image(s)`,
						timeout: 4000
					});
					goto(`/collections/${collectionUid}`);
				} else {
					toastState.addToast({
						type: "warning",
						message: `Collection created but failed to add images (${addRes.status})`,
						timeout: 4000
					});
				}
			} else {
				toastState.addToast({
					type: "warning",
					message: "Collection created but no uploaded image UIDs found",
					timeout: 4000
				});
			}
		} catch (err) {
			console.error("Add-to-collection drop error:", err);
			toastState.addToast({
				type: "error",
				message: `Failed to create collection from dropped images: ${err}`,
				timeout: 5000
			});
		}
	}

	/**
	 * Create a collection from the currently selected images (keyboard/click path).
	 */
	async function createCollectionFromSelected() {
		const items = Array.from(selectedAssets);
		if (!items || items.length === 0) {
			toastState.addToast({
				type: "info",
				message: "Select images first, or drag files here to upload",
				timeout: 3000
			});
			return;
		}

		try {
			const uids = items.map((i) => i.uid);
			const createRes = await createCollection({
				name: `New collection ${new Date().toLocaleString()}`,
				description: "Created from selected images",
				private: false
			});
			if (createRes.status !== 201) {
				toastState.addToast({
					type: "error",
					message: `Failed to create collection (${createRes.status})`,
					timeout: 4000
				});
				return;
			}

			const collectionUid = createRes.data.uid;
			const addRes = await addCollectionImages(collectionUid, { uids });
			if (addRes.status === 200) {
				toastState.addToast({
					type: "success",
					message: `Collection created with ${uids.length} image(s)`,
					timeout: 4000
				});
				await invalidateAll();
				goto(`/collections/${collectionUid}`);
			} else {
				toastState.addToast({
					type: "warning",
					message: `Collection created but failed to add images (${addRes.status})`,
					timeout: 4000
				});
			}
		} catch (err) {
			console.error("createCollectionFromSelected error", err);
			toastState.addToast({
				type: "error",
				message: `Failed to create collection: ${err}`,
				timeout: 5000
			});
		}
	}

	let pendingNewRaw: ImageUploadSuccess[] = [];
	let addImagesDebounceTimer: number | undefined;
	const ADD_IMAGES_DEBOUNCE_MS = 550;

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

				throw new Error(res.data.error);
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
				images.push(...imagesToAdd);
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

	function handleDragEnter(e: DragEvent) {
		e.preventDefault();
		dragCounter++;
		if (dragCounter === 1) {
			isDragging = true;
		}
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		dragCounter--;
		if (dragCounter === 0) {
			isDragging = false;
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = "copy";
		}
	}

	hotkeys("escape", (e) => {
		e.preventDefault();
		selectedAssets.clear();

		singleSelectedAsset = undefined;
		lightboxImage = undefined;

		if (modal.show) {
			modal.show = false;
			showUploadConfirm = false;
			showCollectionCreate = false;
		}
	});

	function handleConfirmUploadOnly() {
		showUploadConfirm = false;
		modal.show = false;
		processUploads(uploadCandidates);
		uploadCandidates = [];
	}

	function handleConfirmUploadCollection() {
		showUploadConfirm = false;
		// Keep modal open, switch to collection create
		collectionCreateData = {
			name: suggestedCollectionName,
			description: "",
			private: false
		};
		showCollectionCreate = true;
	}

	async function handleCollectionSubmit() {
		collectionCreatePending = true;
		try {
			// 1. Create Collection
			const createRes = await createCollection(collectionCreateData);
			if (createRes.status !== 201) {
				toastState.addToast({
					type: "error",
					message: `Failed to create collection (${createRes.status})`,
					timeout: 4000
				});
				collectionCreatePending = false;
				return;
			}

			const collectionUid = createRes.data.uid;

			// 2. Upload Images
			showCollectionCreate = false;
			modal.show = false;

			const uploadedImages = await processUploads(uploadCandidates);

			// 3. Add to Collection
			const uids = uploadedImages
				.filter((img) => img && img.uid)
				.map((i: any) => i.uid);

			if (uids.length > 0) {
				const addRes = await addCollectionImages(collectionUid, { uids });
				if (addRes.status === 200) {
					toastState.addToast({
						type: "success",
						message: `Added ${uids.length} images to collection **${collectionCreateData.name}**`,
						timeout: 4000
					});
					await invalidateAll();
					goto(`/collections/${collectionUid}`);
				} else {
					toastState.addToast({
						type: "warning",
						message: `Images uploaded but failed to add to collection: **${addRes.status}**`,
						timeout: 4000
					});
				}
			}
		} catch (err) {
			console.error("Collection/Upload flow failed", err);
			toastState.addToast({
				type: "error",
				message: `Operation failed: ${err}`,
				timeout: 5000
			});
		} finally {
			collectionCreatePending = false;
			uploadCandidates = [];
		}
	}
</script>

<svelte:body
	ondragenter={handleDragEnter}
	ondragleave={handleDragLeave}
	ondragover={handleDragOver}
	ondrop={handleDrop}
/>

<svelte:head>
	<title>Photos</title>
</svelte:head>

{#if downloadInProgress}
	<div class="download-spinner" aria-live="polite" title="Download in progress">
		<LoadingContainer />
	</div>
{/if}

{#if showUploadConfirm && modal.show}
	<ConfirmationModal title="Upload Options">
		<p>You dropped {uploadCandidates.length} file(s).</p>
		<p>How would you like to upload them?</p>

		{#snippet actions()}
			<Button onclick={handleConfirmUploadOnly}>Upload Individually</Button>
			<Button
				onclick={handleConfirmUploadCollection}
				style="background-color: var(--imag-primary); color: white;"
			>
				Create Collection & Upload
			</Button>
		{/snippet}
	</ConfirmationModal>
{/if}

{#if showCollectionCreate && modal.show}
	<CollectionModal
		heading="Create Collection"
		bind:data={collectionCreateData}
		buttonText={collectionCreatePending ? "Creating..." : "Create & Upload"}
		modalAction={handleCollectionSubmit}
	/>
{/if}

{#if isDragging}
	<div class="drop-overlay" transition:fade={{ duration: 150 }}>
		<div class="drop-overlay-content">
			<MaterialIcon
				iconName="upload"
				style="font-size: 4rem; margin-bottom: 1rem; color: var(--imag-10-dark);"
			/>
			<p style="font-size: 1.5rem; font-weight: 600;">Drop files to upload</p>
			<p style="font-size: 1rem; opacity: 0.8;">Supports images and folders</p>

			<!-- Small Add to Collection drop box placed below the main content -->
			<div
				class="add-to-collection-box"
				class:hover={addBoxHover}
				role="button"
				tabindex="0"
				aria-label="Add to Collection — drop images here or press Enter to create from selected images"
				onclick={async () => {
					// Keyboard/click activation: create collection from selected images (if any)
					await createCollectionFromSelected();
				}}
				onkeydown={async (e: KeyboardEvent) => {
					if (e.key === "Enter" || e.key === " ") {
						e.preventDefault();
						await createCollectionFromSelected();
					}
				}}
				ondragenter={(e) => {
					e.preventDefault();
					addBoxHover = true;
				}}
				ondragleave={(e) => {
					e.preventDefault();
					addBoxHover = false;
				}}
				ondragover={(e) => {
					e.preventDefault();
					if (e.dataTransfer) {
						e.dataTransfer.dropEffect = "copy";
					}

					addBoxHover = true;
				}}
				ondrop={async (e) => {
					addBoxHover = false;
					await handleDropCreateCollection(e);
				}}
			>
				<MaterialIcon
					iconName="collections_bookmark"
					style="font-size: 1.6rem; margin-bottom: 0.25rem; color: var(--imag-10-dark);"
				/>
				<span>Add to Collection</span>
			</div>
		</div>
	</div>
{/if}

{#if lightboxImage}
	<ImageLightbox bind:lightboxImage {prevLightboxImage} {nextLightboxImage} />
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
	bind:data={images}
	{hasMore}
	paginate={() => paginate()}
>
	{#if images.length > 0}
		{#if selectedAssets.size > 1}
			<AssetToolbar class="selection-toolbar" stickyToolbar={true}>
				<button
					class="toolbar-button"
					title="Clear selection"
					aria-label="Clear selection"
					style="margin-right: 1em;"
					onclick={() => selectedAssets.clear()}
				>
					<MaterialIcon iconName="close" />
				</button>
				<span style="font-weight: 600;">{selectedAssets.size} selected</span>
				<div
					style="margin-left: auto; display: flex; gap: 0.5rem; align-items: center;"
				>
					<Dropdown
						class="toolbar-button"
						icon="more_horiz"
						options={actionMenuOptions}
						showSelectionIndicator={false}
						onSelect={handleActionMenu}
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
							const { asset, anchor } = detail;
							if (!selectedAssets.has(asset) || selectedAssets.size <= 1) {
								singleSelectedAsset = asset;
								selectedAssets.clear();
								selectedAssets.add(asset);
							}

							ctxItems = actionMenuOptions.map((opt) => ({
								id: opt.title,
								label: opt.title,
								icon: opt.icon,
								action: () => {
									if (opt.title === "Download") {
										const selected = Array.from(selectedAssets).map(
											(i) => i.uid
										);
										performImageDownloads(selected);
									} else {
										handleActionMenu(opt);
									}
								}
							}));
							ctxAnchor = anchor;
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

	.download-spinner {
		position: fixed;
		top: 1rem;
		right: 1rem;
		z-index: 2000;
		width: 2.5rem;
		height: 2.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(0, 0, 0, 0.6);
		border-radius: 0.5rem;
		padding: 0.25rem;
	}

	#add_to_imagine-container {
		display: flex;
		flex-direction: column;
		justify-content: left;
	}

	.drop-overlay {
		position: fixed;
		inset: 0;
		z-index: 1000;
		color: var(--imag-10-dark);
		background: rgba(0, 0, 0, 0.85);
		backdrop-filter: blur(8px);
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}

	.drop-overlay-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: var(--imag-10-dark);
		pointer-events: auto;
		border: 2px solid var(--imag-primary);
		border-radius: 1rem;
		padding: 3rem 4rem;
		background: rgba(0, 0, 0, 0.5);
	}

	#viz-no_assets {
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.add-to-collection-box {
		pointer-events: auto;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 12rem;
		height: 4.25rem;
		background-color: var(--imag-10-light);
		border: 2px solid var(--imag-primary);
		color: var(--imag-10-dark);
		border-radius: 0.75rem;
		gap: 0.25rem;
		font-weight: 600;
		margin-top: 1rem;
		padding: 0.75rem 1rem;

		&:focus {
			outline: 3px solid rgba(0, 0, 0, 0.35);
			outline-offset: 2px;
		}

		&:hover {
			background: var(--imag-20-light);
		}
	}
</style>
