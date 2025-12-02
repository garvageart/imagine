<script lang="ts">
	interface Props {
		label: string;
		value?: string;
		options?: string[];
		description?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	}

	let { label, value = $bindable(""), options = [], description = "", disabled = false, onchange }: Props = $props();

	function handleChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		value = target.value;
		if (onchange) {
			onchange(value);
		}
	}
</script>

<div class="select-container" class:disabled>
	<div class="label-group">
		<label for="select-{label}" class="label">{label}</label>
		{#if description}
			<span class="description">{description}</span>
		{/if}
	</div>
	<select id="select-{label}" bind:value onchange={handleChange} {disabled} class="select-input">
		{#each options as option}
			<option value={option}>{option}</option>
		{/each}
	</select>
</div>

<style lang="scss">
	.select-container {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 0;
		border-bottom: 1px solid var(--imag-80);
		width: 100%;

		&.disabled {
			opacity: 0.5;
		}
	}

	.label-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.label {
		font-weight: 500;
		color: var(--imag-text-color);
	}

	.description {
		font-size: 0.875rem;
		color: var(--imag-text-color);
	}

	.select-input {
		padding: 0.5rem 2rem 0.5rem 1rem;
		border-radius: 0.375rem;
		background-color: var(--imag-100);
		color: var(--imag-text-color);
		border: 1px solid var(--imag-80);
		outline: none;
		cursor: pointer;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%239ca3af'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.5rem center;
		background-size: 1rem;
		font-family: var(--imag-font-family);

		&:focus {
			border-color: var(--imag-70);
		}

		&:disabled {
			cursor: not-allowed;
		}
	}
</style>
