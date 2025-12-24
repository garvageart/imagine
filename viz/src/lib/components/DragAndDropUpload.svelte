<script lang="ts">
	import { fade } from "svelte/transition";
	import MaterialIcon from "./MaterialIcon.svelte";
	import { modal } from "$lib/states/index.svelte";
	import CollectionModal from "./modals/CollectionModal.svelte";
	import { traverseFileTree } from "$lib/utils/files";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import {
		SUPPORTED_IMAGE_TYPES,
		SUPPORTED_RAW_FILES,
		type SupportedImageTypes
	} from "$lib/types/images";
	import ConfirmationModal from "./modals/ConfirmationModal.svelte";
	import Button from "./Button.svelte";
	import {
		waitForUploadCompletion,
		type ImageUploadSuccess
	} from "$lib/upload/manager.svelte";
	import { invalidateAll, goto } from "$app/navigation";
	import { createCollection, addCollectionImages, type Image } from "$lib/api";
	import { SelectionScope } from "$lib/states/selection.svelte";
	import { UploadState } from "$lib/upload/asset.svelte";
	import UploadManager from "$lib/upload/manager.svelte";

	interface Props {
		scopeId: string; // might be useful soon
		selectionScope: SelectionScope<Image>;
		showCollectionCreateBox?: boolean;
		createCollectionFromSelected?: () => Promise<void>;
	}

	let { showCollectionCreateBox, scopeId, selectionScope }: Props = $props();

	// Drag and drop upload state
	let isDragging = $state(false);
	let dragCounter = $state(0);

	// Upload confirmation state
	let showUploadConfirm = $state(false);
	let uploadCandidates: File[] = $state([]);
	let suggestedCollectionName = $state("");

	let showCollectionCreate = $state(false);
	let collectionCreateData = $state({
		name: "",
		description: "",
		private: false
	});
	let collectionCreatePending = $state(false);

	// Small drop-target state for 'Add to Collection' box
	let addBoxHover = $state(false);

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

	/**
	 * Create a collection from the currently selected images (keyboard/click path).
	 */
	async function createCollectionFromSelected() {
		const items = Array.from(selectionScope.selected);
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
</script>

<svelte:body
	ondragenter={handleDragEnter}
	ondragleave={handleDragLeave}
	ondragover={handleDragOver}
	ondrop={handleDrop}
/>

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

<div class="drop-overlay" transition:fade={{ duration: 150 }}>
	<div class="drop-overlay-content">
		<MaterialIcon
			iconName="upload"
			style="font-size: 4rem; margin-bottom: 1rem; color: var(--imag-10-dark);"
		/>
		<p style="font-size: 1.5rem; font-weight: 600;">Drop files to upload</p>
		<p style="font-size: 1rem; opacity: 0.8;">Supports images and folders</p>

		{#if showCollectionCreateBox}
			<!-- Small Add to Collection drop box placed below the main content -->
			<div
				class="add-to-collection-box"
				class:hover={addBoxHover}
				role="button"
				tabindex="0"
				aria-label="Add to Collection — drop images here or press Enter to create from selected images"
				onclick={async () => {
					// Keyboard/click activation: create collection from selected images (if any)
					await createCollectionFromSelected?.();
				}}
				onkeydown={async (e: KeyboardEvent) => {
					if (e.key === "Enter" || e.key === " ") {
						e.preventDefault();
						await createCollectionFromSelected?.();
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
					await handleDropCreateCollection?.(e);
				}}
			>
				<MaterialIcon
					iconName="collections_bookmark"
					style="font-size: 1.6rem; margin-bottom: 0.25rem; color: var(--imag-10-dark);"
				/>
				<span>Add to Collection</span>
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
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
