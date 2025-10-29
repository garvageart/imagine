<script lang="ts">
	import type { Snippet } from "svelte";
	import type { SvelteHTMLElements } from "svelte/elements";

	let {
		children,
		stickyToolbar = true,
		...props
	}: { children: Snippet; stickyToolbar?: boolean } & SvelteHTMLElements["div"] = $props();
</script>

<div
	{...props}
	class="viz-toolbar-container"
	style={`${props.style} ${stickyToolbar === true ? "position: sticky; top: 0px;" : "position: relative;"}`}
>
	{@render children?.()}
</div>

<style lang="scss">
	.viz-toolbar-container {
		z-index: 1;
		padding: 0.5em 2em;
		background-color: var(--imag-100);
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
