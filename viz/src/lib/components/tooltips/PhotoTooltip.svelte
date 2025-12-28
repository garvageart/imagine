<script lang="ts">
	import type { Image } from "$lib/api";
	import { getTakenAt } from "$lib/utils/images";
	import { DateTime } from "luxon";

	let { asset }: { asset: Image } = $props();
	let takenAt = $derived(getTakenAt(asset));
</script>

<div class="photo-tooltip-content">
	{#if asset.image_metadata?.file_name}
		<div class="tooltip-row">
			<span class="tooltip-value" title={asset.image_metadata.file_name}
				>{asset.image_metadata.file_name}</span
			>
		</div>
	{/if}
	<div class="tooltip-row">
		{#if takenAt}
			<span class="tooltip-value"
				>{DateTime.fromJSDate(takenAt).toFormat("dd LLL yyyy • HH:mm")}</span
			>
			<span class="tooltip-value">{asset.owner?.username}</span>
		{/if}
	</div>
</div>

<style lang="scss">
	:global(.tippy-box[data-theme~="viz"]) {
		background-color: var(--imag-100);
		color: var(--imag-text-color);
		border: 1px solid var(--imag-80);
		border-radius: 0.5rem;
	}

	:global(
			.tippy-box[data-theme~="viz"][data-placement^="bottom"]
				> .tippy-arrow::before
		) {
		border-bottom-color: var(--imag-100);
	}

	.photo-tooltip-content {
		display: flex;
		flex-direction: column;
		gap: 2px;
		text-align: left;
	}

	.tooltip-row {
		display: flex;
		gap: 8px;
	}

	// .tooltip-label {
	// 	font-weight: 600;
	// }

	.tooltip-value {
		opacity: 0.9;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
</style>
