<script lang="ts">
	import { LabelColours } from "$lib/images/constants";

	interface Props {
		label: LabelColours | null;
		onSelect?: (label: LabelColours | null) => void;
		variant?: "expanded" | "compact";
		enableSelection?: boolean;
	}

	let {
		label = $bindable(),
		onSelect,
		variant = "expanded",
		enableSelection = true
	}: Props = $props();

	let isOpen = $state(false);
	let dropdownContainer = $state<HTMLElement>();

	const nullNoneLabelStyle = $derived(
		`background-color: transparent;${enableSelection === false ? " border: none;" : ""}`
	);

	function handleSelect(newLabel: LabelColours | null) {
		if (!enableSelection) {
			return;
		}
		label = newLabel;
		onSelect?.(newLabel);
	}

	function handleClickOutside(event: MouseEvent) {
		if (
			isOpen &&
			dropdownContainer &&
			!dropdownContainer.contains(event.target as Node)
		) {
			isOpen = false;
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

{#if variant === "expanded"}
	<div class="label-selector-container">
		{#each Object.entries(LabelColours).filter(([_, colour]) => colour !== LabelColours.None) as [name, colour]}
			<button
				class="label-selector"
				title={name}
				style="background-color: {colour};"
				class:selected={label === colour}
				disabled={!enableSelection}
				onclick={() => {
					if (label === colour) {
						handleSelect(LabelColours.None);
					} else {
						handleSelect(colour as LabelColours);
					}
				}}
				type="button"
			>
			</button>
		{/each}
	</div>
{:else if variant === "compact"}
	{#if enableSelection || label !== null || (label && label !== LabelColours.None)}
		<div class="label-selector-container" class:compact={variant === "compact"}>
			<div class="compact-container" bind:this={dropdownContainer}>
				<button
					class="label-selector-trigger"
					class:disable-select={!enableSelection}
					style={!label || label === LabelColours.None
						? nullNoneLabelStyle
						: `background-color: ${label};`}
					onclick={() => {
						if (enableSelection) {
							isOpen = !isOpen;
						}
					}}
					disabled={!enableSelection}
					title="Select label"
					type="button"
				>
				</button>
				{#if isOpen}
					<div class="label-dropdown">
						{#each Object.entries(LabelColours).filter(([_, colour]) => colour !== LabelColours.None) as [name, colour]}
							<button
								class="label-selector"
								title={name}
								style="background-color: {colour};"
								class:selected={label === colour}
								onclick={() => {
									if (label === colour) {
										handleSelect(LabelColours.None);
									} else {
										handleSelect(colour as LabelColours);
									}
									isOpen = false;
								}}
								type="button"
							>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
{/if}

<style lang="scss">
	.compact-container {
		position: relative;
	}

	.label-selector-container {
		display: flex;
		gap: 0.5em;

		&.compact {
			gap: 0em;
		}
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
		padding: 0;

		&:disabled {
			opacity: 1;
		}

		&.selected {
			outline: 1.5px solid var(--imag-primary);
			outline-offset: 0.5px;
		}
	}

	.disable-select {
		cursor: not-allowed;
		pointer-events: none;
	}

	.label-selector-trigger {
		height: 0.75rem;
		width: 0.75rem;
		outline: none;
		border: 1px solid var(--imag-60);
		background-color: var(--imag-bg-color);
		cursor: pointer;
		padding: 0;

		&:disabled {
			opacity: 1;
		}

		&:focus-visible {
			outline: 1.5px solid var(--imag-primary);
			border-color: var(--imag-primary);
		}
	}

	.label-dropdown {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		z-index: 1000;
		background-color: var(--imag-bg-color);
		border: 1px solid var(--imag-60);
		padding: 0.5rem;
		display: flex;
		gap: 0.5rem;
		width: max-content;
		box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
		border-radius: 0.25rem;
	}
</style>
