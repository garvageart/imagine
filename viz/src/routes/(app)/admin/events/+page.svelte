<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { getSSEStats, getSSEMetrics, getEventHistory, clearEventHistory } from "$lib/api/sse";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import type { PageData } from "./$types";

	let { data }: { data: PageData } = $props();

	let refreshing = $state(false);

	// SSE Stats - initialize from load data
	let stats = $state(data.stats);

	// SSE Metrics - initialize from load data
	let metrics = $state(data.metrics);

	// Event History - initialize from load data
	let history: any[] = $state(data.history || []);
	let historyFilter = $state("all");
	let historySearch = $state("");

	// Auto-refresh
	let autoRefresh = $state(true);
	let refreshInterval: number | null = null;

	function showMessage(message: string, type: "success" | "error" | "info" = "info") {
		toastState.addToast({ message, type });
	}

	async function loadStats() {
		try {
			const res = await getSSEStats();
			if (res.status === 200) {
				stats = res.data;
			}
		} catch (e) {
			console.error("Failed to load stats:", e);
		}
	}

	async function loadMetrics() {
		try {
			const res = await getSSEMetrics();
			if (res.status === 200) {
				metrics = res.data;
			}
		} catch (e) {
			console.error("Failed to load metrics:", e);
		}
	}

	async function loadHistory() {
		try {
			const res = await getEventHistory();
			if (res.status === 200) {
				history = res.data.events || [];
			}
		} catch (e) {
			console.error("Failed to load history:", e);
		}
	}

	async function refreshAll() {
		refreshing = true;
		try {
			await Promise.all([loadStats(), loadMetrics(), loadHistory()]);
		} catch (e) {
			showMessage("Refresh failed: " + (e as Error).message, "error");
		} finally {
			refreshing = false;
		}
	}

	async function clearHistory() {
		if (!confirm("Clear all event history?")) return;
		try {
			const res = await clearEventHistory();
			if (res.status === 200) {
				history = [];
				showMessage("Event history cleared", "success");
			}
		} catch (e) {
			showMessage("Clear failed: " + (e as Error).message, "error");
		}
	}

	function toggleAutoRefresh() {
		autoRefresh = !autoRefresh;
		if (autoRefresh) {
			startAutoRefresh();
		} else {
			stopAutoRefresh();
		}
	}

	function startAutoRefresh() {
		if (refreshInterval) return;
		refreshInterval = window.setInterval(refreshAll, 5000);
	}

	function stopAutoRefresh() {
		if (refreshInterval) {
			clearInterval(refreshInterval);
			refreshInterval = null;
		}
	}

	const filteredHistory = $derived(() => {
		let filtered = history;

		if (historyFilter !== "all") {
			filtered = filtered.filter((e) => e.eventType === historyFilter);
		}

		if (historySearch) {
			const search = historySearch.toLowerCase();
			filtered = filtered.filter(
				(e) => e.eventType.toLowerCase().includes(search) || JSON.stringify(e.data).toLowerCase().includes(search)
			);
		}

		return filtered;
	});

	const eventTypes = $derived(() => {
		const types = new Set<string>();
		history.forEach((e) => types.add(e.eventType));
		return Array.from(types).sort();
	});

	function formatTimestamp(ts: string) {
		return new Date(ts).toLocaleString();
	}

	function getMaxEventCount(): number {
		const values = Object.values(metrics.eventsByType || {});
		const nums = values.map((v) => (typeof v === "number" ? v : Number(v) || 0));
		return nums.length ? Math.max(...nums) : 0;
	}

	function getEventFillWidth(count: unknown): number {
		const num = typeof count === "number" ? count : Number(count) || 0;
		const max = getMaxEventCount();
		return max > 0 ? (num / max) * 100 : 0;
	}

	function formatJSON(data: any) {
		return JSON.stringify(data, null, 2);
	}

	onMount(async () => {
		if (autoRefresh) {
			startAutoRefresh();
		}
	});

	onDestroy(() => {
		stopAutoRefresh();
	});
