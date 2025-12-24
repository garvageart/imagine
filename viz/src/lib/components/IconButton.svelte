<script lang="ts">
	import type { HTMLButtonAttributes } from "svelte/elements";
	import MaterialIcon, { type IconProps } from "./MaterialIcon.svelte";

	interface ButtonProps extends HTMLButtonAttributes {
		hoverColor?: string;
		variant?: "primary" | "small" | "mini";
		element?: HTMLButtonElement;
	}

	type Props = ButtonProps & IconProps;

	let {
		iconName = $bindable(),
		iconStyle = "sharp",
		fill = $bindable(false),
		weight = 400,
		grade = 0,
		opticalSize = 24,
		children,
		hoverColor = "var(--imag-80)",
		variant = "primary",
		element = $bindable(),
		...props
	}: Props = $props();
</script>

<button
	{...props}
	bind:this={element}
	class="{variant} {props.class || ''}"
	aria-label={props["aria-label"] ?? props.title}
	style:--button-hover-bg={hoverColor}
>
	<MaterialIcon {iconName} {iconStyle} {fill} {grade} {opticalSize} />
	{@render children?.()}
</button>

<style lang="scss">
	button {
		cursor: pointer;
		color: var(--imag-text-color);
		font-weight: 400;
		letter-spacing: 0.02em;
		height: max-content;
		border: none;
		padding: 0.25em;
		gap: 0.25em;
		display: inline-flex;
		align-items: center;
		flex-direction: row;
		text-align: center;
		transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
		border-radius: 100px;

		&:hover {
			background-color: var(--button-hover-bg);
		}

		&:disabled {
			cursor: not-allowed;
			opacity: 0.5;
			background-color: transparent;
		}

		&.small {
			font-size: 0.85em;
		}

		&.mini {
			font-size: 0.75em;
		}
	}
</style>
