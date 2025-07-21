<script>
	import CollectionCard from "$lib/components/CollectionCard.svelte";
	import LoadingSpinner from "$lib/components/LoadingSpinner.svelte";
	import { performSearch } from "$lib/search/execute";
	import { search } from "$lib/states/index.svelte";
	import { onMount } from "svelte";

	let collections = $derived(search.data.collections);
	let images = $derived(search.data.images);
	let totalResults = $derived(collections.length + images.length);

	onMount(() => {
		if (search.value) {
			performSearch();
		}
	});
</script>

<svelte:head>
	<title>Search{search.value ? ` - ${search.value}` : ""}</title>
</svelte:head>

<div class="search-container">
	<h1>Search</h1>
	{#if search.loading}
		<div class="loading-container">
			<p id="search-loading-text">Searching for "{search.value}"...</p>
			<LoadingSpinner />
		</div>
	{:else if search.executed}
		<div class="results">
			{#if totalResults === 0}
				<!-- TODO: Create a suggestions component to show other collections
			 something like a closest match to what was searched			 
			-->
				<div class="no-results">
					<p>No results found for "{search.value}"</p>
				</div>
			{:else}
				{#if collections.length > 0}
					<section class="collections-section">
						<h2>Collections ({collections.length})</h2>
						<div class="collections-grid">
							{#each collections as collection}
								<CollectionCard {collection} />
							{/each}
						</div>
					</section>
				{/if}

				{#if images.length > 0}
					<section class="images-section">
						<h2>Images ({images.length})</h2>
						<div class="images-grid">
							{#each images as image}
								<div class="image-card">
									<img src={image.urls.preview} alt="{image.name} by {image.uploaded_by.username}" loading="lazy" />
									<div class="image-info">
										<h4 style="color: #222;">{image.name}</h4>
										<p>By {image.uploaded_by.username}</p>
									</div>
								</div>
							{/each}
						</div>
					</section>
				{/if}
			{/if}
		</div>
	{/if}
</div>

<style>
	.search-container {
		padding: 2rem 1rem;
		white-space: wrap;
		display: flex;
		align-items: center;
		flex-direction: column;
		overflow-y: auto;
	}

	h1 {
		margin-bottom: 1.5rem;
		text-align: center;
	}

	.loading-container,
	.no-results {
		text-align: center;
		padding: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-direction: column;
	}

	#search-loading-text {
		font-size: 1em;
		margin-bottom: 1rem;
	}

	.results {
		width: 100%;
	}

	.collections-section,
	.images-section {
		margin-bottom: 3rem;
		width: 100%;
	}

	h2 {
		margin-bottom: 1rem;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--imag-20);
	}

	.collections-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.5rem;
		width: 100%;
	}

	.images-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 1.5rem;
	}

	.image-card {
		border-radius: 8px;
		overflow: hidden;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s;
	}

	.image-card img {
		width: 100%;
		height: 180px;
		object-fit: cover;
		display: block;
	}

	.image-info {
		padding: 1rem;
		background: white;
	}

	.image-info h4 {
		margin: 0 0 0.5rem 0;
		font-size: 1rem;
	}

	.image-info p {
		margin: 0;
		font-size: 0.9rem;
		color: #666;
	}

	@media (max-width: 768px) {
		.collections-grid,
		.images-grid {
			grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		}
	}

	@media (max-width: 480px) {
		.collections-grid,
		.images-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