</script>

<svelte:head>
	<title>Events - Admin</title>
</svelte:head>

<div class="admin-page">
	<header class="page-header">
		<div class="header-left">
			<a href="/admin" class="back-link">
				<MaterialIcon iconName="arrow_back" />
			</a>
			<div>
				<h1>Event Monitor</h1>
				<p class="subtitle">SSE metrics and event history</p>
			</div>
		</div>
		<div class="header-actions">
			<Button onclick={toggleAutoRefresh} class="control-button">
				<MaterialIcon iconName={autoRefresh ? "pause" : "play_arrow"} />
				{autoRefresh ? "Auto-refresh: ON" : "Auto-refresh: OFF"}
			</Button>
			<Button onclick={refreshAll} disabled={refreshing} class="control-button">
				<MaterialIcon iconName="refresh" />
				Refresh
			</Button>
		</div>
	</header>

	<!-- Connection Stats -->
	<section class="content-section">
		<div class="section-header">
			<MaterialIcon iconName="link" />
			<h2>Connection Statistics</h2>
		</div>
		<div class="stats-grid">
			<div class="stat-card">
				<MaterialIcon iconName="sensors" />
				<div class="stat-content">
					<span class="stat-value">{stats.connectedClients}</span>
					<span class="stat-label">Connected Clients</span>
				</div>
			</div>
			<div class="stat-card">
				<MaterialIcon iconName="timeline" />
				<div class="stat-content">
					<span class="stat-value">{metrics.totalEvents}</span>
					<span class="stat-label">Total Events</span>
				</div>
			</div>
			<div class="stat-card">
				<MaterialIcon iconName="schedule" />
				<div class="stat-content">
					<span class="stat-value">{new Date(stats.timestamp).toLocaleTimeString()}</span>
					<span class="stat-label">Last Updated</span>
				</div>
			</div>
			<div class="stat-card">
				<MaterialIcon iconName="groups" />
				<div class="stat-content">
					<span class="stat-value">{stats.clientIds.length}</span>
					<span class="stat-label">Active Clients</span>
				</div>
			</div>
		</div>
	</section>

	<!-- Performance Metrics -->
	<section class="content-section">
		<div class="section-header">
			<MaterialIcon iconName="analytics" />
			<h2>Performance Metrics</h2>
		</div>
		<div class="metrics-grid">
			<div class="metric-card">
				<div class="metric-icon">
					<MaterialIcon iconName="speed" />
				</div>
				<div class="metric-content">
					<span class="metric-value">{metrics.totalEvents}</span>
					<span class="metric-label">Total Events</span>
				</div>
			</div>
			<div class="metric-card">
				<div class="metric-icon">
					<MaterialIcon iconName="trending_up" />
				</div>
				<div class="metric-content">
					<span class="metric-value">{metrics.connectedClients}</span>
					<span class="metric-label">Active Connections</span>
				</div>
			</div>
			<div class="metric-card">
				<div class="metric-icon">
					<MaterialIcon iconName="lightbulb" />
				</div>
				<div class="metric-content">
					<span class="metric-value">{Object.keys(metrics.eventsByType).length}</span>
					<span class="metric-label">Event Types</span>
				</div>
			</div>
			<div class="metric-card">
				<div class="metric-icon">
					<MaterialIcon iconName="update" />
				</div>
				<div class="metric-content">
					<span class="metric-value">{new Date(metrics.timestamp).toLocaleTimeString()}</span>
					<span class="metric-label">Last Update</span>
				</div>
			</div>
		</div>
	</section>
	{#if Object.keys(metrics.eventsByType || {}).length > 0}
		<!-- Event Types Distribution -->
		<section class="content-section">
			<div class="section-header">
				<MaterialIcon iconName="bar_chart" />
				<h2>Event Types Distribution</h2>
			</div>
			<div class="event-types">
				{#each Object.entries(metrics.eventsByType || {}) as [type, count]}
					<div class="event-type-card">
						<div class="event-type-info">
							<span class="event-type-name">{type}</span>
							<span class="event-type-count">{count} events</span>
						</div>
						<div class="event-type-bar">
							<div class="event-type-fill" style="width: {getEventFillWidth(count)}%"></div>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Event History -->
	<section class="content-section">
		<div class="section-header">
			<MaterialIcon iconName="history" />
			<h2>Event History</h2>
			<span class="badge">{history.length}</span>
		</div>

		<div class="history-controls">
			<div class="filter-group">
				<select bind:value={historyFilter} aria-label="Filter by event type">
					<option value="all">All Events</option>
					{#each eventTypes() as type}
						<option value={type}>{type}</option>
					{/each}
				</select>
			</div>
			<div class="search-group">
				<MaterialIcon iconName="search" />
				<input type="text" bind:value={historySearch} placeholder="Search events..." />
			</div>
			<Button onclick={clearHistory} class="control-button">
				<MaterialIcon iconName="delete_sweep" />
				Clear History
			</Button>
		</div>

		{#if filteredHistory().length === 0}
			<div class="empty-state">
				<MaterialIcon iconName="inbox" />
				<p>No events found</p>
			</div>
		{:else}
			<div class="history-list">
				{#each filteredHistory() as event}
					<details class="event-item">
						<summary class="event-summary">
							<div class="event-header">
								<span class="event-type">{event.eventType}</span>
								<span class="event-time">{formatTimestamp(event.timestamp)}</span>
							</div>
							<MaterialIcon iconName="arrow_drop_down" />
						</summary>
						<div class="event-details">
							<div class="event-field">
								<strong>Client ID:</strong>
								<code>{event.clientId}</code>
							</div>
							<div class="event-field">
								<strong>Data:</strong>
								<pre>{formatJSON(event.data)}</pre>
							</div>
						</div>
					</details>
				{/each}
			</div>
		{/if}
	</section>
</div>

<style lang="scss">
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		gap: 1rem;
		flex-wrap: wrap;

		.header-left {
			display: flex;
			align-items: center;
			gap: 1rem;
		}

		.back-link {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 40px;
			height: 40px;
			border-radius: 50%;
			background: var(--imag-90);
			color: var(--imag-text-color);
			transition: all 0.2s;

			&:hover {
				background: var(--imag-80);
			}
		}

		h1 {
			margin: 0;
			font-size: 1.75rem;
			font-weight: 600;
		}

		.subtitle {
			margin: 0.25rem 0 0 0;
			color: var(--imag-40);
			font-size: 0.95rem;
		}

		.header-actions {
			display: flex;
			gap: 0.75rem;
		}
	}

	.content-section {
		background: var(--imag-100);
		border-radius: 12px;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
		border: 1px solid var(--imag-90);
	}

	.section-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1.25rem;

		h2 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
		}
	}

	.badge {
		padding: 0.25rem 0.625rem;
		background: var(--imag-80);
		border-radius: 1rem;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--imag-text-color);
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12em, 1fr));
		gap: 1rem;
	}

	.stat-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.25rem;
		background: var(--imag-90);
		border-radius: 1rem;
		border: 2px solid var(--imag-80);
		transition: border-color 0.2s;

		&:hover {
			border-color: var(--imag-70);
		}

		.stat-content {
			display: flex;
			flex-direction: column;
		}

		.stat-value {
			font-size: 2rem;
			font-weight: 700;
			line-height: 1;
			color: var(--imag-text-color);
		}

		.stat-label {
			font-size: 0.875rem;
			color: var(--imag-40);
			margin-top: 0.25rem;
		}
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.metric-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.25rem;
		background: var(--imag-90);
		border-radius: 1rem;
		border: 2px solid var(--imag-80);
		transition: border-color 0.2s;

		&:hover {
			border-color: var(--imag-70);
		}

		.metric-icon {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 3em;
			height: 3em;
			background: var(--imag-primary);
			color: var(--imag-text-color);
			border-radius: 1rem;
		}

		.metric-content {
			display: flex;
			flex-direction: column;
		}

		.metric-value {
			font-size: 1.5rem;
			font-weight: 700;
			line-height: 1;
		}

		.metric-label {
			font-size: 0.875rem;
			color: var(--imag-40);
			margin-top: 0.25rem;
		}
	}

	:global(.control-button) {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.875rem 1.25rem;
		border-radius: 0.5rem;
		font-size: 0.95rem;
		font-weight: 500;
		background-color: var(--imag-80);
		color: var(--imag-text-color);
		transition: background-color 0.2s;

		&:hover {
			background-color: var(--imag-70);
		}
	}

	.event-types {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.event-type-card {
		padding: 1rem;
		background: var(--imag-95);
		border-radius: 0.5rem;
	}

	.event-type-info {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}

	.event-type-name {
		font-weight: 600;
		font-size: 0.95rem;
	}

	.event-type-count {
		font-size: 0.875rem;
		color: var(--imag-40);
	}

	.event-type-bar {
		height: 0.5rem;
		background: var(--imag-90);
		border-radius: 0.25rem;
		overflow: hidden;
	}

	.event-type-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--imag-primary), var(--imag-accent-color));
		transition: width 0.3s ease;
	}

	.history-controls {
		display: flex;
		gap: 1rem;
		margin-bottom: 1.25rem;
		flex-wrap: wrap;
		align-items: center;
	}

	.filter-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;

		select {
			padding: 0.5rem 0.75rem;
			border: 1px solid var(--imag-80);
			border-radius: 0.375rem;
			background: var(--imag-100);
			color: var(--imag-text-color);
			font-size: 0.875rem;

			&:focus {
				outline: none;
				border-color: var(--imag-primary);
			}
		}
	}

	.search-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex: 1;
		max-width: 400px;
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--imag-80);
		border-radius: 0.375rem;
		background: var(--imag-100);

		&:focus-within {
			border-color: var(--imag-primary);
		}

		input {
			border: none;
			background: transparent;
			flex: 1;
			font-size: 0.875rem;
			color: var(--imag-text-color);

			&:focus {
				outline: none;
			}
		}
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		color: var(--imag-40);

		p {
			margin: 0.5rem 0 0 0;
			font-size: 1rem;
		}
	}

	.history-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		max-height: 600px;
		overflow-y: auto;
	}

	.event-item {
		background: var(--imag-95);
		border-radius: 8px;
		border: 1px solid var(--imag-90);
		overflow: hidden;
	}

	.event-summary {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		cursor: pointer;
		list-style: none;

		&:hover {
			background: var(--imag-90);
		}
	}

	.event-header {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.event-type {
		font-weight: 600;
		font-size: 0.95rem;
		padding: 0.25rem 0.75rem;
		background: var(--imag-primary);
		color: white;
		border-radius: 4px;
	}

	.event-time {
		font-size: 0.825rem;
		color: var(--imag-40);
	}

	.event-details {
		padding: 1rem;
		border-top: 1px solid var(--imag-90);
		background: var(--imag-100);
	}

	.event-field {
		margin-bottom: 1rem;

		&:last-child {
			margin-bottom: 0;
		}

		strong {
			display: block;
			margin-bottom: 0.5rem;
			font-size: 0.875rem;
		}

		code {
			display: block;
			padding: 0.5rem;
			background: var(--imag-90);
			border-radius: 4px;
			font-size: 0.825rem;
			font-family: "Courier New", monospace;
		}

		pre {
			padding: 1rem;
			background: var(--imag-90);
			border-radius: 4px;
			overflow-x: auto;
			font-size: 0.825rem;
			font-family: "Courier New", monospace;
			line-height: 1.4;
		}
	}
</style>
