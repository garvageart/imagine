<script lang="ts">
	import type { SvelteHTMLElements } from "svelte/elements";
	interface Props {
		label?: string;
	}

	let { value = $bindable(), label, ...props }: Props & SvelteHTMLElements["input"] = $props();
</script>

<div class="input-container">
	<input
		{...props}
		type={props.type ?? "text"}
		placeholder={label ?? props.placeholder}
		bind:value
		oninput={(e) => {
			props.oninput?.(e);
		}}
		onchange={(e) => {
			props.onchange?.(e);
		}}
		onfocus={(e) => {
			props.onfocus?.(e);
		}}
		onblur={(e) => {
			props.onblur?.(e);
		}}
	/>
</div>

<style lang="scss">
	.input-container {
		display: flex;
		min-width: 0%;
		position: relative;
		width: 100%;
	}

	.input-label {
		font-size: 0.8em;
		position: absolute;
		left: 0.5em;
		bottom: 0.75em;
		background: var(--imag-bg-color);
		padding: 0.1em 0.5em;
		border-radius: 0.1em;
		font-weight: 600;
	}

	input:not([type="submit"]) {
		width: 100%;
		max-width: 100%;
		min-height: 2.5rem;
		color: var(--imag-text-color);
		background-color: var(--imag-bg-color);
		outline: none;
		border: none;
		box-shadow: 0 -1px 0 var(--imag-60) inset;
		font-family: var(--imag-font-family);
		font-size: 1rem;
		padding: 0.5rem 1rem;
		margin-bottom: 1rem;

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
		}

		&:focus::placeholder {
			color: var(--imag-60);
		}

		&:focus {
			box-shadow: 0 -2px 0 var(--imag-primary) inset;
		}

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
		}

		&:focus::placeholder {
			color: var(--imag-60);
		}

		&:focus {
			background-color: var(--imag-100);
			box-shadow: 0 -2px 0 var(--imag-primary) inset;
		}

		&:-webkit-autofill,
		&:-webkit-autofill:focus {
			-webkit-text-fill-color: var(--imag-text-color);
			-webkit-box-shadow: 0 0 0px 1000px var(--imag-100) inset;
			-webkit-box-shadow: 0 -5px 0 var(--imag-primary) inset;
			transition:
				background-color 0s 600000s,
				color 0s 600000s !important;
		}
	}
</style>
