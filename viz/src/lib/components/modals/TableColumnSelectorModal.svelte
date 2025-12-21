<script lang="ts">
	import ModalOverlay from "./ModalContainer.svelte";
	import CheckboxGroup from "../dom/CheckboxGroup.svelte";
	import { tableColumnSettings } from "$lib/states/index.svelte";
	import { snakeToSentence } from "$lib/utils/strings";

	interface Props {
		availableKeys: string[];
	}

	let { availableKeys }: Props = $props();

	let options = $derived(
		availableKeys.map((key) => ({
			label: snakeToSentence(key),
			value: key
		}))
	);

	let selectedColumns = $state([...tableColumnSettings.value]);

	$effect(() => {
		if (
			JSON.stringify(selectedColumns) !==
			JSON.stringify(tableColumnSettings.value)
		) {
			tableColumnSettings.set(selectedColumns);
		}
	});
</script>

<ModalOverlay>
	<div id="viz-column-selector-modal">
		<h1>Table Columns</h1>
		<p class="subtitle">Select which columns to display in the list view</p>

		<div class="column-list">
			<CheckboxGroup {options} bind:value={selectedColumns} />
		</div>
	</div>
</ModalOverlay>

<style lang="scss">
	#viz-column-selector-modal {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		align-items: center;
		padding: 2rem;
		color: var(--imag-text-color);
	}

	h1 {
		font-family: var(--imag-display-font);
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
	}

	.subtitle {
		color: var(--imag-60);
		margin-bottom: 2rem;
		font-size: 1.1rem;
	}

	.column-list {
		width: 100%;
		max-width: 600px;
		overflow-y: auto;
		padding: 1rem;
		background: var(--imag-bg-color);
		border-radius: 1rem;
		box-shadow: inset 0 0 10px rgba(0, 0, 0, 0.5);

		/* Custom scrollbar matching project style */
		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: var(--imag-100);
		}
		&::-webkit-scrollbar-thumb {
			background: var(--imag-60);
			border-radius: 3px;
		}

		/* Styling the CheckboxGroup to match the previous grid layout */
		:global(.checkbox-group) {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
			gap: 1rem;
		}

		:global(.checkbox-group-item) {
			background: var(--imag-100);
			padding: 0.75rem 1rem;
			border-radius: 0.5rem;
			border: 1px solid var(--imag-90);
			transition: all 0.2s;

			&:hover {
				border-color: var(--imag-primary);
				background: var(--imag-90);
			}
		}
	}
</style>
