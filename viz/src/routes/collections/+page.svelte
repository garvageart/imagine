<script lang="ts">
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import { DateTime } from "luxon";
	import type { PageProps } from "./$types";

	let { data }: PageProps = $props();
	function openModal(collectionId: string) {
		// window.location.href = `/collections/${collectionId}`;
		console.log("Opening modal for collection ID:", collectionId);
	}
</script>

<VizViewContainer name="Collections">
	<div id="viz-card-container">
		{#each data.response as collection}
			<button type="button" class="coll-card" onclick={() => openModal(collection.id)} data-collection-id={collection.id}>
				<div class="image-container"></div>
				<div class="metadata">
					<span class="coll-name" title={collection.name}>{collection.name}</span>
					<span class="coll-description">{collection.description}</span>
					<span class="coll-created_on">{DateTime.fromJSDate(collection.created_on).toFormat("dd-MM-yyyy")}</span>
				</div>
			</button>
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

	.coll-name {
		font-size: 1em;
		font-weight: bold;
		font-family: var(--imag-font-family);
		color: var(--imag-text-color);
		border: none;
		outline: none;
		padding: 0.2em 0em;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.coll-card {
		min-width: 100%;
		max-width: 100%;
		height: auto;
		border: 1px solid var(--imag-20);
		background-color: var(--imag-60);
		transition: background-color 0.1s ease;
		text-align: left;

		&:hover {
			background-color: var(--imag-70);
			cursor: pointer;
		}

		&:active {
			outline: 2px solid var(--imag-80);
			outline-offset: -2px;
		}
	}

	.image-container {
		height: 13em;
		background-color: var(--imag-80);
	}

	.metadata {
		display: flex;
		flex-direction: column;
		padding: 1em;
		height: 10em;
		max-height: 10em;
		overflow: hidden;
	}
</style>
