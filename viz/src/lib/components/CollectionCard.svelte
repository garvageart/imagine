<script lang="ts" module>
	export function openCollection(collection: Collection, currentContent: Content) {
		const collectionPath = `/collections/${collection.uid}`;
		if (page.url.pathname !== "/") {
			goto(collectionPath, { state: { from: page.url.pathname, data: collection } });
			return;
		}

		const currentParentIdx = findPanelIndex(layoutState.tree, getSubPanelParent(layoutState.tree, currentContent.paneKeyId)!);
		if (currentParentIdx === -1) {
			console.warn("Can't find panel in layout, navigating to collection page");
			goto(collectionPath, { state: { from: page.url.pathname, data: collection } });
			return;
		}

		const currentParent = layoutState.tree[currentParentIdx];
		const childIndex = currentParent.childs.content.findIndex((subPanel) => subPanel.paneKeyId === currentContent.paneKeyId);

		if (childIndex === -1) {
			console.warn(`Can't find child inside panel ${currentParent.paneKeyId}, navigating to collection page`);
			goto(collectionPath, { state: { from: page.url.pathname, data: collection } });
			return;
		}

		// Check if a view with this collection path already exists
		let existingView: VizView | undefined;
		let existingContent: Content | undefined;
		let existingParentPanel: VizSubPanelData | undefined;

		for (let i = 0; i < layoutState.tree.length; i++) {
			const panel = layoutState.tree[i];
			for (const content of panel.childs.content) {
				const view = content.views.find((v) => v.path === collectionPath);
				if (view) {
					existingView = view;
					existingContent = content;
					existingParentPanel = panel;
					break;
				}
			}
			if (existingView) {
				break;
			}
		}

		if (existingView && existingContent && existingParentPanel) {
			// Deactivate all views in the content and activate the existing one
			existingContent.views.forEach((v) => v.setActive(false));
			existingView.setActive(true);
			return;
		}

		const view = new VizView({
			name: collection.name,
			component: CollectionPage as any,
			path: collectionPath
		});

		addViewToContent(view, currentParentIdx, childIndex);

		// Find the parent subpanel and make the new view active
		const parentSubPanel = layoutState.tree[currentParentIdx];
		const targetContent = parentSubPanel.childs.content[childIndex];

		if (parentSubPanel) {
			// Deactivate all views in the content and activate the new one
			targetContent.views.forEach((v: VizView) => v.setActive(false));
			view.setActive(true);
			parentSubPanel.makeViewActive(view);
		}
	}
</script>

<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { addViewToContent, findSubPanel } from "$lib/utils/layout";
	import VizView from "$lib/views/views.svelte";
	import { DateTime } from "luxon";
	import CollectionPage from "../../routes/(app)/collections/[uid]/+page.svelte";
	import VizSubPanelData, { Content } from "$lib/layouts/subpanel.svelte";
	import { layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { findPanelIndex, getSubPanelParent } from "$lib/views/utils";
	import type { SvelteHTMLElements } from "svelte/elements";
	import { getFullImagePath, type Collection } from "$lib/api";

	interface Props {
		collection: Collection;
	}

	let { collection, ...props }: Props & SvelteHTMLElements["div"] = $props();
</script>

<div {...props} class="coll-card" data-asset-id={collection.uid}>
	<div class="image-container">
		{#if collection.thumbnail}
			<img src={getFullImagePath(collection.thumbnail.image_paths.thumbnail)} alt={collection.name} class="collection-image" />
		{:else}
			<div class="coll-no_thumbnail"></div>
		{/if}
	</div>
	<div class="metadata">
		<span class="coll-name" title={collection.name}>{collection.name}</span>
		<span class="coll-created_at">{DateTime.fromISO(collection.created_at).toFormat("dd.MM.yyyy")}</span>
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
		background-color: var(--imag-80);
		pointer-events: none;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.collection-image {
		width: 100%;
		height: 100%;
	}

	.coll-no_thumbnail {
		background-color: var(--imag-40);
		width: 60%;
		height: 90%;
	}

	.metadata {
		display: flex;
		flex-direction: column;
		padding: 1em;
		max-height: 10em;
		overflow: hidden;
	}
</style>
