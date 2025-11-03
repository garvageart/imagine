<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { MEDIA_SERVER } from "$lib/constants";
	import { createServerURL } from "$lib/utils/url";
	import { createSSEConnection } from "$lib/api/sse";

	interface Props {
		autoConnect?: boolean;
		showHistory?: boolean;
		maxEvents?: number;
	}

	let { autoConnect = true, showHistory = true, maxEvents = 50 }: Props = $props();

	interface SSEEvent {
		timestamp: Date;
		event: string;
		data: any;
	}

	let eventSource: EventSource | null = null;
	let connected = $state(false);
	let events: SSEEvent[] = $state([]);
	let stats = $state({
		connectedClients: 0,
		totalEvents: 0,
		jobsProcessed: 0
	});

	let currentJob: string | null = $state(null);
	let currentProgress = $state(0);
	let currentStep = $state("");

	// Event type colors
	const eventColors: Record<string, string> = {
		connected: "#3b82f6",
		"job-started": "#10b981",
		"job-progress": "#f59e0b",
		"job-completed": "#059669",
		"job-failed": "#ef4444",
		test: "#8b5cf6"
	};

	function connect() {
		if (eventSource) return;

		eventSource = createSSEConnection(
			(event, data) => {
				switch (event) {
					case "connected":
						handleConnected({ data: JSON.stringify(data) } as MessageEvent);
						break;
					case "job-started":
						handleJobStarted({ data: JSON.stringify(data) } as MessageEvent);
						break;
					case "job-progress":
						handleJobProgress({ data: JSON.stringify(data) } as MessageEvent);
						break;
					case "job-completed":
						handleJobCompleted({ data: JSON.stringify(data) } as MessageEvent);
						break;
					case "job-failed":
						handleJobFailed({ data: JSON.stringify(data) } as MessageEvent);
						break;
					case "test":
						handleTestEvent({ data: JSON.stringify(data) } as MessageEvent);
						break;
				}
			},
			(error) => {
				console.error("SSE error:", error);
				connected = false;
			}
		);

		// Reflect connection state quickly
		eventSource.onopen = () => {
			connected = true;
		};
	}

	function disconnect() {
		if (eventSource) {
			eventSource.close();
			eventSource = null;
			connected = false;
		}
	}

	function addEvent(event: string, data: any) {
		events = [
			{
				timestamp: new Date(),
				event,
				data
			},
			...events
		].slice(0, maxEvents);
		stats.totalEvents++;
	}

	function handleConnected(e: MessageEvent) {
		const data = JSON.parse(e.data);
		connected = true;
		addEvent("connected", data);
	}

	function handleJobStarted(e: MessageEvent) {
		const data = JSON.parse(e.data);
		currentJob = data.filename || data.imageId || "Unknown";
		currentProgress = 0;
		currentStep = "Job started";
		addEvent("job-started", data);
	}

	function handleJobCompleted(e: MessageEvent) {
		const data = JSON.parse(e.data);
		stats.jobsProcessed++;
		addEvent("job-completed", data);
		setTimeout(() => {
			currentJob = null;
			currentProgress = 0;
			currentStep = "";
		}, 3000);
	}

	function handleJobFailed(e: MessageEvent) {
		const data = JSON.parse(e.data);
		currentStep = `Error: ${data.error}`;
		addEvent("job-failed", data);
	}

	function handleJobProgress(e: MessageEvent) {
		const data = JSON.parse(e.data);
		currentProgress = data.progress || 0;
		currentStep = data.step || "";
		addEvent("job-progress", data);
	}

	function handleTestEvent(e: MessageEvent) {
		const data = JSON.parse(e.data);
		addEvent("test", data);
	}

	function clearEvents() {
		events = [];
		stats.totalEvents = 0;
	}

	async function loadStats() {
		try {
			const url = createServerURL(MEDIA_SERVER) + "/api/events/stats";
			const response = await fetch(url);
			if (response.ok) {
				const data = await response.json();
				stats.connectedClients = data.connectedClients || 0;
			}
		} catch (e) {
			console.error("Failed to load stats:", e);
		}
	}

	async function loadHistory() {
		try {
			const url = createServerURL(MEDIA_SERVER) + "/api/events/history?limit=20";
			const response = await fetch(url);
			if (response.ok) {
				const data = await response.json();
				if (data.events && Array.isArray(data.events)) {
					events = data.events.map((e: any) => ({
						timestamp: new Date(e.timestamp),
						event: e.event,
						data: e.data
					}));
				}
			}
		} catch (e) {
			console.error("Failed to load history:", e);
		}
	}

	onMount(() => {
		if (autoConnect) {
			connect();
		}
		if (showHistory) {
			loadHistory();
		}
		loadStats();

		// Refresh stats every 10 seconds
		const interval = setInterval(loadStats, 10000);
		return () => clearInterval(interval);
	});

	onDestroy(() => {
		disconnect();
	});
</script>

