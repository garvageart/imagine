<script module lang="ts">
	export function getAssetDate(asset: Image): Date {
		return asset.exif?.date_time_original
			? new Date(convertEXIFDateTime(asset.exif?.date_time_original))
			: new Date(asset.image_metadata?.file_created_at ?? asset.created_at);
	}
</script>

<script lang="ts">
	import { DateTime } from "luxon";
	import { getFullImagePath, type Image } from "$lib/api";
	import { thumbHashToDataURL } from "thumbhash";
	import { onMount } from "svelte";
	import { convertEXIFDateTime } from "$lib/utils/images";
	import { normalizeBase64 } from "$lib/utils/misc";

	let { asset }: { asset: Image } = $props();
	const imageDate = DateTime.fromJSDate(getAssetDate(asset));

	let placeholderDataURL = $state<string | undefined>();
	let imageLoaded = $state(false);

	// Generate thumbhash placeholder
	onMount(() => {
		if (asset.image_metadata?.thumbhash) {
			try {
				const normalizedThumbhash = normalizeBase64(asset.image_metadata.thumbhash);
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
</script>

<div class="image-card" data-asset-id={asset.uid}>
	<div class="image-container">
		{#if placeholderDataURL && !imageLoaded}
			<img class="image-card-placeholder" src={placeholderDataURL} alt="" aria-hidden="true" />
		{/if}
		<img
			class="image-card-image"
			class:loaded={imageLoaded}
			src={getFullImagePath(asset.image_paths?.thumbnail)}
			alt="{asset.name}{asset.uploaded_by ? ` by ${asset.uploaded_by.username}` : ''}"
			title="{asset.name}{asset.uploaded_by ? ` by ${asset.uploaded_by.username}` : ''}"
			loading="lazy"
			onload={() => (imageLoaded = true)}
		/>
	</div>
	<div class="image-card-meta">
		<span class="image-card-name" title={asset.image_metadata?.file_name}>{asset.image_metadata?.file_name ?? asset.name}</span>
		<div class="image-card-date_time" title="Photo taken at {imageDate.toFormat('dd/MM/yyyy - HH:mm')}">
			<span class="image-card-date">{imageDate.toFormat("dd/MM/yyyy")}</span>
			<span class="image-card-divider">•</span>
			<span class="image-card-time">{imageDate.toFormat("HH:mm")}</span>
		</div>
	</div>
</div>

<style lang="scss">
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

	.image-card img {
		width: 100%;
		height: 100%;
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

	.image-container {
		height: 13em;
		background-color: var(--imag-80);
		display: flex;
		justify-content: center;
		align-items: center;
		position: relative;
	}

	.image-card-placeholder {
		position: absolute;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.image-card-image {
		opacity: 0;
		transition: opacity 0.3s ease-in-out;

		&.loaded {
			opacity: 1;
		}
	}
</style>
