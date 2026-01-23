<script lang="ts">
	import { onMount } from "svelte";
	import { api, executeSearch, type Image, type Collection } from "$lib/api";
	import CompactListItem from "./CompactListItem.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { slide, fade } from "svelte/transition";
	import { initDB } from "$lib/db/client";
	import { VizMimeTypes } from "$lib/constants";
	import { DragData } from "$lib/drag-drop/data";

	let favouriteImages = $state<Image[]>([]);
	let favouriteCollections = $state<Collection[]>([]);
	let db = $state(initDB());
	let loading = $state(true);
	let isDraggingOver = $state(false);
	let isInternalDrag = false;
	let dragCounter = 0;

	// Collapse state
	let showCollections = $state(false);
	let showImages = $state(false);
	let settingsLoaded = false;

	// Load settings directly into state
	db.then(async (db) => {
		try {
			const settings = await db.get("settings", "favourites_panel");
			if (settings) {
				if (settings.showCollections !== undefined) {
					showCollections = settings.showCollections;
				}
				if (settings.showImages !== undefined) {
					showImages = settings.showImages;
				}
			}
		} catch (e) {
			console.error("Failed to load favourites panel settings", e);
		} finally {
			settingsLoaded = true;
		}
	});

	async function loadFavourites() {
		loading = true;
		try {
			const res = await executeSearch("favourited:true");
			if (res.status === 200) {
				favouriteImages = res.data.images;
				favouriteCollections = res.data.collections;
			}
		} catch (e) {
			console.error("Failed to load favourites", e);
		} finally {
			loading = false;
		}
	}

	async function saveSettings() {
		if (!settingsLoaded) {
			return;
		}
		try {
			const fetchedDb = await db;
			await fetchedDb.put(
				"settings",
				{
					showCollections,
					showImages
				},
				"favourites_panel"
			);
		} catch (e) {
			console.error("Failed to save favourites panel settings", e);
		}
	}

	function handleDragStart(e: DragEvent) {
		isInternalDrag = true;
	}

	function handleDragEnd(e: DragEvent) {
		isInternalDrag = false;
		isDraggingOver = false;
		dragCounter = 0;
	}

	async function handleDrop(e: DragEvent) {
		isDraggingOver = false;
		dragCounter = 0;
		if (!e.dataTransfer || isInternalDrag) return;

		const imageUidsData = DragData.getData<string[]>(
			e.dataTransfer,
			VizMimeTypes.IMAGE_UIDS
		);
		const collectionUidsData = DragData.getData<string[]>(
			e.dataTransfer,
			VizMimeTypes.COLLECTION_UIDS
		);

		let changed = false;

		if (imageUidsData) {
			for (const uid of imageUidsData.payload) {
				await api.updateImage(uid, { favourited: true });
				changed = true;
			}
		}

		if (collectionUidsData) {
			for (const uid of collectionUidsData.payload) {
				await api.updateCollection(uid, { favourited: true });
				changed = true;
			}
		}

		if (changed) {
			loadFavourites();
		}
	}

	function handleDragEnter(e: DragEvent) {
		if (isInternalDrag) return;

		if (
			DragData.isType(e.dataTransfer!, VizMimeTypes.IMAGE_UIDS) ||
			DragData.isType(e.dataTransfer!, VizMimeTypes.COLLECTION_UIDS)
		) {
			e.preventDefault();
			dragCounter++;
			isDraggingOver = true;
		}
	}

	function handleDragLeave(e: DragEvent) {
		if (isInternalDrag) return;

		dragCounter--;
		
		// Use relatedTarget to check if we are truly leaving the container
		const relatedTarget = e.relatedTarget as Node;
		const currentTarget = e.currentTarget as Node;
		
		if (dragCounter <= 0 || (currentTarget && !currentTarget.contains(relatedTarget))) {
			isDraggingOver = false;
			dragCounter = 0;
		}
	}

	function handleDragOver(e: DragEvent) {
		if (isInternalDrag) return;

		if (
			DragData.isType(e.dataTransfer!, VizMimeTypes.IMAGE_UIDS) ||
			DragData.isType(e.dataTransfer!, VizMimeTypes.COLLECTION_UIDS)
		) {
			e.preventDefault();
			if (e.dataTransfer) {
				e.dataTransfer.dropEffect = "copy";
			}
			// Ensure we stay in dragging state
			if (!isDraggingOver) isDraggingOver = true;
		}
	}

	onMount(() => {
		loadFavourites();
	});

	$effect(() => {
		// Track dependencies
		const _c = showCollections;
		const _i = showImages;
		saveSettings();
	});
