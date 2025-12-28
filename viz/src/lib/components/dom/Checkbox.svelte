<script lang="ts">
	import { generateRandomString } from "$lib/utils/misc";
	import type { SvelteHTMLElements } from "svelte/elements";

	interface Props {
		checked: boolean;
		label?: string;
		id?: string;
		disabled?: boolean;
		onchange?: (e: Event & { currentTarget: HTMLInputElement }) => void;
	}

	let {
		checked = $bindable(),
		label,
		id,
		disabled = false,
		onchange,
		...props
	}: Props & SvelteHTMLElements["div"] = $props();

	const uniqueId = $derived(id || `checkbox-${generateRandomString(6)}`);

	function handleKeydown(event: KeyboardEvent) {
		if (disabled) {
			return;
		}

		if (event.key === "Enter") {
			event.preventDefault();
			checked = !checked;

			// Dispatch change event manually since programmatic change doesn't trigger it
			const changeEvent = new Event("change", { bubbles: true });
			const input = document.getElementById(uniqueId) as HTMLInputElement;
			if (input) {
				input.dispatchEvent(changeEvent);
			}
		}
	}
</script>

<div class="checkbox-wrapper" class:disabled {...props}>
	<input
		type="checkbox"
		id={uniqueId}
		bind:checked
		{disabled}
		{onchange}
		onkeydown={handleKeydown}
	/>
	<label for={uniqueId}>
		<span class="viz-checkbox" aria-hidden="true">
			<svg
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="3"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<polyline points="20 6 9 17 4 12"></polyline>
			</svg>
		</span>
		{#if label}
			<span class="label-text">{label}</span>
		{/if}
	</label>
</div>

<style lang="scss">
	.checkbox-wrapper {
		display: inline-flex;
		align-items: center;
		cursor: pointer;
		user-select: none;

		&.disabled {
			opacity: 0.5;
			pointer-events: none;
		}
	}

	input[type="checkbox"] {
		position: absolute;
		opacity: 0;
		cursor: pointer;
		height: 0;
		width: 0;
	}

	.viz-checkbox {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1rem;
		height: 1rem;
		border: 1.5px solid var(--imag-60);
		border-radius: 4px;
		background-color: transparent;
		transition: all 0.2s ease;
		color: transparent;
	}

	/* Checkmark size */
	svg {
		width: 0.75rem;
		height: 0.75rem;
	}

	/* Checked state */
	input[type="checkbox"]:checked + label .viz-checkbox {
		background-color: var(--imag-primary);
		border: 1.5px solid var(--imag-90);
		color: white;
	}

	/* Focus state */
	input[type="checkbox"]:focus + label .viz-checkbox {
		box-shadow: 0 0 0 3px rgba(var(--imag-primary), 0.2);
		border-color: var(--imag-primary);
	}

	label {
		display: flex;
		align-items: center;
		cursor: pointer;
		font-family: var(--imag-display-font);
		color: var(--imag-text-color);
		gap: 0.5rem;
	}

	.label-text {
		font-size: 1rem;
	}
</style>
