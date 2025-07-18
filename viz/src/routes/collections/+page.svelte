<script lang="ts">
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import type { PageProps } from "./$types";
	import CollectionCard from "$lib/components/CollectionCard.svelte";

	let { data }: PageProps = $props();
</script>

<VizViewContainer name="Collections">
	<div id="viz-card-container">
		{#each data?.response as collection}
			<CollectionCard {collection} />
		{/each}
	</div>
</VizViewContainer>

<style lang="scss">
	#viz-card-container {
		margin: 1em 0em;
		display: grid;
		gap: 1em;
		width: 100%;
		text-overflow: clip;
		justify-content: center;
		grid-template-columns: repeat(auto-fit, minmax(15em, 1fr));
	}
</style>
