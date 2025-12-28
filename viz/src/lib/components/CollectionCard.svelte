<script lang="ts" module>
	export function openCollection(
		collection: Collection,
		currentContent: Content
	) {
		const collectionPath = `/collections/${collection.uid}`;
		if (page.url.pathname !== "/") {
			goto(collectionPath, { state: { from: page.url.pathname } });
			return;
		}

		const currentParentIdx = findPanelIndex(
			layoutState.tree,
			getSubPanelParent(layoutState.tree, currentContent.paneKeyId)!
		);
		if (currentParentIdx === -1) {
			console.warn("Can't find panel in layout, navigating to collection page");
			goto(collectionPath, { state: { from: page.url.pathname } });
			return;
		}

		const currentParent = layoutState.tree[currentParentIdx];
		const childIndex = currentParent.childs.content.findIndex(
			(subPanel) => subPanel.paneKeyId === currentContent.paneKeyId
		);

		if (childIndex === -1) {
			console.warn(
				`Can't find child inside panel ${currentParent.paneKeyId}, navigating to collection page`
			);
			goto(collectionPath, { state: { from: page.url.pathname } });
			return;
		}

		// Check if a view with this collection path already exists
		let existingView: VizView | undefined;
		let existingContent: Content | undefined;
		let existingParentPanel: VizSubPanelData | undefined;

		console.debug(
			`[openCollection] Searching for view with path: "${collectionPath}"`
		);
		for (let i = 0; i < layoutState.tree.length; i++) {
			const panel = layoutState.tree[i];
			// Ensure we have content to iterate over
			if (!panel.childs?.content) continue;
			for (const content of panel.childs.content) {
				for (const v of content.views) {
					console.debug(
						`[openCollection] Checking view id=${v.id} path="${v.path}" name="${v.name}"`
					);
					// Match by exact path or by component type if it's a collection page
					// This allows reusing the same "slot" for different collections
					if (v.path === collectionPath || (v.component === CollectionPage as any && content.paneKeyId === currentContent.paneKeyId)) {
						console.debug(`[openCollection] Found match: ${v.id}`);
						existingView = v;
						existingContent = content;
						existingParentPanel = panel;
						break;
					}
				}
				if (existingView) break;
			}

			if (existingView) {
				break;
			}
		}

		if (existingView && existingContent && existingParentPanel) {
			console.debug(
				`[openCollection] Activating existing view: ${existingView.id}`
			);
			
			// If it's the same component but different collection, reset it
			if (existingView.path !== collectionPath) {
				existingView.reset(collection.name);
				existingView.path = collectionPath;
			}

			// Deactivate all views in the content and activate the existing one
			existingContent.views.forEach((v) => v.setActive(false));
			existingView.setActive(true);
			// IMPORTANT: Trigger layout update so the UI reflects the change
			existingParentPanel.makeViewActive(existingView);
			return;
		}

		console.debug(`[openCollection] No match found. Creating new view.`);
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
	import { addViewToContent } from "$lib/utils/layout";
	import VizView from "$lib/views/views.svelte";
	import { DateTime } from "luxon";
	import CollectionPage from "../../routes/(app)/collections/[uid]/+page.svelte";
	import VizSubPanelData, { Content } from "$lib/layouts/subpanel.svelte";
	import { layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { findPanelIndex, getSubPanelParent } from "$lib/views/utils";
	import type { SvelteHTMLElements } from "svelte/elements";
	import { getFullImagePath, getImage, type Collection } from "$lib/api";

	interface Props {
		collection: Collection;
	}

	let { collection, ...props }: Props & SvelteHTMLElements["div"] = $props();

	let thumbnail = $derived(collection.thumbnail);

	$effect(() => {
		if (collection.thumbnail) {
			thumbnail = collection.thumbnail;
		} else if (collection.images && collection.images.length > 0) {
			getImage(collection.images[0].uid).then((res) => {
				if (res.status === 200) {
					thumbnail = res.data;
				}
			});
		} else {
			thumbnail = undefined;
		}
	});
</script>

<div {...props} class="coll-card" data-asset-id={collection.uid}>
	<div class="image-container">
		{#if thumbnail}
			<img
				src={getFullImagePath(thumbnail.image_paths.thumbnail)}
				alt={collection.name}
				class="collection-image"
			/>
		{:else}
			<div class="coll-no_thumbnail"></div>
		{/if}
	</div>
	<div class="metadata">
		<span class="coll-name" title={collection.name}>{collection.name}</span>
		<span class="coll-created_at"
			>{DateTime.fromISO(collection.created_at).toFormat("dd.MM.yyyy")}</span
		>
		<span class="coll-image_count"
			>{collection.image_count}
			{collection.image_count === 1 ? "image" : "images"}</span
		>
	</div>
</div>

<style lang="scss">
	.coll-name {
		font-size: 1em;
		font-weight: bold;
		font-family: var(--imag-display-font);
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
		object-fit: contain;
		object-position: center;
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
