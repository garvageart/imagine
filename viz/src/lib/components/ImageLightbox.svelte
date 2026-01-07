<script lang="ts">
	import { loadImage } from "$lib/utils/dom";
	import { fade } from "svelte/transition";
	import Lightbox from "./Lightbox.svelte";
	import LoadingContainer from "./LoadingContainer.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import {
		getFullImagePath,
		updateImage,
		type Image,
		type ImageUpdate
	} from "$lib/api";
	import hotkeys from "hotkeys-js";
	import {
		formatBytes,
		getFlashMode,
		getImageLabel,
		getTakenAt,
		getThumbhashURL
	} from "$lib/utils/images";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import IconButton from "./IconButton.svelte";
	import { downloadOriginalImageFile } from "$lib/utils/http";
	import { ZoomPanCrop } from "$lib/images/zoom/crop";
	import { setRating } from "$lib/images/exif";
	import InputText from "./dom/InputText.svelte";
	import LabelSelector from "./LabelSelector.svelte";
	import { LabelColours, type ImageLabel } from "$lib/images/constants";
	import CropOverlay from "./CropOverlay.svelte";
	import CropTools from "./CropTools.svelte";

	interface Props {
		lightboxImage: Image | undefined;
		prevLightboxImage?: () => void;
		nextLightboxImage?: () => void;
		onImageUpdated?: (image: Image) => void;
	}

	let {
		lightboxImage = $bindable(),
		prevLightboxImage,
		nextLightboxImage,
		onImageUpdated
	}: Props = $props();

	let show = $derived(lightboxImage !== undefined);
	let imageToLoad = $derived(
		getFullImagePath(
			lightboxImage!.image_paths?.preview ||
				lightboxImage!.image_paths?.original
		)
	);

	// Crop State
	let isCropping = $state(false);
	let cropAspectRatio = $state<number | null>(null);
	let currentCrop = $state<{
		x: number;
		y: number;
		width: number;
		height: number;
	} | null>(null);
	let cropMenuPosition = $state<{ x: number; y: number } | null>(null);
	// Store crop edits (original/natural coordinates) to restore them when re-entering crop mode
	let cropEdits = $state<
		Record<string, { x: number; y: number; width: number; height: number }>
	>({});

	let overriddenImages = $state<Record<string, string>>({});
	let displayURL = $derived(
		lightboxImage?.uid && overriddenImages[lightboxImage.uid] && !isCropping
			? overriddenImages[lightboxImage.uid]
			: imageToLoad
	);

	let direction = $state<"left" | "right">("right");
	let showMetadata = $state(true);
	let editNameMode = $state(false);

	let imageEl: HTMLImageElement = $state()!;
	let canvasEl: HTMLCanvasElement = $state()!;
	let imageContainerEl: HTMLDivElement = $state()!;
	let zoomTargetEl: HTMLDivElement = $state()!; // Wrapper for image + overlay
	let zoomer = $state<ZoomPanCrop>();

	// Transform State for CropOverlay (Scale only needed for UI inverse scaling)
	let transformState = $state({
		scale: 1
	});

	// Helper to get render dimensions
	let imageDimensions = $state<{ width: number; height: number } | null>(null);

	// Effect 1: Lifecycle - Initialize Zoomer when DOM elements are ready
	$effect(() => {
		if (imageContainerEl && zoomTargetEl) {
			// Initialize ZoomPan on the wrapper
			const newZoomer = new ZoomPanCrop(imageContainerEl, zoomTargetEl);

			// Listen for transform changes
			newZoomer.onTransformChange((t) => {
				transformState = { scale: t.scale };
			});

			zoomer = newZoomer;

			return () => {
				newZoomer.destroy();
			};
		}
	});

	// Effect 2: Restoration - Apply crop when in crop mode and everything is ready
	$effect(() => {
		if (isCropping && imageEl && imageEl.complete && zoomer) {
			restoreCrop(zoomer);
		}
	});

	function goToPrev() {
		if (isCropping) return;
		direction = "left";
		prevLightboxImage?.();
	}

	function goToNext() {
		if (isCropping) return;
		direction = "right";
		nextLightboxImage?.();
	}

	function restoreCrop(targetZoomer?: ZoomPanCrop) {
		const z = targetZoomer || zoomer;
		if (!z || !imageEl || !lightboxImage) return;

		// Ensure we capture dimensions for overlay
		if (imageEl.clientWidth > 0 && imageEl.clientHeight > 0) {
			imageDimensions = {
				width: imageEl.clientWidth,
				height: imageEl.clientHeight
			};
		}

		const saved = cropEdits[lightboxImage.uid];
		let initialCrop = null;

		// If we have a saved crop (in original coordinates), scale it to current render dimensions
		if (saved && lightboxImage.width && lightboxImage.height) {
			const scaleX = imageEl.clientWidth / lightboxImage.width;
			const scaleY = imageEl.clientHeight / lightboxImage.height;

			initialCrop = {
				x: saved.x * scaleX,
				y: saved.y * scaleY,
				width: saved.width * scaleX,
				height: saved.height * scaleY
			};
		} else if (saved && imageEl.naturalWidth > 0) {
			// Fallback to naturalWidth if original dimensions missing (shouldn't happen often)
			const scaleX = imageEl.clientWidth / imageEl.naturalWidth;
			const scaleY = imageEl.clientHeight / imageEl.naturalHeight;

			initialCrop = {
				x: saved.x * scaleX,
				y: saved.y * scaleY,
				width: saved.width * scaleX,
				height: saved.height * scaleY
			};
		}

		z.initCrop(initialCrop);

		z.onCropChange((c) => {
			currentCrop = c;
		});
		currentCrop = z.getCrop();
	}

	function toggleCropMode() {
		if (!imageEl) return;

		if (!isCropping) {
			// Enter crop mode
			isCropping = true;
			// Hide metadata panel to give space if it was open
			// We don't toggle showMetadata state here to preserve user preference when exiting,
			// but the template logic will hide it.

			// Reset zoom to ensure user sees the whole image context for cropping
			zoomer?.reset();

			// Note: switching isCropping will trigger a re-render of the image wrapper (key block)
			// So restoration happens in the effect/onload, not here immediately.
		} else {
			// Exit crop mode (cancel)
			isCropping = false;
			currentCrop = null;
			cropMenuPosition = null;
		}
	}

	function handleCropApply() {
		if (!currentCrop || !imageEl || !lightboxImage) {
			console.warn("Missing requirements for crop apply", {
				currentCrop,
				imageEl,
				lightboxImage,
				canvasEl
			});
			return;
		}

		// Calculate crop relative to the ORIGINAL image (for backend/storage)
		// We use lightboxImage dimensions if available, otherwise fall back to rendered.
		const originalWidth = lightboxImage.width || imageEl.naturalWidth;
		const originalHeight = lightboxImage.height || imageEl.naturalHeight;

		const scaleToOriginalX = originalWidth / imageEl.clientWidth;
		const scaleToOriginalY = originalHeight / imageEl.clientHeight;

		const originalCrop = {
			x: Math.round(currentCrop.x * scaleToOriginalX),
			y: Math.round(currentCrop.y * scaleToOriginalY),
			width: Math.round(currentCrop.width * scaleToOriginalX),
			height: Math.round(currentCrop.height * scaleToOriginalY)
		};

		console.log("Applying Crop (Original):", originalCrop);

		// Save the crop to state so we can restore it later
		cropEdits[lightboxImage.uid] = originalCrop;

		// Calculate crop relative to the CURRENTLY LOADED image (preview) for client-side visual feedback
		// imageEl.naturalWidth is the width of the loaded source (preview), NOT necessarily the original.
		const scaleToPreviewX = imageEl.naturalWidth / imageEl.clientWidth;
		const scaleToPreviewY = imageEl.naturalHeight / imageEl.clientHeight;

		const previewCrop = {
			x: Math.round(currentCrop.x * scaleToPreviewX),
			y: Math.round(currentCrop.y * scaleToPreviewY),
			width: Math.round(currentCrop.width * scaleToPreviewX),
			height: Math.round(currentCrop.height * scaleToPreviewY)
		};

		// Client-side apply using previewCrop
		let canvas = canvasEl;
		if (!canvas) {
			canvas = document.createElement("canvas");
		}

		canvas.width = previewCrop.width;
		canvas.height = previewCrop.height;
		const ctx = canvas.getContext("2d");
		if (ctx) {
			ctx.drawImage(
				imageEl,
				previewCrop.x,
				previewCrop.y,
				previewCrop.width,
				previewCrop.height,
				0,
				0,
				previewCrop.width,
				previewCrop.height
			);
			const dataURL = canvas.toDataURL("image/jpeg", 0.9);
			overriddenImages[lightboxImage.uid] = dataURL;
		} else {
			console.error("Failed to get 2D context");
		}

		toastState.addToast({
			type: "success",
			title: "Crop Applied (Client-side)",
			message: `Image updated in view.`,
			timeout: 4000
		});

		// Exit mode
		isCropping = false;
		currentCrop = null;
		cropMenuPosition = null;
		zoomer?.reset(); // Reset zoom to see the new image full frame
	}

	function handleCropReset() {
		if (!zoomer || !lightboxImage) return;

		// Clear saved crop state
		delete cropEdits[lightboxImage.uid];

		// Reset zoomer crop to full image
		zoomer.initCrop();

		// Update current crop state
		currentCrop = zoomer.getCrop();
	}

	function handleContextMenu(e: MouseEvent) {
		if (isCropping) {
			e.preventDefault();
			cropMenuPosition = { x: e.clientX, y: e.clientY };
		} else {
			e.preventDefault();
		}
	}

	let thumbhashURL = $derived(
		lightboxImage ? getThumbhashURL(lightboxImage) : undefined
	);

	let imageUid = $derived(lightboxImage?.uid);
	let currentImageEl: HTMLImageElement | undefined = $derived(
		imageUid ? document.createElement("img") : undefined
	);

	// Rating UI state: previewRating for hover preview, rating is the set value
	let previewRating = $state<number | null>(null);
	let rating = $state<number | null>(
		lightboxImage?.image_metadata?.rating ?? null
	);

	// Star values moved to state to avoid rebuilding the array on each render
	let starValues = $state<number[]>([1, 2, 3, 4, 5]);

	// Prevent concurrent rating updates
	let updatingRating = $state(false);

	async function setImageRating(newRating: number | null) {
		if (!lightboxImage) {
			return;
		}

		if (updatingRating) {
			return;
		}

		updatingRating = true;
		const prev = rating;
		rating = newRating;

		try {
			const newSuccessfulRating = await setRating(
				lightboxImage,
				prev,
				newRating
			);
			rating = newSuccessfulRating;
		} catch (err: any) {
			const ratingErr = err as Error;
			toastState.addToast({
				type: "error",
				title: "Failed to update rating",
				message: `An error occurred while updating the image rating: ${ratingErr.message}`
			});
			rating = prev;
		} finally {
			updatingRating = false;
		}
	}

	hotkeys("left,right", (e, handler) => {
		if (!show || isCropping) {
			return;
		}

		e.preventDefault();
		if (handler.key === "left") {
			goToPrev();
		} else if (handler.key === "right") {
			goToNext();
		}
	});

	hotkeys("enter,esc", (e, handler) => {
		if (!show || !isCropping) {
			return;
		}

		e.preventDefault();
		if (handler.key === "enter") {
			handleCropApply();
		} else if (handler.key === "esc") {
			toggleCropMode();
		}
	});

	function formatFileSize() {
		const size = lightboxImage?.image_metadata?.file_size;
		return formatBytes(size) ?? "—";
	}

	const lightboxMaterialIconColour =
		"color: var(--imag-10-dark); fill: var(--imag-10-dark);";
