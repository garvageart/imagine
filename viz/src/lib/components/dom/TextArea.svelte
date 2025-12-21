<script lang="ts">
	import { generateRandomString } from "$lib/utils/misc";
	import type { SvelteHTMLElements } from "svelte/elements";
	interface Props {
		label?: string;
		description?: string;
		disabled?: boolean;
	}

	let {
		value = $bindable(),
		label,
		description,
		disabled = false,
		...props
	}: Props & SvelteHTMLElements["textarea"] = $props();
	const inputId = props.id ?? generateRandomString(6);
</script>

<div class="textarea-container" class:disabled>
	{#if label}
		<label for={inputId} class="textarea-label"
			>{label}
			{#if props.required}
				<span class="required-asterisk">*</span>
			{/if}
		</label>
	{/if}
	<textarea
		{...props}
		id={inputId}
		name={props.name}
		placeholder={props.placeholder}
		bind:value
		{disabled}
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
	></textarea>
	{#if description}
		<div class="textarea-description">{description}</div>
	{/if}
</div>

<style lang="scss">
	.textarea-container {
		display: flex;
		flex-direction: column;
		min-width: 0%;
		position: relative;
		width: 100%;
		gap: 0.5rem;

		&.disabled {
			opacity: 0.5;

			textarea {
				cursor: not-allowed;
			}
		}
	}

	.textarea-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--imag-40);
	}

	.textarea-description {
		font-size: 0.85rem;
		color: var(--imag-60);
		padding-left: 0.5rem;
	}

	textarea {
		max-width: 100%;
		min-height: 2.5rem;
		color: var(--imag-text-color);
		background-color: var(--imag-100);
		outline: none;
		border: none;
		box-shadow: 0 -1px 0 var(--imag-60) inset;
		font-family: var(--imag-display-font);
		font-size: 1rem;
		padding: 0.5rem 1rem;
		margin-bottom: 0;

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-display-font);
		}

		&:focus::placeholder {
			color: var(--imag-60);
			opacity: 1;
		}

		&:focus {
			box-shadow: 0 -2px 0 var(--imag-primary) inset;
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
