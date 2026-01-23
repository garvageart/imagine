<script lang="ts">
	import { type ImageAsset } from "$lib/api";
	import { thumbHashToDataURL } from "thumbhash";
	import { onMount } from "svelte";
	import { getImageLabel, getTakenAt } from "$lib/utils/images";
	import { normalizeBase64 } from "$lib/utils/misc";
	import { VizMimeTypes } from "$lib/constants";
	import { DragData } from "$lib/drag-drop/data";
	import LabelSelector from "./LabelSelector.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import AssetImage, { type AssetImageProps } from "./AssetImage.svelte";

	let {
		asset,
		variant = "full",
		showMetadata = $bindable(true),
		objectFit = "cover",
		imageVariant = "thumbnail",
		priority = false
	}: {
		asset: ImageAsset;
		variant?: "mini" | "full";
		showMetadata?: boolean;
	} & Omit<AssetImageProps, "asset" | "variant"> & {
			imageVariant?: AssetImageProps["variant"];
		} = $props();

	let imageDate = $derived(getTakenAt(asset));
	let placeholderDataURL = $state<string | undefined>();
	let imageLoaded = $state(false);

	// Generate thumbhash placeholder
	onMount(() => {
		if (asset.image_metadata?.thumbhash) {
			try {
				const normalizedThumbhash = normalizeBase64(
					asset.image_metadata.thumbhash
				);
				const binaryString = atob(normalizedThumbhash);
				const bytes = new Uint8Array(binaryString.length);
				for (let i = 0; i < binaryString.length; i++) {
					bytes[i] = binaryString.charCodeAt(i);
				}
				placeholderDataURL = thumbHashToDataURL(bytes);
			} catch (error) {
				console.warn("Failed to decode thumbhash:", error);
			}
		}
	});

	function handleDragStart(e: DragEvent) {
		if (variant === "mini") {
			return;
		}

		if (!e.dataTransfer) {
			return;
		}

		const dragData = new DragData(VizMimeTypes.IMAGE_UIDS, [asset.uid]);
		dragData.setData(e.dataTransfer);
		e.dataTransfer.effectAllowed = "copy";

		const img = (e.currentTarget as HTMLElement).querySelector(
			".image-card-image"
		) as HTMLImageElement;
		if (img) {
			e.dataTransfer.setDragImage(img, 0, 0);
		}
	}
</script>

{#if variant === "mini"}
	<div title={asset.name} class="mini-card">
		<div class="mini-image-wrapper">
			<AssetImage
				variant={imageVariant}
				{asset}
				{objectFit}
				{priority}
				alt={asset.name}
				loading="lazy"
			/>
			{#if asset.favourited}
				<div class="favorite-badge">
					<MaterialIcon iconName="favorite" size="1rem" fill={true} />
				</div>
			{/if}
		</div>
		<div class="mini-footer">
			<span class="mini-filename"
				>{asset.image_metadata?.file_name ?? asset.name}</span
			>
			<div class="mini-meta">
				<div class="mini-rating">
					{#each Array(5) as _, i}
						<div
							class="dot"
							class:filled={i < (asset.image_metadata?.rating ?? 0)}
						></div>
					{/each}
				</div>
				{#if asset.image_metadata?.label && asset.image_metadata?.label !== "None"}
					<div
						class="mini-label-indicator"
						style="background-color: {getImageLabel(asset)}"
					></div>
				{/if}
			</div>
		</div>
	</div>
{:else}
	<div
		class="image-card"
		draggable="true"
		title={asset.name}
		role="button"
		tabindex="0"
		data-asset-id={asset.uid}
		ondragstart={handleDragStart}
		ondragend={() => {
			DragData.clear();
		}}
	>
		<div class="image-container">
			<AssetImage
				{asset}
				variant={imageVariant}
				{objectFit}
				{priority}
				alt="{asset.name}{asset.uploaded_by
					? ` by ${asset.uploaded_by.username}`
					: ''}"
				title="{asset.name}{asset.uploaded_by
					? ` by ${asset.uploaded_by.username}`
					: ''}"
				loading="lazy"
				crossorigin="use-credentials"
				onload={() => (imageLoaded = true)}
			/>
		</div>
		{#if showMetadata}
			<div class="image-card-meta">
				<span class="image-card-name" title={asset.image_metadata?.file_name}
					>{asset.image_metadata?.file_name ?? asset.name}</span
				>
				<div
					class="image-card-details"
					title="Photo taken at {imageDate.toLocaleString()}"
				>
					<div class="image-card-date-group">
						<span class="image-card-date">{imageDate.toLocaleDateString()}</span
						>
						<span class="image-card-divider">•</span>
						<span class="image-card-time"
							>{imageDate.toLocaleTimeString().replace(/:\d{2}$/, "")}</span
						>
					</div>
					<div class="image-card-indicators">
						{#if asset.image_metadata?.label && asset.image_metadata?.label !== "None"}
							<LabelSelector
								label={getImageLabel(asset)}
								enableSelection={false}
								variant={"compact"}
							/>
						{/if}
						{#if asset.favourited}
							<MaterialIcon iconName="favorite" size="1rem" fill={true} />
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}

<style lang="scss">
	.mini-card {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
		overflow: hidden;
	}

	.mini-image-wrapper {
		flex: 1;
		width: 100%;
		height: 0;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: var(--imag-90);

		.favorite-badge {
			position: absolute;
			top: 1px;
			right: 1px;
			color: #df2e69;
			filter: drop-shadow(0 0 2px rgba(0, 0, 0, 0.8));
			display: flex;
			align-items: center;
			justify-content: center;
			z-index: 5;
		}
	}

	.mini-footer {
		padding: 0.25rem;
		background-color: var(--imag-100);
		display: flex;
		flex-direction: column;
		gap: 2px;
		border-top: 1px solid var(--imag-90);

		.mini-filename {
			font-size: 0.7rem;
			font-weight: 600;
			color: var(--imag-40);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			text-align: center;
			width: 100%;
		}

		.mini-meta {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 0.4rem;
			height: 0.5rem;
		}

		.mini-rating {
			display: flex;
			gap: 2px;

			.dot {
				width: 0.25rem;
				height: 0.25rem;
				border-radius: 50%;
				background-color: var(--imag-80);

				&.filled {
					background-color: var(--imag-10);
				}
			}
		}

		.mini-label-indicator {
			width: 0.5rem;
			height: 0.5rem;
			border-radius: 1px;
			flex-shrink: 0;
		}
	}

	.image-card {
		max-height: 25em;
		background-color: var(--imag-100);
		padding: 0.8em;
		border-radius: 0.5em;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;

		&:focus {
			outline: none;
		}

		&:hover {
			background-color: var(--imag-90);
		}
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

	.image-card-details {
		color: var(--imag-20);
		display: flex;
		align-items: center;
		width: 100%;
		justify-content: space-between;
	}

	.image-card-divider {
		color: var(--imag-40);
	}

	.image-card-time {
		font-size: 0.9rem;
	}

	.image-card-indicators {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		margin-left: auto;
		min-height: 1.2em;
	}

	.image-container {
		height: 13em;
		background-color: var(--imag-80);
		display: flex;
		justify-content: center;
		align-items: center;
		position: relative;
	}
</style>
