<script lang="ts">
	import type { SvelteHTMLElements } from "svelte/elements";

	let { iconName, ...props }: { iconName: string } & SvelteHTMLElements["span"] = $props();
</script>

<span class="material-symbols-sharp" {...props}> {iconName} </span>

<style lang="scss">
	.material-symbols-sharp {
		transition: all 150ms linear;
		padding: 0.1em;
		margin: 0.2em;
		display: inline-block;
		border-radius: 100%;

		&:hover {
			background-color: var(--imag-blue-90);
		}

		&:active {
			background-color: var(--imag-blue-80);
		}
	}
</style>
