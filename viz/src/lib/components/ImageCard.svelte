<script lang="ts">
	import { DateTime } from "luxon";
	import type { ImageObjectData } from "$lib/entities/image";

	let { asset }: { asset: ImageObjectData } = $props();
	const imageDate = asset.added_at ? DateTime.fromJSDate(asset.added_at) : null;
</script>

<div class="image-card" data-asset-id={asset.uid}>
	<div class="image-container">
		<img
			class="image-card-image"
			src={asset.image_paths?.preview_path}
			alt="{asset.name}{asset.uploaded_by ? ` by ${asset.uploaded_by}` : ''}"
			title="{asset.name}{asset.uploaded_by ? ` by ${asset.uploaded_by}` : ''}"
			loading="lazy"
		/>
	</div>
	<div class="image-card-meta">
		<span class="image-card-name" title={asset.image_metadata?.file_name}>{asset.image_metadata?.file_name ?? asset.name}</span>
		{#if imageDate}
			<div class="image-card-date_time" title="Photo taken at {imageDate.toFormat('dd/MM/yyyy - HH:mm')}">
				<span class="image-card-date">{imageDate.toFormat("dd/MM/yyyy")}</span>
				<span class="image-card-divider">•</span>
				<span class="image-card-time">{imageDate.toFormat("HH:mm")}</span>
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.image-card {
		max-height: 20em;
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
		max-width: 100%;
		min-height: 100%;
		height: auto;
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
	}
</style>
