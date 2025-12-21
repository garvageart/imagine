<script lang="ts">
	import { SvelteMap } from "svelte/reactivity";

	type Variants = "small" | "medium" | "large" | "xlarge";
	type BarColour =
		| "primary"
		| "secondary"
		| "100"
		| "90"
		| "80"
		| "70"
		| "60"
		| "50"
		| "40"
		| "30"
		| "20"
		| "10";

	const variantMappings = new SvelteMap<Variants, number>([
		["small", 2],
		["medium", 4],
		["large", 8],
		["xlarge", 16]
	]);

	interface Props {
		width: number;
		variant?: Variants;
		colour?: BarColour;
	}

	let {
		width = $bindable(),
		variant = "medium",
		colour = "primary"
	}: Props = $props();
</script>

<div class="progress-bar">
	<div
		class="progress-fill"
		style="width: {width}%; height: {variantMappings.get(
			variant
		)}px; background-color: var(--imag-{colour})"
	></div>
</div>

<style lang="scss">
	.progress-bar {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		background-color: var(--imag-80);

		.progress-fill {
			height: 100%;
			transition: width 0.3s ease;
		}
	}
</style>
