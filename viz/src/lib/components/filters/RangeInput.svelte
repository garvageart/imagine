<script lang="ts">
	import InputText from "../dom/InputText.svelte";

	interface Props {
		label: string;
		min: number;
		max: number;
		value: { min?: number; max?: number };
		step?: number;
		onChange: (value: { min?: number; max?: number }) => void;
		unit?: string;
	}

	let {
		label,
		min,
		max,
		value,
		step = 1,
		onChange,
		unit = ""
	}: Props = $props();

	function handleMinChange(e: Event) {
		const val = parseFloat((e.target as HTMLInputElement).value);
		if (!isNaN(val)) {
			onChange({ ...value, min: val });
		} else {
			onChange({ ...value, min: undefined });
		}
	}

	function handleMaxChange(e: Event) {
		const val = parseFloat((e.target as HTMLInputElement).value);
		if (!isNaN(val)) {
			onChange({ ...value, max: val });
		} else {
			onChange({ ...value, max: undefined });
		}
	}
</script>

<div class="range-container">
	<div class="header">
		<span class="label">{label}</span>
		<span class="range-display">
			{min} - {max}{unit}
		</span>
	</div>
	<div class="inputs">
		<div class="input-wrapper">
			<InputText
				type="number"
				placeholder="Min"
				{min}
				{max}
				{step}
				value={value.min !== undefined ? value.min : ""}
				onchange={handleMinChange}
			/>
		</div>
		<span class="separator">-</span>
		<div class="input-wrapper">
			<InputText
				type="number"
				placeholder="Max"
				{min}
				{max}
				{step}
				value={value.max !== undefined ? value.max : ""}
				onchange={handleMaxChange}
			/>
		</div>
	</div>
</div>

<style lang="scss">
	.range-container {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		margin-bottom: 0.5rem;
	}

	.header {
		display: flex;
		justify-content: space-between;
		font-size: 0.8rem;
		color: var(--imag-60);
	}

	.label {
		font-weight: 500;
		color: var(--imag-text-color);
	}

	.range-display {
		font-size: 0.75rem;
	}

	.inputs {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.input-wrapper {
		flex: 1;
		/* Force inputs to be smaller */
		:global(input) {
			min-height: 2rem;
			padding: 4px 8px;
			font-size: 0.9rem;
		}
	}

	.separator {
		color: var(--imag-60);
	}
</style>