<div class="sse-monitor">
	<div class="monitor-header">
		<h3>Real-time Events</h3>
		<div class="status {connected ? 'connected' : 'disconnected'}">
			<span class="indicator"></span>
			{connected ? "Connected" : "Disconnected"}
		</div>
	</div>

	<div class="stats-bar">
		<div class="stat">
			<span class="stat-value">{stats.connectedClients}</span>
			<span class="stat-label">Clients</span>
		</div>
		<div class="stat">
			<span class="stat-value">{stats.totalEvents}</span>
			<span class="stat-label">Events</span>
		</div>
		<div class="stat">
			<span class="stat-value">{stats.jobsProcessed}</span>
			<span class="stat-label">Jobs Done</span>
		</div>
	</div>

	{#if currentJob}
		<div class="progress-section">
			<div class="progress-info">
				<span class="job-name">{currentJob}</span>
				<span class="progress-percent">{currentProgress}%</span>
			</div>
			<div class="progress-bar">
				<div class="progress-fill" style="width: {currentProgress}%"></div>
			</div>
			{#if currentStep}
				<div class="progress-step">{currentStep}</div>
			{/if}
		</div>
	{/if}

	<div class="controls">
		{#if connected}
			<button class="btn-secondary" onclick={disconnect}>Disconnect</button>
		{:else}
			<button class="btn-primary" onclick={connect}>Connect</button>
		{/if}
		<button class="btn-secondary" onclick={clearEvents}>Clear</button>
	</div>

	{#if showHistory}
		<div class="event-list">
			{#each events as event}
				<div class="event" style="border-left-color: {eventColors[event.event] || '#6b7280'}">
					<div class="event-header">
						<span class="event-type">{event.event}</span>
						<span class="event-time">{event.timestamp.toLocaleTimeString()}</span>
					</div>
					<div class="event-data">
						{JSON.stringify(event.data, null, 2)}
					</div>
				</div>
			{/each}
			{#if events.length === 0}
				<div class="no-events">No events yet</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.sse-monitor {
		background: var(--imag-100, #ffffff);
		border-radius: 8px;
		padding: 1.25rem;
		border: 1px solid rgba(0, 0, 0, 0.06);
	}

	.monitor-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.monitor-header h3 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 600;
	}

	.status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		padding: 0.375rem 0.75rem;
		border-radius: 20px;
	}

	.status.connected {
		background: #d1fae5;
		color: #065f46;
	}

	.status.disconnected {
		background: #fee2e2;
		color: #991b1b;
	}

	.indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.status.connected .indicator {
		animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	.stats-bar {
		display: flex;
		gap: 1rem;
		margin-bottom: 1rem;
		padding: 0.75rem;
		background: var(--imag-95, #f9fafb);
		border-radius: 6px;
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		flex: 1;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--imag-primary, #3b82f6);
	}

	.stat-label {
		font-size: 0.75rem;
		color: var(--imag-60, #6b7280);
		margin-top: 0.25rem;
	}

	.progress-section {
		background: var(--imag-95, #f9fafb);
		padding: 1rem;
		border-radius: 6px;
		margin-bottom: 1rem;
	}

	.progress-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.job-name {
		font-weight: 500;
		font-size: 0.875rem;
	}

	.progress-percent {
		font-weight: 700;
		color: var(--imag-primary, #3b82f6);
	}

	.progress-bar {
		width: 100%;
		height: 24px;
		background: #e5e7eb;
		border-radius: 12px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #3b82f6, #2563eb);
		transition: width 0.3s ease;
		border-radius: 12px;
	}

	.progress-step {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		color: var(--imag-60, #6b7280);
	}

	.controls {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	button {
		padding: 0.5rem 1rem;
		border-radius: 6px;
		border: none;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 120ms ease;
	}

	.btn-primary {
		background: var(--imag-primary, #3b82f6);
		color: white;
	}

	.btn-primary:hover {
		background: var(--imag-primary-dark, #2563eb);
	}

	.btn-secondary {
		background: var(--imag-90, #e5e7eb);
		color: var(--imag-text-color, #1f2937);
	}

	.btn-secondary:hover {
		background: var(--imag-80, #d1d5db);
	}

	.event-list {
		max-height: 400px;
		overflow-y: auto;
		border: 1px solid var(--imag-90, #e5e7eb);
		border-radius: 6px;
		padding: 0.5rem;
		background: var(--imag-98, #fafafa);
	}

	.event {
		background: white;
		border-left: 4px solid #3b82f6;
		border-radius: 4px;
		padding: 0.75rem;
		margin-bottom: 0.5rem;
		font-family: monospace;
		font-size: 0.75rem;
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.event-type {
		font-weight: 700;
		color: var(--imag-primary, #3b82f6);
	}

	.event-time {
		color: var(--imag-60, #6b7280);
		font-size: 0.7rem;
	}

	.event-data {
		color: var(--imag-40, #4b5563);
		white-space: pre-wrap;
		word-break: break-all;
	}

	.no-events {
		text-align: center;
		padding: 2rem;
		color: var(--imag-60, #6b7280);
	}
</style>
