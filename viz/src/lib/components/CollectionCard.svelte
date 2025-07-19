<script lang="ts">
	import type CollectionData from "$lib/entities/collection";
	import { DateTime } from "luxon";

	interface Props {
		collection: CollectionData;
	}

	let { collection }: Props = $props();

	function openModal(url: string) {
		const linkDiv = document.createElement("a");
		linkDiv.href = url;
		linkDiv.target = "_blank";
		linkDiv.click();
		linkDiv.remove();
	}
</script>

<button
	type="button"
	class="coll-card"
	onclick={() => openModal(collection.thumbnail?.urls.original!)}
	data-collection-id={collection.id}
>
	<div class="image-container">
		<img src={collection.thumbnail?.urls.thumbnail} alt={collection.name} class="collection-image" />
	</div>
	<div class="metadata">
		<span class="coll-name" title={collection.name}>{collection.name}</span>
		<span class="coll-created_on">{DateTime.fromJSDate(collection.created_on).toFormat("dd.MM.yyyy")}</span>
		<span class="coll-image_count">{collection.image_count} {collection.image_count === 1 ? "image" : "images"}</span>
	</div>
</button>

<style lang="scss">
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

	.coll-image_count {
		margin-bottom: 0.5em;
	}

	.coll-card {
		min-width: 100%;
		max-width: 100%;
		height: auto;
		border: 1px solid var(--imag-20);
		border-radius: 0.5em;
		background-color: var(--imag-80);
		transition: background-color 0.1s ease;
		text-align: left;
		overflow: overlay;

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
		background-color: var(--imag-60);
	}

	.collection-image {
		width: 100%;
		height: 100%;
	}

	.metadata {
		display: flex;
		flex-direction: column;
		padding: 1em;
		max-height: 10em;
		overflow: hidden;
	}
</style>
