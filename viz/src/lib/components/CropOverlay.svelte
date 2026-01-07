<script lang="ts">
	import type {
		ZoomPanCrop,
		CropRect,
		DragAction
	} from "$lib/images/zoom/crop";
	import { DragActionName } from "$lib/images/zoom/crop";

	interface Props {
		width: number;
		height: number;
		crop: CropRect;
		zoomer: ZoomPanCrop;
		scale?: number;
	}

	let { width, height, crop, zoomer, scale = 1 }: Props = $props();

	function handleStart(action: DragAction, e: MouseEvent | TouchEvent) {
		zoomer.startCropDrag(action, e);
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === "Enter" || e.key === " ") {
			e.preventDefault();
		}
	}

	// Since we are inside the transformed container, we use image-relative coordinates directly.
	let tStyle = $derived(
		`left: ${crop.x}px; top: ${crop.y}px; width: ${crop.width}px; height: ${crop.height}px;`
	);

	// Inverse scale for UI elements to keep them constant visual size.
	let invScale = $derived(1 / scale);
	let handleStyle = $derived(
		`transform: translate(-50%, -50%) scale(${invScale});`
	);
	// Ensure minimum border width visibility (at least 1px)
	let borderWidth = $derived(Math.max(1, invScale));
	let outlineStyle = $derived(`outline-width: ${borderWidth}px;`);
</script>

{#snippet dragHandleSnippet(dragAction: DragAction)}
	<div
		class="handle {dragAction}"
		style={handleStyle}
		onmousedown={(e) => handleStart(dragAction, e)}
		ontouchstart={(e) => handleStart(dragAction, e)}
		ondragstart={(e) => e.preventDefault()}
		role="button"
		tabindex="-1"
		aria-label="{dragAction === 'move' ? 'Move' : 'Resize'} {dragAction}"
	></div>
{/snippet}

<div class="crop-overlay-container" style="width: 100%; height: 100%;">
	<div
		class="crop-box"
		style="{tStyle} {outlineStyle}"
		onmousedown={(e) => handleStart("move", e)}
		ontouchstart={(e) => handleStart("move", e)}
		onmouseup={(e) => e.stopPropagation()}
		onclick={(e) => e.stopPropagation()}
		onkeydown={handleKeyDown}
		ondragstart={(e) => e.preventDefault()}
		role="button"
		tabindex="0"
		aria-label="Crop Area"
	>
		<!-- Grid lines (Rule of thirds) -->
		<div class="grid-line v v1" style="width: {borderWidth}px"></div>
		<div class="grid-line v v2" style="width: {borderWidth}px"></div>
		<div class="grid-line h h1" style="height: {borderWidth}px"></div>
		<div class="grid-line h h2" style="height: {borderWidth}px"></div>

		<!-- Handles -->
		{#each ["nw", "n", "ne", "e", "se", "s", "sw", "w"] as dragAction}
			{@render dragHandleSnippet(dragAction as DragAction)}
		{/each}
	</div>
</div>

<style>
	.crop-overlay-container {
		position: absolute;
		top: 0;
		left: 0;
		pointer-events: none; /* Let clicks pass through to image if not hitting crop box */
		touch-action: none;
	}

	.crop-box {
		position: absolute;
		/* Use a massive box-shadow to create the dimmed overlay effect */
		/* Inner white border (1px), Outer black semi-transparent border (1px), then the dimmer */
		box-shadow:
			inset 0 0 0 1px rgba(255, 255, 255, 0.7),
			0 0 0 1px rgba(0, 0, 0, 0.3),
			0 0 0 9999px rgba(0, 0, 0, 0.6);
		outline: none;
		pointer-events: auto;
		cursor: move;
		touch-action: none;
	}

	.grid-line {
		position: absolute;
		background-color: rgba(255, 255, 255, 0.4);
		pointer-events: none;
	}
	.v {
		top: 0;
		bottom: 0;
		width: 1px;
		transform: translateX(-50%); /* Center the line */
	}
	.h {
		left: 0;
		right: 0;
		height: 1px;
		transform: translateY(-50%); /* Center the line */
	}
	.v1 {
		left: 33.33%;
	}
	.v2 {
		left: 66.66%;
	}
	.h1 {
		top: 33.33%;
	}
	.h2 {
		top: 66.66%;
	}

	/* Handles */
	.handle {
		position: absolute;
		width: 14px;
		height: 14px;
		background-color: white;
		border: 1px solid rgba(0, 0, 0, 0.2);
		border-radius: 50%;
		z-index: 10;
		touch-action: none;
	}

	/* For bigger grab area */
	.handle::after {
		content: "";
		position: absolute;
		top: -10px;
		left: -10px;
		right: -10px;
		bottom: -10px;
		background: transparent;
		border-radius: 50%;
	}

	.nw {
		top: 0;
		left: 0;
		cursor: nw-resize;
	}
	.n {
		top: 0;
		left: 50%;
		cursor: n-resize;
	}
	.ne {
		top: 0;
		left: 100%;
		cursor: ne-resize;
	}
	.e {
		top: 50%;
		left: 100%;
		cursor: e-resize;
	}
	.se {
		top: 100%;
		left: 100%;
		cursor: se-resize;
	}
	.s {
		top: 100%;
		left: 50%;
		cursor: s-resize;
	}
	.sw {
		top: 100%;
		left: 0;
		cursor: sw-resize;
	}
	.w {
		top: 50%;
		left: 0;
		cursor: w-resize;
	}
</style>
