<script lang="ts">
	import type { SvelteHTMLElements } from "svelte/elements";

	let { iconName, showHoverBG, ...props }: { iconName: string; showHoverBG?: boolean } & SvelteHTMLElements["span"] = $props();
	let showHover: boolean = $state(false);
</script>

<span
	class="material-symbols-sharp {props.class} {showHoverBG && showHover ? 'hover' : ''}"
	{...props}
	onmouseover={() => {
		showHover = true;
	}}
	onmouseout={() => {
		showHover = false;
	}}
>
	{iconName}
</span>

<style lang="scss">
	.material-symbols-sharp {
		transition: all 150ms linear;
		padding: 0.1em;
		margin: 0.2em;
		display: inline-block;
		border-radius: 100%;

		&:active {
			background-color: var(--imag-80);
		}
	}

	.hover {
		background-color: var(--imag-90);
	}
</style>
