<script lang="ts">
	interface Props {
		label: string;
		value?: string;
		description?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	}

	let { label, value = $bindable(""), description = "", disabled = false, onchange }: Props = $props();

	let error = $state<string | null>(null);

	function handleInput(event: Event) {
		const target = event.target as HTMLTextAreaElement;
		const newValue = target.value;

		try {
			JSON.parse(newValue);
			error = null;
			value = newValue;
			if (onchange) {
				onchange(newValue);
			}
		} catch (e) {
			error = "Invalid JSON format";
		}
	}
</script>

<div class="json-container" class:disabled>
	<div class="header">
		<div class="label-group">
			<label for="json-{label}" class="label">{label}</label>
			{#if description}
				<span class="description">{description}</span>
			{/if}
		</div>
		<!-- I hate this btw, must look better -->
		{#if error}
			<span class="error-msg">{error}</span>
		{/if}
	</div>

	<textarea id="json-{label}" {value} oninput={handleInput} {disabled} class="json-input" class:error={!!error} rows="5"
	></textarea>
</div>

<style lang="scss">
	.json-container {
		display: flex;
		flex-direction: column;
		padding: 1rem 0;
		border-bottom: 1px solid var(--imag-80);
		width: 100%;
		gap: 0.5rem;

		&.disabled {
			opacity: 0.5;
		}
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
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
		color: var(--imag-40);
	}

	.error-msg {
		color: var(--imag-error-color, #ef4444);
		font-size: 0.875rem;
	}

	.json-input {
		width: 100%;
		padding: 0.5rem;
		border-radius: 0.375rem;
		background-color: var(--imag-100);
		color: var(--imag-text-color);
		border: 1px solid var(--imag-80);
		outline: none;
		font-family: var(--imag-code-font);
		resize: vertical;

		&:focus {
			border-color: var(--imag-70);
		}

		&.error {
			border-color: var(--imag-error-color, #ef4444);
		}

		&:disabled {
			cursor: not-allowed;
		}
	}
</style>
