<script lang="ts">
	import type { Snippet } from "svelte";
	import type { SvelteHTMLElements } from "svelte/elements";

	let { children, ...props }: { children: Snippet } & SvelteHTMLElements["div"] = $props();
</script>

<div {...props} id="viz-toolbar-container" style="top: {0}px; {props.style}">
	{@render children?.()}
</div>

<style lang="scss">
	#viz-toolbar-container {
		position: sticky;
		z-index: 1;
		padding: 0.5em 2em;
		background-color: rgba(39, 51, 74, 0.9);
		backdrop-filter: blur(5px);
		border-bottom: 1px solid var(--imag-60);
		min-height: 4em;
		width: 100%;
		max-width: 100%;
		display: flex;
		align-items: center;
		flex-direction: row;
		box-sizing: border-box;
	}
</style>
