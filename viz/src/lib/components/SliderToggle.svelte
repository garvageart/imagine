<!-- Modified from here: https://svelte.dev/playground/d65a4e9f0ae74d1eb1b08d13e428af32?version=5.36.8 -->
<script lang="ts">
	import { generateRandomString } from "$lib/utils/misc";
	import type { MouseEventHandler, SvelteHTMLElements } from "svelte/elements";

	interface Props {
		label: string;
		value: "on" | "off";
		labelPos?: "side" | "top";
	}

	let { label, value = $bindable(), labelPos = "side", ...props }: Props & SvelteHTMLElements["div"] = $props();

	let checked = $state(true);
	const uniqueID = generateRandomString(6);

	$effect(() => {
		checked = value === "on" ? true : false;
	});

	const handleClick: MouseEventHandler<HTMLButtonElement> = (e) => {
		e.preventDefault();
		const target = e.currentTarget;
		const state = target.getAttribute("aria-checked");

		checked = state === "true" ? false : true;
		value = checked === true ? "on" : "off";
	};
</script>

<div {...props} class="toggle-slider {labelPos === 'side' ? 'side' : 'top'}">
	<label for="switch-{uniqueID}" id={`switch-${uniqueID}-label`}>{label}</label>
	<button
		id="switch-{uniqueID}"
		style={labelPos === "side" ? "margin-left: 0.5em;" : ""}
		role="switch"
		aria-checked={checked}
		data-checked={checked}
		aria-labelledby={`switch-${uniqueID}`}
		onclick={handleClick}
	>
	</button>
</div>

<style lang="scss">
	:root {
		--gray: #ccc;
	}

	:global([role="switch"][aria-checked="true"] :first-child),
	:global([role="switch"][aria-checked="false"] :last-child) {
		display: none;
		color: #fff;
	}

	.toggle-slider {
		display: flex;
		align-items: center;

		&.side {
			flex-direction: row;
			align-items: center;
			justify-content: space-between;
		}

		&.top {
			flex-direction: column;
			align-items: flex-start;
			justify-content: space-between;
			height: 3.2em;
		}
	}

	.toggle-slider button {
		width: 3em;
		height: 1.6em;
		position: relative;
		background: var(--gray);
		border: none;
	}

	.toggle-slider button::before {
		content: "";
		position: absolute;
		width: 1.3em;
		height: 1.3em;
		background: #fff;
		top: 0.13em;
		right: 1.5em;
		transition: transform 0.3s;
	}

	.toggle-slider button[aria-checked="true"] {
		background-color: var(--imag-primary);
	}

	.toggle-slider button[aria-checked="true"]::before {
		transform: translateX(1.3em);
		transition: transform 0.3s;
	}

	.toggle-slider button:focus {
		box-shadow: 0 0px 0px 1px var(--imag-primary);
	}

	.toggle-slider button {
		border-radius: 1.5em;
	}

	.toggle-slider button::before {
		border-radius: 100%;
	}

	.toggle-slider button:focus {
		box-shadow: 0 0px 8px var(--imag-primary);
		border-radius: 1.5em;
	}
</style>
