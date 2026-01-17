<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { computeHistogram, type HistogramData } from "$lib/histogram";
	import { getFullImagePath, type Image as APIImage } from "$lib/api";
	import InputSelect from "$lib/components/dom/InputSelect.svelte";
	import { selectionManager } from "$lib/states/selection.svelte";
	import IconButton from "$lib/components/IconButton.svelte";
	import type { HistogramChannels } from "$lib/histogram/types";
	import type { HistogramChannelData } from "$lib/third-party/photo-histogram/js";

	let activeScope = $derived(selectionManager.activeScope);
	let activeItem = $derived(activeScope?.active as APIImage | undefined);

	let imageSrcPath = $derived(
		activeItem?.image_paths?.preview
			? getFullImagePath(activeItem.image_paths.preview)
			: null
	);

	interface Props {
		src?: HTMLImageElement;
	}

	let { src }: Props = $props();

	let histReady = $state(false);
	let drawScheduled = $state(false);
	let lastMax: any = $state();

	// lol idk if this computationally expensive, maybe we should just get the actual image element
	// instead of creating a new one every time
	//
	// point being it WILL come from somewhere
	let histoSrc: HTMLImageElement | undefined = $derived.by(() => {
		if (src) {
			src.onload = () => runImageLoad(src);
			return src;
		} else if (imageSrcPath) {
			let newImage = new Image();
			newImage.crossOrigin = "anonymous";

			newImage.onload = () => runImageLoad(newImage);
			newImage.onerror = () => {
				console.warn("Histogram: image failed to load");
			};

			newImage.src = imageSrcPath;
			return newImage;
		} else {
			return undefined;
		}
	});

	// clear cached histogram when there's no image
	$effect(() => {
		if (!histoSrc) {
			histReady = false;
			lastHist = undefined;
			totals = undefined;
			scheduleRender();
		}
	});

	const CSSColours: Record<string, string> = {
		red: "#d94b4b",
		green: "#2bbf59",
		blue: "#2b7bd3",
		luminance: "rgba(220,220,220,0.95)"
	};

	let histogramEl: HTMLDivElement;

	let canvasEl: HTMLCanvasElement;
	let ctx: CanvasRenderingContext2D | undefined = $state();
	let resizeObserver: ResizeObserver | undefined = $state();

	// UI / selection state
	let selectedChannel = $state<HistogramChannels>("all");
	let isSelecting = $state(false);
	let selectionStartX: number | undefined = $state();
	let selectionEndX: number | undefined = $state();
	let selectionBins: { start: number; end: number } | undefined = $state();
	let lastHist: HistogramChannelData | undefined = $state();
	let totals:
		| {
				red: number;
				green: number;
				blue: number;
				luminance: number;
		  }
		| undefined = $state();
	let stats:
		| {
				count: number;
				percent: number;
				mean: number;
				median: number;
		  }
		| undefined = $state();

	let isReset = $state(false);
	const BINS = 256;

	async function runImageLoad(image: HTMLImageElement) {
		try {
			const data = await computeHistogram(image);
			lastHist = data.hist;
			lastMax = data.max;
			// recompute totals
			totals = {
				red: lastHist.red.reduce((a: number, b: number) => a + b, 0),
				green: lastHist.green.reduce((a: number, b: number) => a + b, 0),
				blue: lastHist.blue.reduce((a: number, b: number) => a + b, 0),
				luminance: lastHist.luminance.reduce((a: number, b: number) => a + b, 0)
			};

			histReady = true;
			isReset = false;
			scheduleRender();
		} catch (e) {
			console.warn("Histogram: failed to compute on image load", e);
			histReady = false;
		}
	}

	function scheduleRender() {
		if (drawScheduled) {
			return;
		}

		drawScheduled = true;
		requestAnimationFrame(() => {
			drawScheduled = false;
			renderHistogram();
		});
	}

	function resizeCanvas() {
		if (!canvasEl || !histogramEl) {
			return;
		}

		const dpr =
			typeof window !== "undefined" ? window.devicePixelRatio || 1 : 1;

		// CSS handles the display size (width: 100%, aspect-ratio: 3/2)
		const rect = canvasEl.getBoundingClientRect();
		const w = rect.width;
		const h = rect.height;

		// actual drawing buffer size
		canvasEl.width = Math.floor(w * dpr);
		canvasEl.height = Math.floor(h * dpr);

		if (ctx) {
			ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		}
	}

	function clearCanvas() {
		if (!ctx || !canvasEl) {
			return;
		}

		const rect = canvasEl.getBoundingClientRect();
		ctx.clearRect(0, 0, rect.width, rect.height);

		// also clear selection/overlay state so any drag highlight is removed
		selectionBins = undefined;
		selectionStartX = undefined;
		selectionEndX = undefined;
		isSelecting = false;
		stats = undefined;
	}

	function drawGrid(
		ctx: CanvasRenderingContext2D,
		width: number,
		height: number,
		padding: number
	) {
		const steps = 4;
		ctx.beginPath();

		for (let i = 1; i < steps; i++) {
			const x = Math.floor((width / steps) * i);
			ctx.moveTo(0.5 + x + padding, padding);
			ctx.lineTo(0.5 + x + padding, height + padding);
		}

		for (let i = 1; i < steps; i++) {
			const y = Math.floor((height / steps) * i);
			ctx.moveTo(padding, 0.5 + y + padding);
			ctx.lineTo(width + padding, 0.5 + y + padding);
		}

		ctx.strokeStyle = "rgba(255,255,255,0.15)";
		ctx.stroke();
	}

	async function renderHistogram() {
		if (!histogramEl) {
			clearCanvas();
			return;
		}

		if (!ctx || !canvasEl) {
			return;
		}

		if (!histReady || !lastHist) {
			clearCanvas();
			return;
		}

		const rect = canvasEl.getBoundingClientRect();
		const width = rect.width;
		const height = rect.height;

		ctx.clearRect(0, 0, width, height);

		const hist = lastHist;

		// choose max across channels for normalization
		const maxVal = Math.max(
			lastMax?.red ?? 1,
			lastMax?.green ?? 1,
			lastMax?.blue ?? 1,
			lastMax?.luminance ?? 1,
			1
		);

		const sqrtMax = Math.sqrt(maxVal);

		function drawChannel(
			channel: number[],
			color: string,
			alpha = 0.7,
			compositeOp: GlobalCompositeOperation = "lighter"
		) {
			if (!ctx) {
				return;
			}

			ctx.save();
			ctx.globalCompositeOperation = compositeOp;
			ctx.globalAlpha = alpha;
			ctx.beginPath();
			ctx.moveTo(0, height);

			for (let i = 0; i < BINS; i++) {
				const x = (i / (BINS - 1)) * width;
				const sqrtValue = Math.sqrt(channel[i]);
				const y = height - (sqrtValue / sqrtMax) * height;
				ctx.lineTo(x, y);
			}

			ctx.lineTo(width, height);
			ctx.closePath();
			ctx.fillStyle = color;
			ctx.fill();
			ctx.restore();
		}

		drawGrid(ctx, width, height, 0);

		// draw channels according to selection
		if (selectedChannel === "all") {
			drawChannel(hist.blue, CSSColours.blue, 0.8, "color");
			drawChannel(hist.green, CSSColours.green, 0.8, "color");
			drawChannel(hist.red, CSSColours.red, 0.8, "color");
			drawChannel(hist.luminance, CSSColours.luminance, 0.1, "source-over");
		} else if (selectedChannel === "red") {
			drawChannel(hist.red, CSSColours.red, 0.95, "source-over");
		} else if (selectedChannel === "green") {
			drawChannel(hist.green, CSSColours.green, 0.95, "source-over");
		} else if (selectedChannel === "blue") {
			drawChannel(hist.blue, CSSColours.blue, 0.95, "source-over");
		} else if (selectedChannel === "luminance") {
			// draw luminance as filled gray
			drawChannel(hist.luminance, CSSColours.luminance, 0.95, "source-over");
		}

		// draw luminance as semi-transparent white line on top
		ctx.save();
		ctx.globalCompositeOperation = "source-over";
		ctx.globalAlpha = 0.9;
		ctx.strokeStyle = CSSColours.luminance;
		ctx.lineWidth = 1.5;
		ctx.beginPath();

		for (let i = 0; i < BINS; i++) {
			const x = (i / (BINS - 1)) * width;
			const sqrtValue = Math.sqrt(hist.luminance[i]);
			const y = height - (sqrtValue / sqrtMax) * height;
			if (i === 0) {
				ctx.moveTo(x, y);
			} else {
				ctx.lineTo(x, y);
			}
		}

		ctx.stroke();
		ctx.restore();

		// draw selection overlay
		if (selectionBins && selectionBins.end >= selectionBins.start) {
			const startX = (selectionBins.start / (BINS - 1)) * width;
			const endX = (selectionBins.end / (BINS - 1)) * width;
			ctx.save();

			ctx.globalCompositeOperation = "source-over";
			ctx.fillStyle = "rgba(255,255,255,0.08)";
			ctx.fillRect(startX, 0, endX - startX, height);
			ctx.strokeStyle = "rgba(255,255,255,0.25)";
			ctx.lineWidth = 1;
			ctx.strokeRect(startX + 0.5, 0.5, endX - startX - 1, height - 1);
			ctx.restore();
		}

		if (selectionBins) {
			computeSelectionStats(selectionBins.start, selectionBins.end);
		} else {
			if (isReset) {
				stats = undefined;
			} else {
				computeSelectionStats();
			}
		}
	}

	function computeSelectionStats(startBin?: number, endBin?: number) {
		if (!lastHist || !totals) {
			return;
		}

		const start = startBin ?? 0;
		const end = endBin ?? BINS - 1;

		const channel = selectedChannel === "all" ? "luminance" : selectedChannel;
		const arr = lastHist[channel === "luminance" ? "luminance" : channel];
		let count = 0;
		let weighted = 0;
		for (let i = start; i <= end; i++) {
			const c = arr[i] || 0;
			count += c;
			weighted += c * i;
		}

		const total = totals[channel === "luminance" ? "luminance" : channel] || 1;
		const mean = count ? weighted / count : 0;

		// median within selected bins
		let median = 0;
		if (count) {
			let half = Math.floor(count / 2);
			let medianCount = 0;

			for (let i = start; i <= end; i++) {
				medianCount += arr[i] || 0;
				if (medianCount >= half) {
					median = i;
					break;
				}
			}
		}

		stats = {
			count,
			percent: (count / (total || 1)) * 100,
			mean,
			median
		};
	}

	const pointerDown = (ev: PointerEvent) => {
		isReset = false;
		canvasEl.setPointerCapture(ev.pointerId);
		isSelecting = true;
		selectionStartX = ev.offsetX;
		selectionEndX = ev.offsetX;

		scheduleRender();
	};

	const pointerMove = (ev: PointerEvent) => {
		if (!isSelecting) {
			return;
		}

		selectionEndX = Math.max(0, Math.min(ev.offsetX, canvasEl.clientWidth));

		const startX = Math.min(selectionStartX ?? 0, selectionEndX ?? 0);
		const endX = Math.max(selectionStartX ?? 0, selectionEndX ?? 0);

		const startBin = Math.round((startX / canvasEl.clientWidth) * (BINS - 1));
		const endBin = Math.round((endX / canvasEl.clientWidth) * (BINS - 1));

		selectionBins = {
			start: startBin,
			end: endBin
		};

		scheduleRender();
	};

	const pointerUp = (ev: PointerEvent) => {
		try {
			canvasEl.releasePointerCapture(ev.pointerId);
		} catch (e) {}

		isSelecting = false;
		if (selectionStartX !== undefined && selectionEndX !== undefined) {
			const sx = Math.min(selectionStartX, selectionEndX);
			const ex = Math.max(selectionStartX, selectionEndX);

			const startBin = Math.round((sx / canvasEl.clientWidth) * (BINS - 1));
			const endBin = Math.round((ex / canvasEl.clientWidth) * (BINS - 1));
			selectionBins = { start: startBin, end: endBin };

			computeSelectionStats(startBin, endBin);
		}

		scheduleRender();
	};

	onMount(() => {
		ctx = canvasEl.getContext("2d") || undefined;

		if (typeof ResizeObserver !== "undefined") {
			resizeObserver = new ResizeObserver(() => {
				resizeCanvas();
				scheduleRender();
			});
			resizeObserver.observe(histogramEl);
		} else {
			const onResize = () => {
				resizeCanvas();
				scheduleRender();
			};
			window.addEventListener("resize", onResize);
			resizeObserver = {
				disconnect: () => window.removeEventListener("resize", onResize)
			} as any;
		}

		canvasEl.addEventListener("pointerdown", pointerDown);
		canvasEl.addEventListener("pointermove", pointerMove);
		window.addEventListener("pointerup", pointerUp);

		// initial render
		scheduleRender();
	});

	onDestroy(() => {
		canvasEl.removeEventListener("pointerdown", pointerDown);
		canvasEl.removeEventListener("pointermove", pointerMove);
		window.removeEventListener("pointerup", pointerUp);
	});

	function resetCanvas() {
		isReset = true;
		selectedChannel = "all";

		resizeCanvas();
		clearCanvas();
		scheduleRender();

		stats = undefined;
		selectionBins = undefined;
	}

	$effect(() => {
		if (histoSrc) {
			scheduleRender();
		} else {
			clearCanvas();
		}
	});

	onDestroy(() => {
		if (resizeObserver) {
			try {
				resizeObserver.disconnect();
			} catch (e) {}
		}
	});
