<script lang="ts">
	interface Props {
		label: string;
		value?: string | number;
		type?: "text" | "number";
		description?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	}

	let { label, value = $bindable(""), type = "text", description = "", disabled = false, onchange }: Props = $props();

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		value = target.value;
		if (onchange) {
			onchange(target.value);
		}
	}
</script>

<div class="input-container" class:disabled>
	<div class="label-group">
		<label for="input-{label}" class="label">{label}</label>
		{#if description}
			<span class="description">{description}</span>
		{/if}
	</div>
	<input id="input-{label}" {type} bind:value oninput={handleInput} {disabled} class="text-input" />
</div>

<style lang="scss">
	.input-container {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 0;
		border-bottom: 1px solid var(--imag-80);
		width: 100%;
		font-family: var(--imag-font-family);

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

	.text-input {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		background-color: var(--imag-100);
		color: var(--imag-text-color);
		border: 1px solid var(--imag-80);
		outline: none;
		font-family: var(--imag-font-family);
		width: 200px;

		&:focus {
			border-color: var(--imag-70);
		}

		&:disabled {
			cursor: not-allowed;
		}
	}
</style>
