<script lang="ts">
	import type { HTMLButtonAttributes } from "svelte/elements";

	interface Props extends HTMLButtonAttributes {
		hoverColor?: string;
		variant?: "primary" | "small" | "mini";
	}

	let { children, hoverColor = "var(--imag-80)", variant = "primary", ...props }: Props = $props();
</script>

<button
	{...props}
	class="{variant} {props.class || ''}"
	aria-label={props["aria-label"] ?? props.title}
	style:--button-hover-bg={hoverColor}
>
	{@render children?.()}
</button>

<style lang="scss">
	button {
		cursor: pointer;
		color: var(--imag-text-colour);
		font-weight: 400;
		font-size: 1em; /* default */
		letter-spacing: 0.02em;
		height: max-content;
		background-color: var(--imag-90);
		border: none;
		padding: 0.5em 1em;
		display: inline-flex;
		align-items: center;
		flex-direction: row;
		text-align: center;
		position: relative;
		transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
		border-radius: 100px;

		&:hover {
			background-color: var(--button-hover-bg);
		}

		&.small {
			font-size: 0.85em;
		}

		&.mini {
			font-size: 0.75em;
		}
	}
</style>
