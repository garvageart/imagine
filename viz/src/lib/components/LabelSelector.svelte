<script lang="ts">
	import { LabelColours } from "$lib/images/constants";
	import { generateRandomString } from "$lib/utils/misc";

	interface Props {
		label: LabelColours | null;
		onSelect?: (label: LabelColours | null) => void;
		variant?: "expanded" | "compact";
		enableSelection?: boolean;
	}

	let {
		label,
		onSelect,
		variant = "expanded",
		enableSelection = true
	}: Props = $props();
</script>

<div class="label-selector-container">
	{#if variant === "expanded"}
		{#each Object.entries(LabelColours) as [name, colour]}
			<button
				class="label-selector"
				title={name}
				style="background-color: {colour};"
				class:selected={label === colour}
				disabled={!enableSelection}
				onclick={() => onSelect?.(colour as LabelColours)}
			>
			</button>
		{/each}
	{:else if variant === "compact"}
		<select
			id="label-selector-{generateRandomString(6)}"
			class="label-selector-select"
			class:disable-select={!enableSelection}
			style="background-color: {label ?? 'var(--imag-bg-color)'};"
			bind:value={label}
			disabled={!enableSelection}
			onchange={(e) => onSelect?.(e.currentTarget.value as LabelColours | null)}
		>
			{#each Object.entries(LabelColours) as [name, colour]}
				<option title={name} value={colour} style="background-color: {colour};"
				></option>
			{/each}
		</select>
	{/if}
</div>

<style lang="scss">
	.label-selector-container {
		display: flex;
		gap: 0.5em;
	}

	.label-selector {
		display: inline-flex;
		align-items: center;
		gap: 0.5em;
		width: 1.25em;
		height: 1.25em;
		border: 1px solid var(--imag-60);
		background-color: var(--imag-bg-color);
		color: var(--imag-text-color);
		cursor: pointer;
		font-size: 0.9em;

		&.selected {
			background-color: var(--imag-primary);
			border: 2px solid var(--imag-primary);
		}
	}

	.disable-select {
		cursor: not-allowed;
		pointer-events: none;
		appearance: none;
		-webkit-appearance: none;
		-moz-appearance: none;
		background-image: none;
	}

	.label-selector-select {
		height: 0.75rem;
		width: 0.75rem;
		outline: none;
		border: 1px solid var(--imag-60);
		background-color: var(--imag-bg-color);
		color: var(--imag-text-color);
		font-family: inherit;
		cursor: pointer;
		font-size: 1em;

		&:focus {
			outline: 1.5px solid var(--imag-primary);
			border-color: var(--imag-primary);
		}

		option {
			padding: 0.5em;

			&::selection {
				border-color: var(--imag-primary);
				background-color: inherit;
				color: inherit;
			}

			&:hover {
				border-color: var(--imag-primary);
				background-color: inherit;
			}
		}
	}
</style>
