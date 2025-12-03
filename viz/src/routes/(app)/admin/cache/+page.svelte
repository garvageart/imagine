<script lang="ts">
	import { invalidate } from "$app/navigation";
	import { page } from "$app/state";
	import { clearImageCache, type CacheStatusResponse } from "$lib/api";
	import Button from "$lib/components/Button.svelte";
	import { formatBytes } from "$lib/utils/images";

	let { data } = $props();

	let cacheStatus = $derived(data.cacheStatus);
	let loading = $state(false);
	let error: string | null = $state(null);
	let message: string | null = $state(null);

	async function clearCache() {
		if (!confirm("Are you sure you want to clear the image cache? This action cannot be undone.")) {
			return;
		}

		loading = true;

		const response = await clearImageCache();
		if (response.status !== 200) {
			error = response.data.error || "Failed to clear image cache.";
			return;
		}
		message = "Image cache cleared successfully.";
		loading = false;

		await invalidate(page.url.pathname);
	}
</script>

<div class="admin-cache-container">
	<h1>Admin Cache Management</h1>
	<div class="cache-status-section">
		<h2>Cache Status</h2>
		<div class="cache-grid">
			<div>
				<p>Total Size:</p>
				<p>{formatBytes(cacheStatus.size)}</p>
			</div>
			<div>
				<p>Total Items:</p>
				<p>{cacheStatus.items}</p>
			</div>
			<div>
				<p>Cache Hits:</p>
				<p>{cacheStatus.hits}</p>
			</div>
			<div>
				<p>Cache Misses:</p>
				<p>{cacheStatus.misses}</p>
			</div>
			<div>
				<p>Hit Ratio:</p>
				<p>{(cacheStatus.hit_ratio * 100).toFixed(2)}%</p>
			</div>
		</div>

		<Button class="clear-cache-button" onclick={clearCache} disabled={loading}>
			{#if loading}
				Clearing Cache...
			{:else}
				Clear Image Cache
			{/if}
		</Button>
	</div>

	{#if message}
		<p class="success-message">{message}</p>
	{/if}
</div>

<style lang="scss">
	.admin-cache-container {
		padding: 1rem;
		width: 40%;
		margin: 0 auto;
	}

	h1 {
		font-size: 2em;
		font-weight: bold;
		margin-bottom: 1em;
	}

	h2 {
		font-size: 1.5em;
		font-weight: bold;
		margin-bottom: 0.5em;
	}

	.cache-status-section {
		background-color: var(--imag-100);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		border-radius: 0.5em;
		padding: 1.5em;
		margin-bottom: 1.5em;
	}

	.cache-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1em;

		div p:first-child {
			color: var(--imag-text-color);
		}

		div p:last-child {
			font-size: 1.1em;
			font-weight: 500;
		}
	}

	// .error-message {
	// 	color: #ef4444; /* Red-500 */
	// 	margin-top: 1em;
	// }

	.success-message {
		color: #22c55e; /* Green-500 */
		margin-top: 1em;
	}
</style>
