<script lang="ts" module>
	export function openCollection(collection: CollectionData, currentContent: Content) {
		if (page.url.pathname !== "/") {
			goto(`/collections/${collection.id}`, { state: { from: page.url.pathname, data: collection } });
			return;
		}

		const currentParentIdx = findPanelIndex(layoutState.tree, getSubPanelParent(layoutState.tree, currentContent.paneKeyId)!);
		if (currentParentIdx === -1) {
			console.warn("Can't find panel in layout, navigating to collection page");
			goto(`/collections/${collection.id}`, { state: { from: page.url.pathname, data: collection } });
			return;
		}

		const currentParent = layoutState.tree[currentParentIdx];
		const childIndex = currentParent.childs.content.findIndex((subPanel) => subPanel.paneKeyId === currentContent.paneKeyId);

		if (childIndex === -1) {
			console.warn(`Can't find child inside panel ${currentParent.paneKeyId}, navigating to collection page`);
			goto(`/collections/${collection.id}`, { state: { from: page.url.pathname, data: collection } });
			return;
		}

		const existingView = layoutState.tree
			.flatMap((panel) => panel.childs.content)
			.find((subPanel) => subPanel.views.some((view) => view.path === `/collections/${collection.id}`));

		if (existingView) {
			const parent = findSubPanel("paneKeyId", existingView.paneKeyId)?.subPanel as VizSubPanelData | undefined;
			if (parent) {
				parent.makeViewActive(existingView.views.find((view) => view.path === `/collections/${collection.id}`)!);
			}
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

<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import type CollectionData from "$lib/entities/collection";
	import { addViewToContent, findSubPanel } from "$lib/utils/layout";
	import VizView from "$lib/views/views.svelte";
	import { DateTime } from "luxon";
	import CollectionPage from "../../routes/(app)/collections/[id]/+page.svelte";
	import VizSubPanelData, { Content } from "$lib/layouts/subpanel.svelte";
	import { layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { findPanelIndex, getSubPanelParent } from "$lib/views/utils";
	import type { SvelteHTMLElements } from "svelte/elements";

	interface Props {
		collection: CollectionData;
	}

	let { collection, ...props }: Props & SvelteHTMLElements["div"] = $props();
</script>

<div {...props} class="coll-card" data-asset-id={collection.id}>
	<div class="image-container">
		<img src={collection.thumbnail?.urls.thumbnail} alt={collection.name} class="collection-image" />
	</div>
	<div class="metadata">
		<span class="coll-name" title={collection.name}>{collection.name}</span>
		<span class="coll-created_on">{DateTime.fromJSDate(collection.created_on).toFormat("dd.MM.yyyy")}</span>
		<span class="coll-image_count">{collection.image_count} {collection.image_count === 1 ? "image" : "images"}</span>
	</div>
</div>

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
		background-color: var(--imag-90);
		transition: background-color 0.1s ease;
		text-align: left;
		overflow: overlay;
	}

	.image-container {
		height: 13em;
		background-color: var(--imag-60);
		pointer-events: none;
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
