<script lang="ts">
	import { onMount } from "svelte";
	import { executeSearch, type Image, type Collection } from "$lib/api";
	import CollectionCard from "$lib/components/CollectionCard.svelte";
	import ImageCard from "$lib/components/ImageCard.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { slide } from "svelte/transition";
	import { initDB } from "$lib/db/client";

	let favouriteImages = $state<Image[]>([]);
	let favouriteCollections = $state<Collection[]>([]);
    let db = $state(initDB());
	let loading = $state(true);

	// Collapse state
	let showCollections = $state(false);
	let showImages = $state(false);
	let settingsLoaded = false;

	// Load settings directly into state as early as possible
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
					style="font-size: 1.2em;"
				/>
				<span>Collections</span>
				<span class="count">({favouriteCollections.length})</span>
			</div>
		</button>
		{#if showCollections}
			<div class="grid-container" transition:slide>
				{#if favouriteCollections.length === 0 && !loading}
					<div class="empty-state">No favourite collections</div>
				{:else}
					<div class="grid collections-grid">
						{#each favouriteCollections as collection}
							<div class="card-wrapper">
								<CollectionCard {collection} />
							</div>
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
					style="font-size: 1.2em;"
				/>
				<span>Images</span>
				<span class="count">({favouriteImages.length})</span>
			</div>
		</button>
		{#if showImages}
			<div class="grid-container" transition:slide>
				{#if favouriteImages.length === 0 && !loading}
					<div class="empty-state">No favourite images</div>
				{:else}
					<div class="grid images-grid">
						{#each favouriteImages as image}
							<div class="card-wrapper">
								<ImageCard asset={image} />
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.favourites-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow-y: auto;
		background-color: var(--imag-bg);
		color: var(--imag-text-color);
		padding: 0.5rem;
		gap: 0.5rem;
	}

	.section {
		display: flex;
		flex-direction: column;
	}

	.section-header {
		display: flex;
		align-items: center;
		width: 100%;
		background: none;
		border: none;
		color: var(--imag-text-color);
		cursor: pointer;
		padding: 0.5rem;
		font-family: inherit;
		font-size: 1rem;
		font-weight: bold;
		text-align: left;
		border-radius: 4px;

		&:hover {
			background-color: var(--imag-hover);
		}
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.count {
		font-weight: normal;
		opacity: 0.7;
		font-size: 0.9em;
	}

	.grid-container {
		overflow: hidden; /* For slide transition */
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
		gap: 0.5rem;
		padding: 0.5rem;
	}

	.empty-state {
		padding: 1rem;
		text-align: center;
		color: var(--imag-text-muted);
		font-style: italic;
	}

	/* Scale down the cards to look like "little thumbnails" */
	.card-wrapper {
		/* Force the cards to be smaller */
		:global(.coll-card) {
			font-size: 0.7rem; /* Scale down text */
		}

		:global(.image-container) {
			height: 6rem !important; /* Override fixed height */
		}

		:global(.image-card) {
			padding: 0.4rem;
			font-size: 0.7rem;
		}

		:global(.image-card-meta) {
			display: none; /* Hide metadata for cleaner thumbnail look */
		}

		:global(.metadata) {
			padding: 0.5rem;
		}

		/* Hide extra details in collection card if needed */
		:global(.coll-created_at),
		:global(.coll-image_count) {
			display: none;
		}
	}
</style>
