<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { createWSConnection, type WSClient } from "$lib/api/websocket";
	import { getWsStats, getEventsSince } from "$lib/api";

	interface Props {
		autoConnect?: boolean;
		showHistory?: boolean;
		maxEvents?: number;
	}

	let { autoConnect = true, showHistory = true, maxEvents = 50 }: Props = $props();

	interface WSEvent {
		timestamp: Date;
		event: string;
		data: any;
	}

	let wsClient: WSClient | null = null;
	let connected = $state(false);
	let events: WSEvent[] = $state([]);
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
		if (wsClient) return;

		wsClient = createWSConnection(
			(event: string, data: any) => {
				switch (event) {
					case "connected":
						handleConnected(data);
						break;
					case "job-started":
						handleJobStarted(data);
						break;
					case "job-progress":
						handleJobProgress(data);
						break;
					case "job-completed":
						handleJobCompleted(data);
						break;
					case "job-failed":
						handleJobFailed(data);
						break;
					case "test":
						handleTestEvent(data);
						break;
				}
			},
			(error: Event) => {
				console.error("WebSocket error:", error);
				connected = false;
			},
			() => {
				connected = true;
			},
			(code: number, reason: string) => {
				// Only set connected to false if the disconnection was not a normal closure (1000) or going away (1001)
				if (code !== 1000 && code !== 1001 && code !== 1005) {
					connected = false;
				}
			}
		);
	}

	function disconnect() {
		if (wsClient) {
			wsClient.close();
			wsClient = null;
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

	function handleConnected(data: any) {
		connected = true;
		addEvent("connected", data);
	}

	function handleJobStarted(data: any) {
		currentJob = data.filename || data.imageId || "Unknown";
		currentProgress = 0;
		currentStep = "Job started";
		addEvent("job-started", data);
	}

	function handleJobCompleted(data: any) {
		stats.jobsProcessed++;
		addEvent("job-completed", data);
		setTimeout(() => {
			currentJob = null;
			currentProgress = 0;
			currentStep = "";
		}, 3000);
	}

	function handleJobFailed(data: any) {
		currentStep = `Error: ${data.error}`;
		addEvent("job-failed", data);
	}

	function handleJobProgress(data: any) {
		currentProgress = data.progress || 0;
		currentStep = data.step || "";
		addEvent("job-progress", data);
	}

	function handleTestEvent(data: any) {
		addEvent("test", data);
	}

	function clearEvents() {
		events = [];
		stats.totalEvents = 0;
	}

	async function loadStats() {
		try {
			const response = await getWsStats();
			if (response.data && "connectedClients" in response.data) {
				stats.connectedClients = response.data.connectedClients || 0;
			}
		} catch (e) {
			console.error("Failed to load stats:", e);
		}
	}

	async function loadHistory() {
		try {
			const response = await getEventsSince({ limit: 20 });
			if (response.data && "events" in response.data) {
				events = response.data.events.map((e: any) => ({
					timestamp: new Date(e.timestamp),
					event: e.event,
					data: e.data
				}));
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

<div class="ws-monitor">
	<div class="monitor-header">
		<h3>Real-time Events (WebSocket)</h3>
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
	.ws-monitor {
		background: var(--card-bg, #1f2937);
		border-radius: 8px;
		padding: 1rem;
		color: var(--text-primary, #f3f4f6);
	}

	.monitor-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--border-color, #374151);
	}

	.monitor-header h3 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.375rem 0.75rem;
		border-radius: 4px;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.status.connected {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
	}

	.status.disconnected {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}

	.indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.stats-bar {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		text-align: center;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--accent-color, #3b82f6);
	}

	.stat-label {
		font-size: 0.75rem;
		color: var(--text-secondary, #9ca3af);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.progress-section {
		background: var(--bg-secondary, #111827);
		padding: 1rem;
		border-radius: 6px;
		margin-bottom: 1rem;
	}

	.progress-info {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}

	.job-name {
		font-size: 0.875rem;
		font-weight: 500;
	}

	.progress-percent {
		font-size: 0.875rem;
		color: var(--accent-color, #3b82f6);
		font-weight: 600;
	}

	.progress-bar {
		height: 8px;
		background: var(--bg-tertiary, #1f2937);
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 0.5rem;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #3b82f6, #10b981);
		transition: width 0.3s ease;
	}

	.progress-step {
		font-size: 0.75rem;
		color: var(--text-secondary, #9ca3af);
	}

	.controls {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 4px;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-primary {
		background: var(--accent-color, #3b82f6);
		color: white;
	}

	.btn-primary:hover {
		background: #2563eb;
	}

	.btn-secondary {
		background: var(--bg-secondary, #374151);
		color: var(--text-primary, #f3f4f6);
	}

	.btn-secondary:hover {
		background: #4b5563;
	}

	.event-list {
		max-height: 400px;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.event {
		background: var(--bg-secondary, #111827);
		border-left: 3px solid;
		border-radius: 4px;
		padding: 0.75rem;
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}

	.event-type {
		font-weight: 600;
		font-size: 0.875rem;
	}

	.event-time {
		font-size: 0.75rem;
		color: var(--text-secondary, #9ca3af);
	}

	.event-data {
		font-family: monospace;
		font-size: 0.75rem;
		background: var(--bg-tertiary, #0f172a);
		padding: 0.5rem;
		border-radius: 4px;
		overflow-x: auto;
		white-space: pre-wrap;
		word-break: break-all;
	}

	.no-events {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary, #9ca3af);
		font-style: italic;
	}

	/* Scrollbar styling */
	.event-list::-webkit-scrollbar {
		width: 8px;
	}

	.event-list::-webkit-scrollbar-track {
		background: var(--bg-tertiary, #1f2937);
		border-radius: 4px;
	}

	.event-list::-webkit-scrollbar-thumb {
		background: var(--accent-color, #3b82f6);
		border-radius: 4px;
	}

	.event-list::-webkit-scrollbar-thumb:hover {
		background: #2563eb;
	}
</style>
