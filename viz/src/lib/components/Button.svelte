<script lang="ts">
	import type { HTMLButtonAttributes } from "svelte/elements";

	let { children, ...props }: HTMLButtonAttributes = $props();
</script>

<button {...props}>
	{@render children?.()}
</button>

<style lang="scss">
	button {
		cursor: pointer;
		color: var(--imag-text-colour);
		font-weight: 400;
		font-size: 1em;
		letter-spacing: 0.02em;
		height: max-content;
		background-color: var(--imag-90);
		border: none;
		margin: 0.5em 1em;
		padding: 0.8em 1em;
		display: inline-block;
		text-align: center;
		position: relative;
		transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
		border-radius: 100px;

		&:hover {
			background-color: var(--imag-100);
		}
	}
</style>
