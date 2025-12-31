<script lang="ts" module>
	import { workspaceState } from "$lib/states/workspace.svelte";
	import { TabGroup } from "$lib/layouts/model.svelte";

	export function openCollection(
		collection: Collection,
		currentGroup: TabGroup | null
	) {
		const collectionPath = `/collections/${collection.uid}`;
		if (page.url.pathname !== "/") {
			goto(collectionPath, { state: { from: page.url.pathname } });
			return;
		}

		const workspace = workspaceState.workspace;
		if (!workspace) {
			console.warn("Workspace not initialized, navigating to collection page");
			goto(collectionPath, { state: { from: page.url.pathname } });
			return;
		}

		// Check if a view with this collection path already exists
		const existingView = workspace.findViewWithPath(collectionPath);
		const existingGroup = workspace.findGroupWithPath(collectionPath);

		if (existingView && existingGroup) {
			console.debug(
				`[openCollection] Activating existing view: ${existingView.id}`
			);
			existingGroup.setActive(existingView.id);
			return;
		}

		console.debug(`[openCollection] No match found. Creating new view.`);
		const view = new VizView({
			name: collection.name,
			component: CollectionPage as any,
			path: collectionPath
		});

		// Add to current group if provided, otherwise add to root or first group found
		if (currentGroup) {
			currentGroup.addTab(view);
		} else {
			// Fallback: find first TabGroup in the tree
			const findFirstGroup = (node: any): TabGroup | null => {
				if (node instanceof TabGroup) return node;
				if (node.children) {
					for (const child of node.children) {
						const found = findFirstGroup(child);
						if (found) return found;
					}
				}
				return null;
			};
			const firstGroup = findFirstGroup(workspace.root);
			if (firstGroup) {
				firstGroup.addTab(view);
			} else {
				console.warn("No TabGroup found to add view to");
				goto(collectionPath, { state: { from: page.url.pathname } });
			}
		}
	}
</script>

<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import VizView from "$lib/views/views.svelte";
	import CollectionPage from "../../routes/(app)/collections/[uid]/+page.svelte";
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
			>{new Date(collection.created_at).toLocaleDateString()}</span
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
