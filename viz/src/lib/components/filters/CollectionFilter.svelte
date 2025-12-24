<script lang="ts">
	import type {
		CollectionFilters,
		CollectionFacets
	} from "$lib/states/filter.svelte";

	interface Props {
		criteria: CollectionFilters;
		facets: CollectionFacets;
		uiState: { expanded: Record<string, boolean> };
		save: () => void;
	}

	let {
		criteria = $bindable(),
		facets,
		uiState = $bindable(),
		save
	}: Props = $props();
</script>

<div class="collection-filter-container">
	<span class="placeholder-text"
		>No collection-specific filters are available at this time.</span
	>
</div>

<style>
	.collection-filter-container {
		padding: 1rem;
		color: var(--imag-60);
		font-style: italic;
		text-wrap: auto;
		text-align: center;
	}
</style>
