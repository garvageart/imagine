<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { createWSConnection, type WSClient } from "$lib/api/websocket";
	import {
		listJobs,
		createJob,
		stopJobType,
		updateJobTypeConcurrency,
		getJobStats,
		startScheduler as apiStartScheduler,
		shutdownScheduler as apiShutdownScheduler
	} from "$lib/api";
	import { getJobsSnapshot, getEventsSince } from "$lib/api";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

	// Connection state
	let connected = $state(false);
	let wsClient: WSClient | null = null;
	let backfilledOnce = $state(false);
	let lastCursor = $state(0);
	let statsSyncInterval: ReturnType<typeof setInterval> | null = null;

	// Stats
	let stats = $state({
		activeCount: 0,
		completedCount: 0,
		failedCount: 0,
		totalProcessed: 0
	});

	// Job tracking
	let jobTracking = $state({
		active: new Map<string, any>(),
		completed: [] as any[],
		failed: [] as any[]
	});

	// Correlation helpers
	const imageToJobId = new Map<string, string>();

	// Topic counters
	let runningByTopic = $state({} as Record<string, number>);
	let queuedByTopic = $state({} as Record<string, number>);

	// Job management UI state
	let jobMgmt = $state({
		loading: false,
		types: [] as any[],
		concurrency: {} as Record<string, number>
	});

	function showMessage(message: string, type: "success" | "error" | "info" = "info") {
		toastState.addToast({ message, type });
	}

	// -------- Helpers (shared by live handling and backfill) --------
	function getTopicForJobId(jobId: string): string {
		switch (jobId) {
			case "thumbnailGeneration":
				return "image_process";
			case "xmpGeneration":
				return "xmp_generation";
			default:
				return jobId;
		}
	}

	function resolveActiveJobId(data: any): string | null {
		if (data?.jobId && jobTracking.active.has(data.jobId)) return data.jobId;
		const img = data?.imageId as string | undefined;
		if (img) {
			const mapped = imageToJobId.get(img);
			if (mapped && jobTracking.active.has(mapped)) return mapped;
			for (const [jid, job] of jobTracking.active.entries()) {
				if (job.imageId === img) {
					imageToJobId.set(img, jid);
					return jid;
				}
			}
		}
		return null;
	}

	function ensurePlaceholderFrom(data: any, ts?: Date) {
		if (!data?.jobId) return;
		const placeholder = {
			jobId: data.jobId,
			imageId: data.imageId,
			type: data.type || "unknown",
			filename: data.filename,
			startTime: ts ? new Date(ts) : new Date(),
			progress: typeof data.progress === "number" ? data.progress : 0,
			step: data.step || "In progress"
		};
		jobTracking.active.set(data.jobId, placeholder);
		if (data.imageId) imageToJobId.set(data.imageId, data.jobId);
		stats.activeCount = jobTracking.active.size;
	}

	function applyProgressUpdate(data: any) {
		const id = resolveActiveJobId(data);
		if (id) {
			const job = jobTracking.active.get(id)!;
			const nextProg = typeof data.progress === "number" ? Math.max(data.progress, job.progress ?? 0) : (job.progress ?? 0);
			jobTracking.active.set(id, { ...job, progress: nextProg, step: data.step ?? job.step });
			return;
		}
		if (data?.jobId) ensurePlaceholderFrom(data);
	}

	function removeActiveFor(data: any) {
		const id = resolveActiveJobId(data);
		if (!id) return null;
		const job = jobTracking.active.get(id);
		jobTracking.active.delete(id);
		if (job?.imageId) imageToJobId.delete(job.imageId);
		return job || null;
	}

	let statsSyncTimer: ReturnType<typeof setTimeout> | null = null;
	function scheduleStatsSync(delay = 300) {
		if (statsSyncTimer) clearTimeout(statsSyncTimer);
		statsSyncTimer = setTimeout(() => void loadJobStats(), delay);
	}

	function afterTerminalUpdate(data: any, removed: boolean) {
		stats.activeCount = jobTracking.active.size;
		const topic = getTopicForJobId(data.type);

		if (removed) {
			// Job was in our active map, so decrement running
			runningByTopic[topic] = Math.max(0, (runningByTopic[topic] || 0) - 1);
		} else {
			// Job completed without being in active map (very fast or missed start)
			// Try decrementing running first, then queued
			if ((runningByTopic[topic] || 0) > 0) {
				runningByTopic[topic] = Math.max(0, (runningByTopic[topic] || 0) - 1);
			} else if ((queuedByTopic[topic] || 0) > 0) {
				queuedByTopic[topic] = Math.max(0, (queuedByTopic[topic] || 0) - 1);
			}
		}

		// If everything quiets down quickly (fast jobs), ask server for the truth shortly after.
		if (jobTracking.active.size === 0) {
			scheduleStatsSync(250);
		}
	}

	// -------- Scheduler controls --------
	async function startScheduler() {
		try {
			const res = await apiStartScheduler();
			if (res.status === 200) {
				showMessage((res.data as any)?.message ?? "Scheduler started", "success");
			} else {
				showMessage((res as any).data?.error ?? `Start failed (${res.status})`, "error");
			}
		} catch (e) {
			showMessage("Start failed: " + (e as Error).message, "error");
		}
	}

	async function shutdownScheduler() {
		try {
			const res = await apiShutdownScheduler();
			if (res.status === 200) {
				showMessage((res.data as any)?.message ?? "Scheduler shutdown", "success");
			} else {
				showMessage((res as any).data?.error ?? `Shutdown failed (${res.status})`, "error");
			}
		} catch (e) {
			showMessage("Shutdown failed: " + (e as Error).message, "error");
		}
	}

	// -------- WebSocket wiring --------
	function connectWS() {
		if (wsClient) {
			return;
		}

		wsClient = createWSConnection(
			async (event: string, data: any) => {
				switch (event) {
					case "connected": {
						connected = true;
						if (data?.clientId) {
							showMessage(`Connected to WebSocket (Client: ${String(data.clientId).substring(0, 8)})`, "success");
						}
						if (!backfilledOnce) await bootstrapState();
						break;
					}
					case "job-started": {
						const idNum = Number((data as any)?.__id ?? 0);
						if (idNum && idNum <= lastCursor) break;
						if (idNum) lastCursor = idNum;
						if (!data?.jobId) break;
						jobTracking.active.set(data.jobId, {
							...data,
							startTime: new Date(),
							progress: 0,
							step: "Starting..."
						});
						if (data.imageId) imageToJobId.set(data.imageId, data.jobId);
						stats.activeCount = jobTracking.active.size;
						const topic = getTopicForJobId(data.type);
						runningByTopic[topic] = (runningByTopic[topic] || 0) + 1;
						if ((queuedByTopic[topic] || 0) > 0) queuedByTopic[topic] = (queuedByTopic[topic] || 0) - 1;
						break;
					}
					case "job-progress": {
						const idNum = Number((data as any)?.__id ?? 0);
						if (idNum && idNum <= lastCursor) break;
						if (idNum) lastCursor = idNum;
						applyProgressUpdate(data);
						break;
					}
					case "job-completed": {
						const idNum = Number((data as any)?.__id ?? 0);
						if (idNum && idNum <= lastCursor) break;
						if (idNum) lastCursor = idNum;
						const removedJob = removeActiveFor(data);
						if (removedJob) {
							jobTracking.completed = [{ ...removedJob, ...data, endTime: new Date() }, ...jobTracking.completed].slice(0, 50);
						} else {
							jobTracking.completed = [{ ...data, endTime: new Date(), startTime: new Date() }, ...jobTracking.completed].slice(
								0,
								50
							);
						}
						stats.completedCount++;
						stats.totalProcessed++;
						if (!removedJob && data.imageId) imageToJobId.delete(data.imageId);
						afterTerminalUpdate(data, !!removedJob);
						break;
					}
					case "job-failed": {
						const idNum = Number((data as any)?.__id ?? 0);
						if (idNum && idNum <= lastCursor) break;
						if (idNum) lastCursor = idNum;
						const removedJob = removeActiveFor(data);
						if (removedJob) {
							jobTracking.failed = [{ ...removedJob, ...data, endTime: new Date() }, ...jobTracking.failed].slice(0, 50);
						} else {
							jobTracking.failed = [{ ...data, endTime: new Date(), startTime: new Date() }, ...jobTracking.failed].slice(0, 50);
						}
						stats.failedCount++;
						if (!removedJob && data.imageId) imageToJobId.delete(data.imageId);
						afterTerminalUpdate(data, !!removedJob);
						break;
					}
				}
			},
			(error: Event) => {
				console.error("WebSocket error:", error);
				connected = false;
				showMessage("WebSocket connection error", "error");
			},
			() => {
				connected = true;
			},
			(code: number, reason: string) => {
				connected = false;
				showMessage(`WebSocket disconnected (${code}): ${reason}`, "info");
			}
		);
	}

	function disconnectWS() {
		if (wsClient) {
			wsClient.close();
			wsClient = null;
			connected = false;
			showMessage("Disconnected from WebSocket", "info");
		}
	}

	onMount(() => {
		connectWS();
		// Load initial server state for consistent counters
		void loadJobStats();
		// Ensure job types are loaded on first render
		void fetchJobTypes();
		// Periodic light resync to prevent drift during bursts of ultra-fast jobs
		statsSyncInterval = setInterval(() => {
			if (connected) void loadJobStats();
		}, 10000);
	});

	onDestroy(() => {
		disconnectWS();
		if (statsSyncInterval) clearInterval(statsSyncInterval);
	});

	function formatDuration(start: Date, end: Date) {
		const ms = end.getTime() - start.getTime();
		if (ms < 1000) return `${ms}ms`;
		if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
		return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
	}

	async function loadJobStats() {
		try {
			const res = await getJobStats();
			if (res.status === 200) {
				// Initialize topic counters from server
				runningByTopic = res.data.running_by_topic || {};
				queuedByTopic = res.data.queued_by_topic || {};
				// Note: we keep stats.activeCount based on our local tracking map
				// since it reflects what we're actually showing in the UI
			}
		} catch (e) {
			console.warn("Failed to load job stats:", e);
		}
	}

	async function fetchJobTypes() {
		jobMgmt.loading = true;
		try {
			const res = await listJobs();
			if (res.status === 200) {
				jobMgmt.types = res.data.items || [];
				jobMgmt.types.forEach((job) => {
					if (!jobMgmt.concurrency[job.id]) jobMgmt.concurrency[job.id] = 5;
				});
			} else {
				showMessage("Failed to fetch job types", "error");
			}
		} catch (e) {
			showMessage("Error fetching job types: " + (e as Error).message, "error");
		} finally {
			jobMgmt.loading = false;
		}
	}

	async function startJobType(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "all" });
			if (res.status === 202) {
				toastState.addToast({ message: res.data.message || `Job ${jobId} started`, type: "success" });
				const count = Number((res.data as any).count ?? 0);
				if (count > 0) {
					const topic = getTopicForJobId(jobId);
					queuedByTopic[topic] = (queuedByTopic[topic] || 0) + count;
				}
				await fetchJobTypes();
			} else {
				toastState.addToast({ message: `Failed to start job ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error starting job: " + (e as Error).message, type: "error" });
		}
	}

	async function _stopJobType(jobId: string) {
		try {
			const res = await stopJobType(jobId);
			if (res.status === 200) {
				toastState.addToast({ message: res.data.message || `Job type ${jobId} stopped`, type: "success" });
			} else {
				toastState.addToast({ message: `Failed to stop job type ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error stopping job type: " + (e as Error).message, type: "error" });
		}
	}

	async function rescanAll(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "all" });
			if (res.status === 202) {
				toastState.addToast({ message: res.data.message || `Rescan all for ${jobId} started`, type: "success" });
				const count = Number((res.data as any).count ?? 0);
				if (count > 0) {
					const topic = getTopicForJobId(jobId);
					queuedByTopic[topic] = (queuedByTopic[topic] || 0) + count;
				}
			} else {
				toastState.addToast({ message: `Failed to start rescan all for ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error starting rescan: " + (e as Error).message, type: "error" });
		}
	}

	async function rescanMissing(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "missing" });
			if (res.status === 202) {
				toastState.addToast({ message: res.data.message || `Rescan missing for ${jobId} started`, type: "success" });
				const count = Number((res.data as any).count ?? 0);
				if (count > 0) {
					const topic = getTopicForJobId(jobId);
					queuedByTopic[topic] = (queuedByTopic[topic] || 0) + count;
				}
			} else {
				toastState.addToast({ message: `Failed to start rescan missing for ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error starting rescan: " + (e as Error).message, type: "error" });
		}
	}

	async function updateConcurrency(jobId: string, value: number) {
		jobMgmt.concurrency[jobId] = value;
		try {
			const res = await updateJobTypeConcurrency(jobId, { concurrency: value });
			if (res.status === 200) {
				toastState.addToast({ message: res.data.message || `Concurrency for ${jobId} set to ${value}`, type: "success" });
			} else {
				toastState.addToast({ message: `Failed to update concurrency for ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error updating concurrency: " + (e as Error).message, type: "error" });
		}
	}

	// -------- Deterministic bootstrap (snapshot + cursor) --------
	async function bootstrapState() {
		// 1) Snapshot counters/active from server (authoritative base)
		const snap = await getJobsSnapshot();
		if (snap.status === 200) {
			const d = snap.data || {};
			runningByTopic = d.running_by_topic || {};
			queuedByTopic = d.queued_by_topic || {};
			// Rebuild active entries with minimal info
			jobTracking.active.clear();
			if (Array.isArray(d.active)) {
				for (const a of d.active) {
					jobTracking.active.set(a.id, {
						jobId: a.id,
						type: a.topic,
						startTime: new Date(),
						progress: 0,
						step: a.status || "Running"
					});
				}
			}
			stats.activeCount = jobTracking.active.size;
		}
		// 2) Establish current cursor by asking events/since with tiny limit
		const since = await getEventsSince({ cursor: lastCursor, limit: 1 });
		if (since.status === 200) {
			const nc = Number(since.data?.nextCursor ?? 0);
			if (nc && nc > lastCursor) lastCursor = nc;
		}
		backfilledOnce = true;
	}
</script>

<svelte:head>
	<title>Jobs - Admin</title>
</svelte:head>

<div class="admin-page">
	<header class="page-header">
		<div class="header-left">
			<a href="/admin" class="back-link">
				<MaterialIcon iconName="arrow_back" />
			</a>
			<div>
				<h1>Job Manager</h1>
				<p class="subtitle">Monitor and manage background jobs</p>
			</div>
		</div>
		<div class="connection-status" class:connected>
			<span class="status-dot"></span>
			{connected ? "Connected" : "Disconnected"}
		</div>
	</header>

	<!-- Controls Section -->
	<section class="content-section">
		<div class="section-header">
			<h2>Scheduler Controls</h2>
		</div>
		<div class="controls-grid">
			<Button onclick={startScheduler} class="control-button">
				<MaterialIcon iconName="play_arrow" />
				Start Scheduler
			</Button>
			<Button onclick={shutdownScheduler} class="control-button warning">
				<MaterialIcon iconName="stop" />
				Shutdown Scheduler
			</Button>
			{#if connected}
				<Button onclick={disconnectWS} class="control-button">
					<MaterialIcon iconName="link_off" />
					Disconnect WebSocket
				</Button>
			{:else}
				<Button onclick={connectWS} class="control-button">
					<MaterialIcon iconName="link" />
					Connect WebSocket
				</Button>
			{/if}
		</div>
	</section>
	<!-- Job Types Management -->
	<section class="content-section">
		<div class="section-header">
			<h2>Job Types</h2>
			<Button onclick={fetchJobTypes} disabled={jobMgmt.loading}>
				<MaterialIcon iconName="refresh" />
				Refresh
			</Button>
		</div>
		{#if jobMgmt.loading}
			<div class="loading">Loading job types...</div>
		{:else if jobMgmt.types.length === 0}
			<div class="empty-state">No job types available</div>
		{:else}
			<div class="job-types-grid">
				{#each jobMgmt.types as job (job.id)}
					<div class="job-type-card">
						<div class="job-type-header">
							<div class="job-type-info">
								<h3>{job.topic}</h3>
								<span class="job-type-status status-{job.status.toLowerCase()}">{job.status}</span>
							</div>
							<div class="job-type-stats">
								<div class="stat-item-small">
									<span class="stat-label-small">Active:</span>
									<span class="stat-value-small">{runningByTopic[getTopicForJobId(job.id)] || 0}</span>
								</div>
								<div class="stat-item-small">
									<span class="stat-label-small">Waiting:</span>
									<span class="stat-value-small">{queuedByTopic[getTopicForJobId(job.id)] || 0}</span>
								</div>
							</div>
						</div>

						<div class="job-type-controls">
							<div class="control-row-small">
								{#if job.status.toLowerCase() === "active" || job.status.toLowerCase() === "running"}
									<Button class="btn-stop btn-full-width" onclick={() => _stopJobType(job.id)}>
										<MaterialIcon iconName="stop" />
										Stop
									</Button>
								{:else}
									<Button class="btn-start btn-full-width" onclick={() => startJobType(job.id)}>
										<MaterialIcon iconName="play_arrow" />
										Start
									</Button>
								{/if}
							</div>
							<div class="control-row-small">
								<Button class="btn-rescan" onclick={() => rescanAll(job.id)}>
									<MaterialIcon iconName="refresh" />
									Rescan All
								</Button>
								<Button class="btn-missing" onclick={() => rescanMissing(job.id)}>
									<MaterialIcon iconName="search" />
									Rescan Missing
								</Button>
							</div>
						</div>
						<div class="concurrency-control-small">
							<label for="concurrency-{job.id}">
								<MaterialIcon iconName="tune" />
								Concurrency:
							</label>
							<div class="number-input-wrapper">
								<input
									id="concurrency-{job.id}"
									type="number"
									min="1"
									max="20"
									value={jobMgmt.concurrency[job.id] || 5}
									oninput={(e) => updateConcurrency(job.id, parseInt((e.target as HTMLInputElement).value))}
								/>
								<div class="spinner-buttons">
									<button
										type="button"
										class="spinner-btn spinner-up"
										onclick={() => {
											const currentVal = jobMgmt.concurrency[job.id] || 5;
											if (currentVal < 20) updateConcurrency(job.id, currentVal + 1);
										}}
									>
										<MaterialIcon iconName="keyboard_arrow_up" />
									</button>
									<button
										type="button"
										class="spinner-btn spinner-down"
										onclick={() => {
											const currentVal = jobMgmt.concurrency[job.id] || 5;
											if (currentVal > 1) updateConcurrency(job.id, currentVal - 1);
										}}
									>
										<MaterialIcon iconName="keyboard_arrow_down" />
									</button>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Statistics -->
	<section class="content-section">
		<div class="stats-grid">
			<div class="stat-card active">
				<MaterialIcon iconName="pending" />
				<div class="stat-content">
					<span class="stat-value">{stats.activeCount}</span>
					<span class="stat-label">Active</span>
				</div>
			</div>
			<div class="stat-card completed">
				<MaterialIcon iconName="check_circle" />
				<div class="stat-content">
					<span class="stat-value">{stats.completedCount}</span>
					<span class="stat-label">Completed</span>
				</div>
			</div>
			<div class="stat-card failed">
				<MaterialIcon iconName="error" />
				<div class="stat-content">
					<span class="stat-value">{stats.failedCount}</span>
					<span class="stat-label">Failed</span>
				</div>
			</div>
			<div class="stat-card total">
				<MaterialIcon iconName="analytics" />
				<div class="stat-content">
					<span class="stat-value">{stats.totalProcessed}</span>
					<span class="stat-label">Total Processed</span>
				</div>
			</div>
		</div>
	</section>

	<!-- Active Jobs -->
	{#if jobTracking.active.size > 0}
		<section class="content-section">
			<div class="section-header">
				<h2>Active Jobs</h2>
				<span class="badge">{jobTracking.active.size}</span>
			</div>
			<div class="jobs-list">
				{#each Array.from(jobTracking.active.values()) as job}
					<div class="job-card active">
						<div class="job-header">
							<div class="job-info">
								<MaterialIcon iconName="image" />
								<div>
									<div class="job-title">{job.filename || job.imageId}</div>
									<div class="job-meta">{job.type} • Started {job.startTime.toLocaleTimeString()}</div>
								</div>
							</div>
							<div class="job-progress-value">{job.progress}%</div>
						</div>
						<div class="progress-bar">
							<div class="progress-fill" style="width: {job.progress}%"></div>
						</div>
						<div class="job-step">{job.step}</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Completed Jobs -->
	{#if jobTracking.completed.length > 0}
		<section class="content-section">
			<div class="section-header">
				<h2>Completed Jobs</h2>
				<span class="badge">{jobTracking.completed.length}</span>
			</div>
			<div class="jobs-list">
				{#each jobTracking.completed.slice(0, 10) as job}
					<div class="job-card completed">
						<div class="job-header">
							<div class="job-info">
								<MaterialIcon iconName="check_circle" />
								<div>
									<div class="job-title">{job.filename || job.imageId}</div>
									<div class="job-meta">
										{job.type} • {formatDuration(job.startTime, job.endTime)}
									</div>
								</div>
							</div>
							<span class="job-time">{job.endTime.toLocaleTimeString()}</span>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Failed Jobs -->
	{#if jobTracking.failed.length > 0}
		<section class="content-section">
			<div class="section-header error">
				<h2>Failed Jobs</h2>
				<span class="badge error">{jobTracking.failed.length}</span>
			</div>
			<div class="jobs-list">
				{#each jobTracking.failed.slice(0, 10) as job}
					<div class="job-card failed">
						<div class="job-header">
							<div class="job-info">
								<MaterialIcon iconName="error" />
								<div>
									<div class="job-title">{job.filename || job.imageId}</div>
									<div class="job-meta error">{job.error}</div>
								</div>
							</div>
							<span class="job-time">{job.endTime.toLocaleTimeString()}</span>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}
</div>

<style lang="scss">
	:global(.admin-page) {
		padding: 2rem;
		overflow-y: auto;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		gap: 1rem;

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
	}

	.connection-status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		background: var(--imag-90);
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--imag-40);

		.status-dot {
			width: 8px;
			height: 8px;
			border-radius: 50%;
			background: var(--imag-60);
		}

		&.connected {
			background: #d1fae5;
			color: #065f46;

			.status-dot {
				background: #10b981;
				animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
			}
		}
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

		&.error h2 {
			color: var(--imag-error-color, #ef4444);
		}
	}

	.badge {
		padding: 0.25rem 0.625rem;
		background: var(--imag-80);
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--imag-text-color);

		&.error {
			background: #fee2e2;
			color: #991b1b;
		}
	}

	.controls-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	:global(.control-button) {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.875rem 1.25rem;
		border-radius: 8px;
		font-size: 0.95rem;
		font-weight: 500;
		background-color: var(--imag-80);
		color: var(--imag-text-color);
		transition: background-color 0.2s;

		&:hover {
			background-color: var(--imag-70);
		}
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.stat-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.25rem;
		background: var(--imag-90);
		border-radius: 12px;
		border: 2px solid var(--imag-80);
		transition: border-color 0.2s;

		&:hover {
			border-color: var(--imag-70);
		}

		&.active {
			border-color: #d97706;
			background: #78350f;
			color: #fef3c7;

			.stat-value {
				color: #fef3c7;
			}

			.stat-label {
				color: #fcd34d;
			}
		}

		&.completed {
			border-color: #059669;
			background: #064e3b;
			color: #d1fae5;

			.stat-value {
				color: #d1fae5;
			}

			.stat-label {
				color: #6ee7b7;
			}
		}

		&.failed {
			border-color: #dc2626;
			background: #7f1d1d;
			color: #fee2e2;

			.stat-value {
				color: #fee2e2;
			}

			.stat-label {
				color: #fca5a5;
			}
		}

		&.total {
			.stat-value {
				color: var(--imag-text-color);
			}

			.stat-label {
				color: var(--imag-40);
			}
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

	.jobs-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.job-card {
		padding: 1.25rem;
		background: var(--imag-100);
		border-radius: 12px;
		border-left: 4px solid var(--imag-80);

		&.active {
			border-left-color: #fbbf24;
		}

		&.completed {
			border-left-color: #10b981;
		}

		&.failed {
			border-left-color: #ef4444;
		}
	}

	.job-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.75rem;
	}

	.job-info {
		display: flex;
		align-items: center;
		gap: 0.875rem;
	}

	.job-title {
		font-weight: 600;
		font-size: 0.95rem;
	}

	.job-meta {
		font-size: 0.825rem;
		color: var(--imag-40);
		margin-top: 0.25rem;

		&.error {
			color: #991b1b;
		}
	}

	.job-progress-value {
		font-weight: 700;
		font-size: 1.125rem;
		color: var(--imag-primary);
	}

	.progress-bar {
		height: 8px;
		background: var(--imag-90);
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 0.5rem;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--imag-primary), var(--imag-accent-color));
		transition: width 0.3s ease;
	}

	.job-step {
		font-size: 0.825rem;
		color: var(--imag-40);
	}

	.job-time {
		font-size: 0.825rem;
		color: var(--imag-40);
	}

	.loading,
	.empty-state {
		text-align: center;
		padding: 2rem;
		color: var(--imag-40);
	}

	.job-types-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 1rem;
	}

	.job-type-card {
		background: var(--imag-100);
		border-radius: 12px;
		border: 2px solid var(--imag-80);
		box-shadow: 0 1px 0 0 var(--imag-90) inset;
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.job-type-header {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.job-type-info {
		display: flex;
		justify-content: space-between;
		align-items: center;

		h3 {
			margin: 0;
			font-size: 1.125rem;
			font-weight: 600;
			color: var(--imag-text-color);
		}
	}

	.job-type-status {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;

		&.status-active {
			background: #78350f;
			color: #fef3c7;
			border: 1px solid #d97706;
		}

		&.status-completed {
			background: #064e3b;
			color: #d1fae5;
			border: 1px solid #059669;
		}

		&.status-failed {
			background: #7f1d1d;
			color: #fee2e2;
			border: 1px solid #dc2626;
		}

		&.status-pending {
			background: #1e3a8a;
			color: #dbeafe;
			border: 1px solid #3b82f6;
		}
	}

	.job-type-stats {
		display: flex;
		gap: 1.5rem;
	}

	.stat-item-small {
		display: flex;
		gap: 0.5rem;
		font-size: 0.875rem;

		.stat-label-small {
			color: var(--imag-40);
			font-weight: 500;
		}

		.stat-value-small {
			color: var(--imag-text-color);
			font-weight: 600;
		}
	}

	.job-type-controls {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.control-row-small {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;

		:global(button) {
			padding: 0.625rem 0.875rem;
			font-size: 0.875rem;
			border-radius: 6px;
			background-color: var(--imag-80);
			color: var(--imag-text-color);
			transition: background-color 0.2s;
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 0.375rem;

			&:hover {
				background-color: var(--imag-70);
			}
		}
	}

	:global(.btn-start) {
		background: var(--imag-primary) !important;
		color: #fff !important;
		border: 1px solid color-mix(in oklab, var(--imag-primary), #000 15%);
	}
	:global(.btn-start:hover) {
		background: color-mix(in oklab, var(--imag-primary), #000 12%);
	}

	:global(.btn-stop) {
		background: #7f1d1d !important;
		border: 1px solid #dc2626;
		color: #fee2e2 !important;
	}
	:global(.btn-stop:hover) {
		background: #991b1b !important;
	}

	:global(.btn-full-width) {
		grid-column: 1 / -1;
	}

	:global(.btn-rescan) {
		background: #1e3a8a !important;
		border: 1px solid #3b82f6;
		color: #dbeafe !important;
		white-space: nowrap;
	}
	:global(.btn-rescan:hover) {
		background: #1d4ed8 !important;
	}

	:global(.btn-missing) {
		background: #064e3b !important;
		border: 1px solid #059669;
		color: #d1fae5 !important;
		white-space: nowrap;
	}
	:global(.btn-missing:hover) {
		background: #065f46 !important;
	}

	.concurrency-control-small {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem;
		background: var(--imag-90);
		border-radius: 0.5rem;

		label {
			display: flex;
			align-items: center;
			gap: 0.5rem;
			font-size: 0.9rem;
			font-weight: 500;
			color: var(--imag-text-color);
			white-space: nowrap;
		}

		.number-input-wrapper {
			position: relative;
			display: flex;
			align-items: stretch;
		}

		input {
			padding: 0.5rem 0.5rem 0.5rem 0.75rem;
			border: 1px solid var(--imag-80);
			border-radius: 0.4rem 0 0 0.4rem;
			background: var(--imag-100);
			color: var(--imag-text-color);
			font-size: 0.9rem;
			width: 3.5rem;
			text-align: center;

			&:focus {
				outline: none;
				border-color: var(--imag-primary);
				z-index: 1;
			}

			/* Hide native spinner */
			&::-webkit-outer-spin-button,
			&::-webkit-inner-spin-button {
				-webkit-appearance: none;
				margin: 0;
			}

			&[type="number"] {
				-moz-appearance: textfield;
				appearance: textfield;
				opacity: 1;
			}
		}

		.spinner-buttons {
			display: flex;
			flex-direction: column;
			border: 1px solid var(--imag-80);
			border-left: none;
			border-radius: 0 0.4rem 0.4rem 0;
			overflow: hidden;
		}

		.spinner-btn {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 1.75rem;
			height: 50%;
			padding: 0;
			background: var(--imag-90);
			color: var(--imag-text-color);
			border: none;
			cursor: pointer;
			transition: background-color 0.15s;

			&:hover {
				background: var(--imag-80);
			}

			&:active {
				background: var(--imag-70);
			}

			:global(.material-icons) {
				font-size: 1.1rem;
			}
		}

		.spinner-up {
			border-bottom: 1px solid var(--imag-70);
		}
	}
</style>
