<script lang="ts">
	import { onMount } from "svelte";
	import type { HTMLButtonAttributes } from "svelte/elements";

	let { children, hoverColor = "var(--imag-90)", ...props }: { hoverColor?: string } & HTMLButtonAttributes = $props();
	let el: HTMLButtonElement | undefined = $state();
	let defaultBackgroundColour: string | undefined = $state();

	onMount(() => {
		defaultBackgroundColour = el?.computedStyleMap().get("background-color")?.toString();

		// I hate this, why am I this person
		if (defaultBackgroundColour === hoverColor) {
			const colourStepValue = 10;
			const currentColour = parseInt(defaultBackgroundColour.split("var(--imag-")[1].split(")")[0], 10);
			if (currentColour <= 40) {
				hoverColor = `var(--imag-${currentColour + colourStepValue})`;
			} else {
				hoverColor = `var(--imag-${currentColour - colourStepValue})`;
			}
		}
	});
</script>

<button
	bind:this={el}
	{...props}
	onmouseenter={(e) => {
		props.onmouseenter?.(e);
		e.currentTarget.style.setProperty("background-color", hoverColor);
	}}
	onmouseleave={(e) => {
		props.onmouseleave?.(e);
		e.currentTarget.style.setProperty("background-color", defaultBackgroundColour!);
	}}
	aria-label={props["aria-label"] ?? props.title}
>
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
		padding: 0.5em 1em;
		display: inline-flex;
		align-items: center;
		flex-direction: row;
		text-align: center;
		position: relative;
		transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
		border-radius: 100px;
	}
</style>