</script>

<div
	class="panel-container"
	ondragenter={handleDragEnter}
	ondragover={handleDragOver}
	ondragleave={handleDragLeave}
	ondrop={handleDrop}
	ondragstart={handleDragStart}
	ondragend={handleDragEnd}
	role="region"
	aria-label="Favourites Panel"
>
	<div class="favourites-panel">
		<!-- Collections Section -->
		<div class="section">
			<button
				class="section-header"
				onclick={() => (showCollections = !showCollections)}
			>
				<div class="header-content">
					<MaterialIcon
						iconName={showCollections ? "keyboard_arrow_down" : "chevron_right"}
					/>
					<span>Collections</span>
					<span class="count">({favouriteCollections.length})</span>
				</div>
			</button>
			{#if showCollections}
				<div class="list-container" transition:slide>
					{#if favouriteCollections.length === 0 && !loading}
						<div class="empty-state">No favourite collections</div>
					{:else}
						<div class="list">
							{#each favouriteCollections as collection (collection.uid)}
								<CompactListItem item={collection} type="collection" />
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Images Section -->
		<div class="section">
			<button class="section-header" onclick={() => (showImages = !showImages)}>
				<div class="header-content">
					<MaterialIcon
						iconName={showImages ? "keyboard_arrow_down" : "chevron_right"}
					/>
					<span>Images</span>
					<span class="count">({favouriteImages.length})</span>
				</div>
			</button>
			{#if showImages}
				<div class="list-container" transition:slide>
					{#if favouriteImages.length === 0 && !loading}
						<div class="empty-state">No favourite images</div>
					{:else}
						<div class="list">
							{#each favouriteImages as image (image.uid)}
								<CompactListItem item={image} type="image" />
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	{#if isDraggingOver}
		<div class="drop-overlay" transition:fade={{ duration: 150 }}>
			<div class="drop-overlay-content">
				<MaterialIcon
					iconName="favorite"
					style="font-size: 3rem; margin-bottom: 0.5rem; color: white;"
					fill={true}
				/>
				<span>Drop to favourite</span>
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.panel-container {
		position: relative;
		height: 100%;
		width: 100%;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.favourites-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow-y: auto;
		background-color: var(--imag-bg-color);
		color: var(--imag-text-color);
		padding: 0.5rem;
		gap: 0.75rem;
		box-sizing: border-box;
	}

	.section {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.section-header {
		display: flex;
		align-items: center;
		width: 100%;
		background: none;
		border: none;
		color: var(--imag-text-color);
		cursor: pointer;
		padding: 0.25rem;
		font-family: inherit;
		font-size: 0.9rem;
		font-weight: 600;
		text-align: left;
		border-radius: 0.25rem;

		&:hover {
			background-color: var(--imag-90);
			opacity: 0.9;
		}
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.count {
		font-weight: normal;
		opacity: 0.5;
		font-size: 0.8em;
	}

	.list-container {
		overflow: hidden;
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.empty-state {
		padding: 1rem;
		text-align: center;
		color: var(--imag-60);
		font-style: italic;
		font-size: 0.85rem;
	}

	.drop-overlay {
		position: absolute;
		inset: 0;
		z-index: 1000;
		color: white;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}

	.drop-overlay-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		pointer-events: none;
		border: 1px solid var(--imag-60);
		border-radius: 1rem;
		padding: 2rem;
		background: rgba(0, 0, 0, 0.5);
		color: white;
		font-weight: bold;
	}
</style>
