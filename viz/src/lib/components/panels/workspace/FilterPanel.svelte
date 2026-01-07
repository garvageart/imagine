<script lang="ts">
	import { filterManager } from "$lib/states/filter.svelte";
	import IconButton from "$lib/components/IconButton.svelte";
	import ImageFilter from "$lib/components/filters/ImageFilter.svelte";
	import CollectionFilter from "$lib/components/filters/CollectionFilter.svelte";

	const imageScope = filterManager.getScope("images");
	const collectionScope = filterManager.getScope("collections");
</script>

<div class="filter-panel">
	{#if filterManager.activeScopeType === "collections" && collectionScope}
		<CollectionFilter
			bind:criteria={collectionScope.criteria}
			facets={collectionScope.facets}
			bind:uiState={collectionScope.uiState}
			save={() => filterManager.save()}
		/>
	{:else if imageScope}
		<ImageFilter
			bind:criteria={imageScope.criteria}
			facets={imageScope.facets}
			bind:uiState={imageScope.uiState}
			save={() => filterManager.save()}
		/>
	{/if}

	<div class="filter-actions">
		<IconButton
			iconName={filterManager.keepFilters ? "keep" : "keep_off"}
			variant="mini"
			title="Keep Filters while browsing"
			style={filterManager.keepFilters
				? "background-color: var(--imag-80);"
				: ""}
			onclick={() => filterManager.toggleKeepFilters()}
		/>
		<IconButton
			iconName="layers_clear"
			variant="mini"
			onclick={() => filterManager.resetActiveScope(true)}
			title="Clear all active filters"
		/>
	</div>
</div>

<style lang="scss">
	.filter-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow-y: scroll;
		overflow-x: hidden;
		padding: 0.5rem;
		color: var(--imag-text-color);
		position: relative;
	}

	.filter-actions {
		position: absolute;
		bottom: 0;
		display: flex;
		justify-content: flex-start;
		align-items: center;
		padding: 0.2rem 0;
		background-color: var(--imag-bg-color);
		gap: 0.5rem;
		width: 100%;
	}
</style>
