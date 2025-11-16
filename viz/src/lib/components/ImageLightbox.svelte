<script lang="ts">
	import { loadImage } from "$lib/utils/dom";
	import { fade } from "svelte/transition";
	import Lightbox from "./Lightbox.svelte";
	import LoadingContainer from "./LoadingContainer.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import { getFullImagePath, updateImage, type Image, type ImageUpdate } from "$lib/api";
	import hotkeys from "hotkeys-js";
	import { formatBytes, getThumbhashURL } from "$lib/utils/images";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

	interface Props {
		lightboxImage: Image | undefined;
		prevLightboxImage?: () => void;
		nextLightboxImage?: () => void;
	}

	let { lightboxImage = $bindable(), prevLightboxImage, nextLightboxImage }: Props = $props();

	let show = $derived(lightboxImage !== undefined);
	let imageToLoad = $derived(getFullImagePath(lightboxImage!.image_paths?.preview || lightboxImage!.image_paths?.original));

	let direction = $state<"left" | "right">("right");
	let showMetadata = $state(false);

	function goToPrev() {
		direction = "left";
		prevLightboxImage?.();
	}

	function goToNext() {
		direction = "right";
		nextLightboxImage?.();
	}

	let thumbhashURL = $derived(lightboxImage ? getThumbhashURL(lightboxImage) : undefined);
	let currentImageEl: HTMLImageElement | undefined = $derived(lightboxImage ? document.createElement("img") : undefined);

	// Rating UI state: previewRating for hover preview, rating is the set value
	let previewRating = $state<number | null>(null);
	let rating = $state<number | null>(lightboxImage?.image_metadata?.rating ?? null);

	// Star values moved to state to avoid rebuilding the array on each render
	let starValues = $state<number[]>([1, 2, 3, 4, 5]);

	// Prevent concurrent rating updates
	let updatingRating = $state(false);

	async function setRating(newRating: number | null) {
		if (!lightboxImage) {
			return;
		}

		if (updatingRating) {
			return;
		}

		updatingRating = true;
		const prev = rating;
		rating = newRating;

		const uid = lightboxImage.uid;
		const dateToUpdate: ImageUpdate = {
			image_metadata: { rating: newRating }
		};

		try {
			const res = await updateImage(uid, dateToUpdate);
			if (res.status === 200) {
				lightboxImage = res.data;
				rating = res.data.image_metadata?.rating ?? null;
			} else {
				toastState.addToast({
					type: "error",
					message: `Failed to update rating: ${res.status} ${res.data.error}`
				});

				rating = prev;
			}
		} catch (err) {
			console.error("Error updating rating", err);
			rating = prev;
		} finally {
			updatingRating = false;
		}
	}

	hotkeys("left,right", (e, handler) => {
		if (!show) {
			return;
		}

		e.preventDefault();
		if (handler.key === "left") {
			goToPrev();
		} else if (handler.key === "right") {
			goToNext();
		}
	});

	function formatFileSize() {
		const size = lightboxImage?.image_metadata?.file_size;
		return formatBytes(size) ?? "—";
	}
</script>

