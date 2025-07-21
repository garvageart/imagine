<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import type CollectionData from "$lib/entities/collection";
	import { addViewToContent } from "$lib/utils/layout";
	import VizView from "$lib/views/views.svelte";
	import { DateTime } from "luxon";
	import CollectionPage from "../../routes/collections/[id]/+page.svelte";
	import { getContext } from "svelte";
	import { Content } from "$lib/layouts/subpanel.svelte";
	import { layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { findPanelIndex, getSubPanelParent } from "$lib/views/utils";

	interface Props {
		collection: CollectionData;
	}

	let { collection }: Props = $props();

	const currentContent = getContext<Content>("content");
	function openCollection() {
		if (page.url.pathname !== "/") {
			goto(`/collections/${collection.id}`, { state: { from: page.url.pathname } });
			return;
		}

		const currentParentIdx = findPanelIndex(layoutState.tree, getSubPanelParent(layoutState.tree, currentContent.paneKeyId)!);
		if (currentParentIdx === -1) {
			console.warn("Can't find panel in layout, navigating to collection page");
			goto(`/collections/${collection.id}`, { state: { from: page.url.pathname } });
			return;
		}

		const currentParent = layoutState.tree[currentParentIdx];
		const childIndex = currentParent.childs.content.findIndex((subPanel) => subPanel.paneKeyId === currentContent.paneKeyId);

		if (childIndex === -1) {
			console.warn(`Can't find child inside panel ${currentParent.paneKeyId}, navigating to collection page`);
			goto(`/collections/${collection.id}`, { state: { from: page.url.pathname } });
			return;
		}

		const view = new VizView({
			name: collection.name,
			component: CollectionPage as any,
			path: `/collections/${collection.id}`
		});

		addViewToContent(view, currentParentIdx, childIndex);
		currentParent.makeViewActive(view);
	}
</script>

<button type="button" class="coll-card" onclick={() => openCollection()} data-collection-id={collection.id}>
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
		border-radius: 0.5em;
		background-color: var(--imag-90);
		transition: background-color 0.1s ease;
		text-align: left;
		overflow: overlay;
		box-shadow: 0 4px 0 var(--imag-100);

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
