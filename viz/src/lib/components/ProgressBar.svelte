<script lang="ts">
	import { SvelteMap } from "svelte/reactivity";

	type Variants = "small" | "medium" | "large" | "xlarge";

	const variantMappings = new SvelteMap<Variants, number>([
		["small", 2],
		["medium", 4],
		["large", 8],
		["xlarge", 16]
	]);

	let {
		width = $bindable(),
		variant = "medium"
	}: { width: number; variant?: Variants } = $props();
</script>

<div class="progress-bar">
	<div
		class="progress-fill"
		style="width: {width}%; height: {variantMappings.get(variant)}px;"
	></div>
</div>

<style lang="scss">
	.progress-bar {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		background: var(--imag-80);

		.progress-fill {
			height: 100%;
			background: var(--imag-primary);
			transition: width 0.3s ease;
		}
	}
</style>