{#snippet metadataEditor()}
	<div class="metadata-editor">
		<div class="metadata-header">
			<button title="Close" onclick={() => (lightboxImage = undefined)}>
				<MaterialIcon iconName="close" />
			</button>
			<h3>Info</h3>
		</div>
		<div class="metadata-exif-box">
			<div class="exif-cards">
				<!-- Camera/Exposure card -->
				<div class="exif-card">
					<div class="card-row main-row">
						<div class="card-values">
							{#if lightboxImage?.exif?.model && lightboxImage?.exif?.make}
								<div class="value-big">
									{lightboxImage.exif.make}
									{lightboxImage.exif.model.replace(new RegExp(`^${lightboxImage.exif.make} `), "")}
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
								<div class="value-sub">{lightboxImage?.exif?.f_number ?? lightboxImage?.exif?.aperture ?? "—"}</div>
								<div class="value-sub">{lightboxImage?.exif?.exposure_time ?? "—"}</div>
							</div>
						</div>
						<div class="card-row meta-row">
							<MaterialIcon iconName="tune" class="exif-material-icon" />
							<div class="card-values">
								<div class="value-sub">
									ISO {lightboxImage?.exif?.iso ?? "—"} · {lightboxImage?.exif?.exposure_value ?? "—"}
								</div>
							</div>
						</div>
					</div>

					<div class="exif-card">
						<div class="card-row main-row">
							<div class="card-values">
								<div class="value-sub">{lightboxImage?.width} x {lightboxImage?.height}</div>
							</div>
						</div>
						<div class="card-row main-row">
							<MaterialIcon iconName="aspect_ratio" class="exif-material-icon" />
							<div class="card-values">
								<div class="value-sub">{Math.floor((lightboxImage?.width! * lightboxImage?.height!) / 1_000_000)} MP</div>
								<div class="value-sub">{formatFileSize()}</div>
							</div>
						</div>
						<div class="card-row meta-row">
							<MaterialIcon iconName="palette" class="exif-material-icon" />
							<div class="card-values">
								<div class="value-sub">{lightboxImage?.image_metadata?.color_space ?? "—"}</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<div class="rating-container">
				<div class="rating-stars" role="group" onmouseleave={() => (previewRating = null)}>
					{#each starValues as i}
						<button
							class="rating-button"
							aria-label={`Set rating ${i}`}
							onmouseenter={() => (previewRating = i)}
							onmouseleave={() => (previewRating = null)}
							onclick={() => setRating(i)}
							disabled={updatingRating}
						>
							<MaterialIcon fill={i <= (previewRating ?? rating ?? 0)} iconName="star" iconStyle={"sharp"} />
						</button>
					{/each}
					{#if rating !== null && rating !== 0}
						<button class="rating-clear" aria-label="Clear rating" onclick={() => setRating(0)} disabled={updatingRating}>
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
	backgroundOpacity={0.9}
	onclick={() => {
		lightboxImage = undefined;
	}}
>
	<div class="image-lightbox-container">
		<div class="image-container">
			<div class="image-icon-buttons">
				<button
					class="image-icon-buttons-shadow"
					title={`${showMetadata ? "Hide" : "Show"} Info`}
					onclick={(e) => {
						e.stopPropagation();
						showMetadata = !showMetadata;
					}}
				>
					<MaterialIcon iconName="info" />
				</button>
			</div>
			{#key lightboxImage?.uid}
				<div class="image-wrapper">
					{#await loadImage(imageToLoad, currentImageEl!)}
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
							/>
						{/if}
					{:then _}
						<img
							src={imageToLoad}
							class="lightbox-image"
							alt={lightboxImage!.name}
							title={lightboxImage!.name}
							loading="eager"
							data-image-id={lightboxImage!.uid}
							in:fade
							out:fade
						/>
					{:catch error}
						<p>Failed to load image</p>
						<p>{error}</p>
					{/await}
				</div>
			{/key}

			{#if prevLightboxImage && nextLightboxImage}
				<div class="lightbox-nav">
					<button
						class="lightbox-nav-btn prev image-icon-buttons-shadow"
						aria-label="Previous image"
						onclick={(e) => {
							e.stopPropagation();
							goToPrev();
						}}
					>
						<MaterialIcon iconName="arrow_back" />
					</button>
					<button
						class="lightbox-nav-btn next image-icon-buttons-shadow"
						aria-label="Next image"
						onclick={(e) => {
							e.stopPropagation();
							goToNext();
						}}
					>
						<MaterialIcon iconName="arrow_forward" />
					</button>
				</div>
			{/if}
		</div>
		{#if showMetadata}
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

	.image-icon-buttons {
		position: absolute;
		top: 1em;
		right: 1em;
		z-index: 10;
		pointer-events: auto;
		display: flex;
		gap: 0.5em;

		button {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			padding: 0.25em;
			background: transparent;
			border: none;
		}
	}

	.image-icon-buttons-shadow {
		filter: drop-shadow(0 8px 22px rgba(0, 0, 0, 1)) drop-shadow(0 2px 6px rgba(0, 0, 0, 1))
			drop-shadow(0 2px 6px rgba(0, 0, 0, 0.1));
		-webkit-filter: drop-shadow(0 8px 22px rgba(0, 0, 0, 1)) drop-shadow(0 2px 6px rgba(0, 0, 0, 1))
			drop-shadow(0 2px 6px rgba(0, 0, 0, 0.1));
		will-change: filter;
	}

	.image-wrapper {
		position: relative;
		display: flex;
		justify-content: center;
		align-items: center;
		max-height: 95%;
		height: 100%;
		width: 100%;
		overflow: hidden;
		pointer-events: none;
	}

	.lightbox-image {
		max-width: 100%;
		max-height: 100%;
		object-fit: contain;
		pointer-events: auto;
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
		width: 25vw;
		max-width: 25vw;
		z-index: 100;
		pointer-events: auto;
		box-sizing: border-box;
		font-size: 0.9rem;
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
		width: 100%;
		padding: 0.6em 0.8em;
		border-radius: 0.5em;
		display: flex;
		flex-direction: column;
		gap: 0.4em;
		box-shadow: 0 1px 0 rgba(0, 0, 0, 0.04) inset;
	}

	.card-row {
		display: flex;
		align-items: center;
		gap: 0.6em;
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
	}

	.value-big {
		font-size: 1.1em;
		font-weight: 600;
	}

	.value-sub {
		font-size: 0.9em;
	}

	.value-big,
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
