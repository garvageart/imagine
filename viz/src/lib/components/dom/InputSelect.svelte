<script lang="ts">
	import { generateKeyId } from "$lib/utils/layout";
	import type { HTMLSelectAttributes } from "svelte/elements";

	interface Props extends HTMLSelectAttributes {
		label?: string;
		labelPosition?: "top" | "side";
		description?: string;
		children?: import("svelte").Snippet;
		value?: string | number | string[];
	}

	let {
		label,
		labelPosition = "top",
		description,
		children,
		value = $bindable(),
		class: className,
		...props
	}: Props = $props();
	const selectId = props.id ?? generateKeyId();
</script>

<div
	class="input-container"
	class:disabled={props.disabled}
	class:side-label={labelPosition === "side"}
>
	{#if label}
		<label for={selectId} class="input-label">{label}</label>
	{/if}
	<div class="input-wrapper">
		<select
			id={selectId}
			bind:value
			class="select-input {className || ''}"
			{...props}
		>
			{@render children?.()}
		</select>
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
		gap: 0.25rem;

		&.side-label {
			flex-direction: row;
			align-items: center;

			.input-label {
				margin-right: 0.5rem;
				white-space: nowrap;
			}
		}

		&.disabled {
			opacity: 0.5;

			select {
				cursor: not-allowed;
			}
		}
	}

	.input-wrapper {
		position: relative;
		width: 100%;
	}

	.input-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--imag-40);
	}

	.input-description {
		font-size: 0.85rem;
		color: var(--imag-text-secondary, #888);
		padding-left: 0.5rem;
	}

	.select-input {
		width: 100%;
		max-width: 100%;
		min-height: 2rem;
		color: var(--imag-text-color);
		background-color: var(--imag-100);
		outline: none;
		border: none;
		box-shadow: 0 -1px 0 var(--imag-60) inset;
		font-family: var(--imag-display-font);
		font-size: 1rem;
		padding: 0.5rem 1rem;
		cursor: pointer;

		/* Custom styling to match InputText */
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%239ca3af'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.5rem center;
		background-size: 1rem;

		&:focus {
			background-color: var(--imag-100);
			box-shadow: 0 -2px 0 var(--imag-primary) inset;
		}
	}
</style>