</script>

<div class="histogram-container" bind:this={histogramEl}>
	<div class="controls">
		<InputSelect
			id="channel"
			bind:value={selectedChannel}
			onchange={() => scheduleRender()}
			labelPosition="side"
			label="Channel"
			style="font-size: 0.75rem; padding: 0.2rem 0.75rem;"
		>
			<option value="all">All (RGB)</option>
			<option value="red">Red</option>
			<option value="green">Green</option>
			<option value="blue">Blue</option>
			<option value="luminance">Luminance</option>
		</InputSelect>

		<IconButton
			iconName="refresh"
			onclick={resetCanvas}
			title="Reset Histogram"
		/>
	</div>
	<canvas
		class="no-select"
		oncontextmenu={(e) => e.preventDefault()}
		bind:this={canvasEl}
	></canvas>
	{#snippet stat(label: string, value: string)}
		<div class="stat-row" title={value}>
			<strong>{label}:</strong>
			<span class="stat-val">{value}</span>
		</div>
	{/snippet}

	<div class="stats">
		{@render stat(
			"Range",
			selectionBins
				? `${selectionBins.start}–${selectionBins.end}`
				: `0–${BINS - 1}`
		)}
		{@render stat(
			"Pixels",
			stats ? `${stats.count} (${stats.percent.toFixed(2)}%)` : "—"
		)}
		{@render stat("Mean", stats ? stats.mean.toFixed(2) : "—")}
		{@render stat("Median", stats ? stats.median.toString() : "—")}
	</div>
</div>

<style lang="scss">
	.histogram-container {
		display: flex;
		align-items: center;
		justify-content: flex-start;
		flex-direction: column;
		box-sizing: border-box;
		height: 100%;
		padding: 0.5rem;
		color: var(--imag-text-color);
		position: relative;
		background-color: var(--imag-100);
		font-size: 0.75rem;
		gap: 0.75rem;
		overflow-y: auto;
		overflow-x: hidden;
		container-type: inline-size;

		canvas {
			display: block;
			width: 100%;
			aspect-ratio: 3 / 2;
			border: 1px solid rgb(105, 105, 105);
			background-color: rgb(56, 56, 56);
			flex-shrink: 0;
		}
	}

	.controls {
		display: flex;
		width: 100%;
		justify-content: space-between;
		align-items: center;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 0.2rem 0.5rem;
		width: 100%;
		font-family: var(--imag-code-font);
		flex-shrink: 0;
	}

	.stat-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 0.5rem;
	}

	.stat-row .stat-val {
		text-align: left;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
</style>
