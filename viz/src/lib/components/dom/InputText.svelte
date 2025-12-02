<script lang="ts">
	import type { SvelteHTMLElements } from "svelte/elements";
	interface Props {
		label?: string;
		description?: string;
		disabled?: boolean;
	}

	let { value = $bindable(), label, description, disabled = false, ...props }: Props & SvelteHTMLElements["input"] = $props();
</script>

<div class="input-container" class:disabled>
	<div class="input-wrapper">
		<input
			{...props}
			type={props.type ?? "text"}
			placeholder={label ?? props.placeholder}
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
		/>
		{#if label}
			<span class="input-label">{label}</span>
		{/if}
	</div>
	{#if description}
		<div class="input-description">{description}</div>
	{/if}
</div>

<style lang="scss">
	.input-container {
		display: flex;
		flex-direction: column;
		min-width: 0%;
		position: relative;
		width: 100%;
		gap: 0.25rem;

		&.disabled {
			opacity: 0.5;
			
			input {
				cursor: not-allowed;
			}
		}
	}
	
	.input-wrapper {
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
		pointer-events: none;
	}
	
	.input-description {
		font-size: 0.85rem;
		color: var(--imag-text-secondary, #888);
		padding-left: 0.5rem;
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
		margin-bottom: 0; /* Changed from 1rem */

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
			opacity: 0; /* Hide placeholder when label is used overlay style */
		}
		
		/* Show placeholder if no label is present */
		&:placeholder-shown:not(:focus) + .input-label {
			/* This selector logic depends on structure, usually floating labels use :placeholder-shown */
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
