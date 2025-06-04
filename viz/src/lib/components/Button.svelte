<script lang="ts">
	import type { HTMLButtonAttributes } from "svelte/elements";

	let { children, ...props }: HTMLButtonAttributes = $props();
</script>

<button {...props}>
	{@render children?.()}
</button>

<style lang="scss">
	button {
		color: var(--imag-text-colour);
		font-weight: 500;
		font-size: 1em;
		background-color: var(--imag-button-colour);
		border: none;
		margin: 0.5em;
		padding: 0.7em 2em;
		display: inline-block;
		text-align: center;
		position: relative;
		transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
		cursor: pointer;

		&:hover {
			background-color: hsl(209, 50%, 39%);
		}
	}
</style>
