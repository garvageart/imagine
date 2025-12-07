<script lang="ts">
	import { invalidate } from "$app/navigation";
	import { page } from "$app/state";
	import { clearImageCache } from "$lib/api";
	import Button from "$lib/components/Button.svelte";
	import ConfirmationModal from "$lib/components/modals/ConfirmationModal.svelte";
	import { modal } from "$lib/states/index.svelte";
	import { formatBytes } from "$lib/utils/images";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import AdminRouteShell from "$lib/components/admin/AdminRouteShell.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";

	let { data } = $props();

	let cacheStatus = $derived(data.cacheStatus);
	let loading = $state(false);
	let showClearConfirm = $state(false);

	function openClearConfirm() {
		showClearConfirm = true;
		modal.show = true;
	}

	function closeClearConfirm() {
		showClearConfirm = false;
		modal.show = false;
	}

	async function handleClearCache() {
		loading = true;
		closeClearConfirm();

		try {
			const response = await clearImageCache();
			if (response.status !== 200) {
				toastState.addToast({
					type: "error",
					message: response.data.error || "Failed to clear image cache."
				});
				return;
			}
			toastState.addToast({
				type: "success",
				message: "Image cache cleared successfully."
			});
			await invalidate(page.url.pathname);
		} catch (e) {
			toastState.addToast({
				type: "error",
				message: "Error clearing cache."
			});
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Cache - Admin</title>
</svelte:head>

<AdminRouteShell heading="Cache Management" description="Monitor and manage the image processing cache">
	{#snippet actions()}
		<Button variant="small" onclick={openClearConfirm} disabled={loading} hoverColor="var(--imag-alert-color)">
			<MaterialIcon iconName="delete_sweep" />
			{#if loading}
				Clearing...
			{:else}
				Clear Cache
			{/if}
		</Button>
	{/snippet}

	<div class="cache-grid-layout">
		<div class="cache-status-section">
			<div class="section-header">
				<MaterialIcon iconName="memory" />
				<h2>Cache Statistics</h2>
			</div>
			<div class="cache-stats">
				<div class="stat-item">
					<span class="label">Total Size</span>
					<span class="value">{formatBytes(cacheStatus.size)}</span>
				</div>
				<div class="stat-item">
					<span class="label">Total Items</span>
					<span class="value">{cacheStatus.items}</span>
				</div>
				<div class="stat-item">
					<span class="label">Cache Hits</span>
					<span class="value">{cacheStatus.hits}</span>
				</div>
				<div class="stat-item">
					<span class="label">Cache Misses</span>
					<span class="value">{cacheStatus.misses}</span>
				</div>
				<div class="stat-item">
					<span class="label">Hit Ratio</span>
					<span class="value">{(cacheStatus.hit_ratio * 100).toFixed(2)}%</span>
				</div>
			</div>
		</div>
	</div>
</AdminRouteShell>

{#if showClearConfirm && modal.show}
	<ConfirmationModal
		title="Clear Image Cache"
		confirmText="Clear Cache"
		onConfirm={handleClearCache}
		onCancel={closeClearConfirm}
	>
		<p>
			Are you sure you want to clear the entire image cache?
			<br />
			This will remove all generated thumbnails and previews. They will be regenerated on demand, which may increase server load temporarily.
		</p>
	</ConfirmationModal>
{/if}

<style lang="scss">
	.cache-status-section {
		background-color: var(--imag-100);
		border-radius: 12px;
		padding: 1.5rem;
		border: 1px solid var(--imag-90);
		max-width: 600px;
	}

	.section-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1.5rem;

		h2 {
			font-size: 1.25rem;
			font-weight: 600;
			margin: 0;
		}
	}

	.cache-stats {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.stat-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: var(--imag-90);
		border-radius: 0.5rem;

		.label {
			color: var(--imag-40);
			font-weight: 500;
		}

		.value {
			font-weight: 600;
			font-family: var(--imag-code-font);
			font-size: 1.1em;
		}
	}
</style>