</script>

{#snippet metadataEditor()}
	<div class="metadata-editor">
		<div class="metadata-header">
			<h3>Metadata</h3>
		</div>
		<div class="metadata-exif-box">
			<div class="exif-cards">
				<div class="exif-card">
					<div class="card-row main-row">
						<MaterialIcon iconName="image" class="exif-material-icon" />
						<div class="card-values">
							{#if editNameMode}
								<InputText
									bind:value={lightboxImage!.name}
									class="value-big"
									style="min-height: auto; padding: 0.5rem;"
									spellcheck="false"
									autofocus={true}
									title={lightboxImage?.name ?? "Edit Image Name"}
									onblur={async (e) => {
										editNameMode = false;
										if (e.currentTarget.value.trim() === lightboxImage!.name) {
											return;
										}

										try {
											const res = await updateImage(lightboxImage!.uid, {
												name: lightboxImage!.name
											});

											if (res.status === 200) {
												lightboxImage = res.data;
												onImageUpdated?.(res.data);
												toastState.addToast({
													type: "success",
													title: "Image name updated",
													message: `Updated to "${lightboxImage?.name}"`,
													timeout: 2000
												});
											} else {
												toastState.addToast({
													type: "error",
													title: "Failed to update image name",
													message: `Failed to update image name: ${res.data.error}`
												});
											}
										} catch (error) {
											toastState.addToast({
												type: "error",
												title: "Failed to update image name",
												message:
													(error as Error).message ??
													"An unknown error occurred."
											});
										}
									}}
									onkeydown={(e) => {
										if (e.key === "Enter") {
											e.currentTarget.blur();
										} else if (e.key === "Escape") {
											editNameMode = false;
										}
									}}
								/>
							{:else}
								<div
									role="textbox"
									tabindex="0"
									onclick={() => (editNameMode = true)}
									onkeydown={() => (editNameMode = true)}
									title={lightboxImage?.name}
									class="value-big"
								>
									{lightboxImage?.name}
								</div>
							{/if}
						</div>
					</div>

					<div class="card-row meta-row">
						<MaterialIcon
							iconName="calendar_today"
							class="exif-material-icon"
						/>
						<div class="card-values">
							<div class="value-big">
								{#if lightboxImage?.image_metadata?.file_created_at}
									{getTakenAt(lightboxImage).toLocaleDateString(undefined, {
										year: "numeric",
										month: "long",
										day: "numeric"
									})}
								{:else}
									Unknown Date
								{/if}
							</div>
							{#if lightboxImage?.image_metadata?.file_created_at}
								<div class="value-sub">
									{getTakenAt(lightboxImage).toLocaleTimeString(undefined, {
										hour: "2-digit",
										minute: "2-digit"
									})}
								</div>
							{/if}
						</div>
					</div>
				</div>
				<!-- Camera/Exposure card -->
				<div class="exif-card">
					<div class="card-row main-row">
						<div class="card-values">
							{#if lightboxImage?.exif?.model && lightboxImage?.exif?.make}
								<div class="value-big">
									{lightboxImage.exif.make}
									{lightboxImage.exif.model.replace(
										new RegExp(`^${lightboxImage.exif.make} `),
										""
									)}
								</div>
							{:else}
								<div class="value-big">Unknown Camera</div>
							{/if}
							{#if lightboxImage?.exif?.lens_model}
								<div class="value-sub">{lightboxImage.exif.lens_model}</div>
							{:else}
								<div class="value-sub">Unknown Lens Make</div>
							{/if}
							{#if lightboxImage?.exif?.focal_length}
								<div class="value-sub">{lightboxImage.exif.focal_length}</div>
							{:else}
								<div class="value-sub">Unknown Focal Length</div>
							{/if}
						</div>
					</div>
				</div>
				<div class="exif-card-group">
					<div class="exif-card">
						<div class="card-row main-row">
							<MaterialIcon iconName="camera" class="exif-material-icon" />
							<div class="card-values">
								<div class="value-sub">
									{lightboxImage?.exif?.f_number ??
										lightboxImage?.exif?.aperture ??
										"—"}
								</div>
								<div class="value-sub">
									{lightboxImage?.exif?.exposure_time ?? "—"}
								</div>
							</div>
						</div>
						<div class="card-row meta-row">
							<MaterialIcon iconName="tune" class="exif-material-icon" />
							<div class="card-values">
								<div class="value-sub">
									ISO {lightboxImage?.exif?.iso ?? "—"}
								</div>
								<div class="value-sub">
									{lightboxImage?.exif?.exposure_value ?? "—"}
								</div>
							</div>
						</div>
						<div class="card-row meta-row">
							<MaterialIcon
								iconName="flash_on"
								fill={true}
								style="color: #FFC107; fill: #FFC107;"
								class="exif-material-icon"
							/>
							<div class="card-values">
								<div class="value-sub">
									Flash {getFlashMode(lightboxImage?.exif?.flash) ??
										"—"}
								</div>
							</div>
						</div>
					</div>

					<div class="exif-card">
						<div class="card-row main-row">
							<div class="card-values">
								<div class="value-sub">
									{lightboxImage?.width} x {lightboxImage?.height}
								</div>
							</div>
						</div>
						<div class="card-row main-row">
							<MaterialIcon
								iconName="aspect_ratio"
								class="exif-material-icon"
							/>
							<div class="card-values">
								<div class="value-sub">
									{Math.floor(
										(lightboxImage?.width! * lightboxImage?.height!) / 1_000_000
									)} MP
								</div>
								<div class="value-sub">{formatFileSize()}</div>
							</div>
						</div>
						<div class="card-row meta-row">
							<MaterialIcon iconName="palette" class="exif-material-icon" />
							<div class="card-values">
								<div class="value-sub">
									{lightboxImage?.image_metadata?.color_space ?? "—"}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<div class="rating-container">
				<LabelSelector
					variant="compact"
					label={getImageLabel(lightboxImage!)}
					onSelect={async (selectedLabel) => {
						if (!lightboxImage) {
							return;
						}

						// Reverse lookup: find the key (Name) for the selected color (Value)
						const entry = Object.entries(LabelColours).find(
							([_, colour]) => colour === selectedLabel
						);
						const labelName = entry
							? entry[0] === "None"
								? null
								: entry[0]
							: null;
						// If "None" is selected, send null to clear the label
						const labelToSend = labelName as ImageLabel | null;

						try {
							const res = await updateImage(lightboxImage.uid, {
								image_metadata: {
									label: labelToSend
								}
							});

							if (res.status === 200) {
								lightboxImage = res.data;
								onImageUpdated?.(res.data);
							} else {
								toastState.addToast({
									type: "error",
									title: "Failed to update image label",
									message: `Failed to update image label: ${res.data.error}`
								});
							}
						} catch (error) {
							toastState.addToast({
								type: "error",
								title: "Failed to update image label",
								message:
									(error as Error).message ?? "An unknown error occurred."
							});
						}
					}}
				/>

				<div
					class="rating-stars"
					role="group"
					onmouseleave={() => (previewRating = null)}
				>
					{#each starValues as i}
						<button
							class="rating-button"
							title={`Set Rating: ${i}`}
							aria-label={`Set Rating: ${i}`}
							onmouseenter={() => (previewRating = i)}
							onmouseleave={() => (previewRating = null)}
							onclick={() => setImageRating(i)}
							disabled={updatingRating}
						>
							<MaterialIcon
								fill={i <= (previewRating ?? rating ?? 0)}
								iconName="star"
								iconStyle={"sharp"}
							/>
						</button>
					{/each}
					{#if rating !== null && rating !== 0}
						<button
							class="rating-clear"
							aria-label="Clear rating"
							onclick={() => setImageRating(0)}
							disabled={updatingRating}
						>
							<MaterialIcon iconName="close" weight={600} />
						</button>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/snippet}

<Lightbox
	bind:show
	backgroundOpacity={0.95}
	onclick={() => {
		if (!isCropping) {
			lightboxImage = undefined;
		} else {
			// Close crop menu if open and clicked outside
			cropMenuPosition = null;
		}
	}}
>
	<div class="image-lightbox-container">
		<div class="image-container">
			<IconButton
				id="lightbox-icon-close"
				class="lightbox-button-icon"
				hoverColor="transparent"
				title="Close"
				iconName="close"
				onclick={() => {
					if (isCropping) {
						isCropping = false;
						cropMenuPosition = null;
					} else {
						lightboxImage = undefined;
					}
				}}
			/>
			{#if !isCropping}
				<div class="image-icon-buttons">
					<IconButton
						class="lightbox-button-icon"
						hoverColor="transparent"
						style={lightboxMaterialIconColour}
						title="Crop"
						iconName="crop"
						onclick={toggleCropMode}
					/>
					<IconButton
						class="lightbox-button-icon"
						hoverColor="transparent"
						style={lightboxMaterialIconColour}
						title="Download"
						iconName="download"
						onclick={() => {
							downloadOriginalImageFile(lightboxImage!);
						}}
					/>
					<IconButton
						class="lightbox-button-icon"
						hoverColor="transparent"
						style={lightboxMaterialIconColour}
						title={`${showMetadata ? "Hide" : "Show"} Info`}
						onclick={(e) => {
							e.stopPropagation();
							showMetadata = !showMetadata;
						}}
						iconName="info"
					/>
				</div>
			{/if}
			{#key displayURL}
				<div
					class="image-wrapper"
					bind:this={imageContainerEl}
					role="presentation"
				>
					<div
						class="zoom-target"
						class:is-crop={isCropping}
						bind:this={zoomTargetEl}
						oncontextmenu={handleContextMenu}
						role="presentation"
					>
						{#await loadImage(displayURL, currentImageEl!)}
							{#if !thumbhashURL}
								<div style="width: 3em; height: 3em">
									<LoadingContainer />
								</div>
							{:else}
								<img
									src={thumbhashURL}
									in:fade
									out:fade
									class="lightbox-image lightbox-placeholder"
									alt="Placeholder image for {lightboxImage!.name}"
									aria-hidden="true"
									style={`aspect-ratio: ${lightboxImage!.width} / ${lightboxImage!.height};`}
								/>
							{/if}
						{:then url}
							<img
								bind:this={imageEl}
								src={url}
								class="lightbox-image"
								class:is-crop={isCropping}
								alt={lightboxImage!.name}
								title={lightboxImage!.name}
								loading="eager"
								crossorigin="use-credentials"
								data-image-id={lightboxImage!.uid}
								onload={() => {
									if (isCropping) restoreCrop();
								}}
								ondragstart={(e) => e.preventDefault()}
								oncontextmenu={handleContextMenu}
								in:fade
								out:fade
							/>
							{#if isCropping && imageDimensions && currentCrop && zoomer}
								<CropOverlay
									width={imageDimensions.width}
									height={imageDimensions.height}
									crop={currentCrop}
									{zoomer}
									scale={transformState.scale}
								/>
							{/if}
						{:catch error}
							<p>Failed to load image</p>
							<p>{error}</p>
						{/await}
					</div>
				</div>
			{/key}

			{#if prevLightboxImage && nextLightboxImage && !isCropping}
				<div class="lightbox-nav">
					<button
						class="lightbox-nav-btn prev lightbox-button-icon"
						aria-label="Previous image"
						onclick={(e) => {
							e.stopPropagation();
							goToPrev();
						}}
					>
						<MaterialIcon
							iconName="arrow_back"
							style={lightboxMaterialIconColour}
						/>
					</button>
					<button
						class="lightbox-nav-btn next lightbox-button-icon"
						aria-label="Next image"
						onclick={(e) => {
							e.stopPropagation();
							goToNext();
						}}
					>
						<MaterialIcon
							iconName="arrow_forward"
							style={lightboxMaterialIconColour}
						/>
					</button>
				</div>
			{/if}

			{#if isCropping && cropMenuPosition}
				<CropTools
					x={cropMenuPosition.x}
					y={cropMenuPosition.y}
					onApply={handleCropApply}
					onReset={handleCropReset}
					onCancel={() => {
						toggleCropMode();
					}}
					onAspectRatioChange={(ratio) => {
						cropAspectRatio = ratio;
						zoomer?.setAspectRatio(ratio);
					}}
				/>
			{/if}
		</div>
		{#if isCropping}
			<div class="crop-tools-sidebar">
				<CropTools
					variant="placed"
					onApply={handleCropApply}
					onReset={handleCropReset}
					onCancel={() => {
						toggleCropMode();
					}}
					onAspectRatioChange={(ratio) => {
						cropAspectRatio = ratio;
						zoomer?.setAspectRatio(ratio);
					}}
				/>
			</div>
		{:else if showMetadata}
			{@render metadataEditor()}
		{/if}
	</div>
</Lightbox>

<style lang="scss">
	.image-lightbox-container {
		position: relative;
		display: flex;
		align-items: center;
		height: 100%;
		width: 100%;
		pointer-events: none;
	}

	.image-container {
		position: relative;
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100%;
		height: 100%;
		pointer-events: none;
	}

	.crop-tools-sidebar {
		background-color: var(--imag-bg-color);
		padding: 1em;
		border-radius: 0.5em;
		height: 100%;
		width: auto;
		max-width: 20vw;
		min-width: 20vw;
		z-index: 100;
		pointer-events: auto;
		box-sizing: border-box;
	}

	:global(.image-icon-buttons) {
		position: absolute;
		top: 1em;
		right: 1em;
		z-index: 10;
		pointer-events: auto;
		display: flex;
		gap: 0.5em;
	}

	:global(#lightbox-icon-close) {
		position: absolute;
		top: 1em;
		left: 1em;
		z-index: 10;
		pointer-events: auto;
	}

	:global(.lightbox-button-icon) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.25em;
		background: transparent;
		border: none;

		:global(span) {
			color: var(--imag-10-dark) !important;
			fill: var(--imag-10-dark) !important;
		}

		filter: drop-shadow(0 8px 22px rgba(0, 0, 0, 1))
			drop-shadow(0 2px 6px rgba(0, 0, 0, 1))
			drop-shadow(0 2px 6px rgba(0, 0, 0, 0.1));
		-webkit-filter: drop-shadow(0 8px 22px rgba(0, 0, 0, 1))
			drop-shadow(0 2px 6px rgba(0, 0, 0, 1))
			drop-shadow(0 2px 6px rgba(0, 0, 0, 0.1));
		will-change: filter;
	}

	.image-wrapper {
		position: relative;
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		width: 100%;
		overflow: hidden;
		pointer-events: none;
	}

	.zoom-target {
		position: relative;
		display: inline-flex;
		justify-content: center;
		align-items: center;
		max-width: 100%;
		max-height: 100%;
		min-width: 0;
		min-height: 0;
		pointer-events: auto;
	}

	.lightbox-image {
		display: block;
		max-width: 100vw;
		max-height: 100vh;
		width: auto;
		height: auto;
		pointer-events: auto;
	}

	// to give space for seeing cropping
	.lightbox-image.is-crop {
		max-width: 97vw;
		max-height: 97vh;
	}

	.lightbox-placeholder {
		width: 100%;
		height: 100%;
	}

	.lightbox-nav {
		position: absolute;
		top: 50%;
		right: 2em;
		display: flex;
		flex-direction: column;
		transform: translateY(-50%);
		pointer-events: none;
	}

	.lightbox-nav-btn {
		border: none;
		color: var(--imag-10);
		width: 3em;
		height: 3em;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 0.3em;
		cursor: pointer;
		pointer-events: auto;
	}

	.metadata-editor {
		background-color: var(--imag-bg-color);
		padding: 1em;
		border-radius: 0.5em;
		color: var(--imag-10);
		height: 100%;
		width: auto;
		max-width: 20vw;
		min-width: 20vw;
		z-index: 100;
		pointer-events: auto;
		box-sizing: border-box;
		font-size: 0.85rem;
	}

	.metadata-header {
		font-size: 1rem;
		display: flex;
		align-items: center;
		margin-bottom: 1em;
		gap: 0.5em;
	}

	.metadata-exif-box {
		display: block;
	}

	.exif-card-group {
		display: flex;
		gap: 0.5em;
	}

	.exif-cards {
		display: flex;
		flex-direction: column;
		gap: 0.75em;
	}

	.exif-card {
		background: var(--imag-100);
		color: var(--imag-text-color);
		box-sizing: border-box;
		width: 100%;
		padding: 0.6em 0.8em;
		border-radius: 0.5em;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 0.4em;
		box-shadow: 0 1px 0 rgba(0, 0, 0, 0.04) inset;
	}

	.card-row {
		display: flex;
		align-items: center;
		gap: 0.6em;
		/* Allow nested flex children to shrink when content is long */
		min-width: 0;
	}

	:global(.exif-material-icon) {
		color: var(--imag-10);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		line-height: 1;
		vertical-align: middle;
	}

	.card-values {
		display: flex;
		flex-direction: column;
		gap: 0.1em;
		justify-content: center;

		/* Flex items inside a row often need a min-width:0 so long text can
		   be ellipsized instead of forcing the container to overflow */
		min-width: 0;
		flex: 1 1 auto;
	}

	:global(.value-big) {
		font-size: 1.1em;
		font-weight: 600;
	}

	.value-sub {
		font-size: 0.9em;
	}

	:global(.value-big),
	.value-sub {
		color: var(--imag-20);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.rating-container {
		margin-top: 0.75em;
		display: flex;
		align-items: center;
		gap: 0.5em;
	}

	.rating-stars {
		display: flex;
		align-items: center;
	}

	.rating-button,
	.rating-clear {
		border: none;
		background: transparent;
		cursor: pointer;
		color: var(--imag-30);
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.rating-clear {
		margin: 0em 0.5em;
	}
</style>
