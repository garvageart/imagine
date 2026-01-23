<script lang="ts">
	import type { Image, Collection } from "$lib/api";
	import { VizMimeTypes } from "$lib/constants";
	import { DragData } from "$lib/drag-drop/data";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { slide } from "svelte/transition";
	import { formatBytes } from "$lib/utils/images";
	import AssetImage from "$lib/components/AssetImage.svelte";

	interface Props {
		item: Image | Collection;
		type: "image" | "collection";
	}

	let { item, type }: Props = $props();
	let expanded = $state(false);

	function handleDragStart(e: DragEvent) {
		if (!e.dataTransfer) return;

		const mimeType =
			type === "image" ? VizMimeTypes.IMAGE_UIDS : VizMimeTypes.COLLECTION_UIDS;

		const dragData = new DragData(mimeType, [item.uid]);
		dragData.setData(e.dataTransfer);
		e.dataTransfer.effectAllowed = "copy";

		// Set drag image if possible
		const imgEl = (e.target as HTMLElement).querySelector("img");
		if (imgEl) {
			e.dataTransfer.setDragImage(imgEl, 10, 10);
		}
	}

	function formatDate(dateString: string) {
		if (!dateString) return "-";
		return new Date(dateString).toLocaleDateString(undefined, {
			year: "numeric",
			month: "short",
			day: "numeric"
		});
	}

	const thumbnailAsset = $derived.by(() => {
		if (type === "image") {
			return item as Image;
		}
		return (item as Collection).thumbnail;
	});
</script>

<div
	class="compact-item"
	draggable="true"
	ondragstart={handleDragStart}
	role="listitem"
>
	<div
		class="header"
		onclick={() => (expanded = !expanded)}
		onkeydown={(e) => e.key === "Enter" && (expanded = !expanded)}
		role="button"
		tabindex="0"
	>
		<div class="thumb-container">
			{#if thumbnailAsset}
				<AssetImage
					asset={thumbnailAsset}
					alt={item.name}
					variant="thumbnail"
					objectFit="cover"
				/>
			{:else}
				<div class="placeholder">
					<MaterialIcon
						iconName={type === "collection" ? "folder" : "image"}
						size="1.2rem"
					/>
				</div>
			{/if}
		</div>
		<div class="info">
			<span class="name" title={item.name}>{item.name}</span>
		</div>
		<button class="expand-btn" aria-label={expanded ? "Collapse" : "Expand"}>
			<MaterialIcon
				iconName={expanded ? "expand_less" : "expand_more"}
				size="1.2rem"
			/>
		</button>
	</div>

	{#if expanded}
		<div class="details" transition:slide={{ duration: 200 }}>
			{#if type === "image"}
				{@const img = item as Image}
				<div class="detail-grid">
					<span class="label">UID:</span>
					<span class="value uid" title={item.uid}>{item.uid}</span>

					<span class="label">Dimensions:</span>
					<span class="value">{img.width} x {img.height}</span>

					<span class="label">Size:</span>
					<span class="value"
						>{formatBytes(img.image_metadata?.file_size) ?? "-"}</span
					>

					<span class="label">Type:</span>
					<span class="value"
						>{img.image_metadata?.file_type?.toUpperCase() ?? "-"}</span
					>

					<span class="label">Created:</span>
					<span class="value">{formatDate(img.created_at)}</span>

					<span class="label">Rating:</span>
					<span class="value">{img.image_metadata?.rating ?? "-"}</span>
				</div>
			{:else}
				{@const col = item as Collection}
				<div class="detail-grid">
					<span class="label">UID:</span>
					<span class="value uid" title={item.uid}>{item.uid}</span>

					<span class="label">Item Count:</span>
					<span class="value">{col.image_count}</span>

					<span class="label">Created:</span>
					<span class="value">{formatDate(col.created_at)}</span>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="scss">
	.compact-item {
		display: flex;
		flex-direction: column;
		background-color: var(--imag-100);
		border-radius: 0.25rem;
		border: 1px solid transparent;
		transition: background-color 0.2s;
		overflow: hidden;

		&:hover {
			background-color: var(--imag-95);
		}
	}

	.header {
		display: flex;
		align-items: center;
		padding: 0.25rem 0.5rem;
		gap: 0.5rem;
		cursor: pointer;
		user-select: none;
		height: 2.5rem;
	}

	.thumb-container {
		width: 2rem;
		height: 2rem;
		flex-shrink: 0;
		border-radius: 0.25rem;
		overflow: hidden;
		background-color: var(--imag-80);
		display: flex;
		align-items: center;
		justify-content: center;

		:global(.asset-image-container) {
			width: 100%;
			height: 100%;
		}

		.placeholder {
			color: var(--imag-60);
			display: flex;
		}
	}

	.info {
		flex: 1;
		overflow: hidden;
		display: flex;
		align-items: center;
	}

	.name {
		font-size: 0.85rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--imag-text-color);
	}

	.expand-btn {
		background: none;
		border: none;
		color: var(--imag-60);
		cursor: pointer;
		padding: 0;
		display: flex;
		align-items: center;

		&:hover {
			color: var(--imag-text-color);
		}
	}

	.details {
		padding: 0.5rem;
		border-top: 1px solid var(--imag-90);
		background-color: var(--imag-95);
		font-size: 0.75rem;
	}

	.detail-grid {
		display: grid;
		grid-template-columns: min-content 1fr;
		gap: 0.25rem 0.75rem;
		margin-top: 0.25rem;

		.label {
			color: var(--imag-40);
			text-align: right;
			white-space: nowrap;
		}

		.value {
			color: var(--imag-text-color);
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}

		.uid {
			font-weight: 700;
			font-family: var(--imag-code-font);
		}
	}
</style>
