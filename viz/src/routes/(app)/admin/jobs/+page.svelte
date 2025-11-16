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
	import type { WorkerJob, JobInfo } from "$lib/api";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

	type UiJob = WorkerJob & {
		filename?: string;
		progress?: number;
		step?: string;
		startTime: Date;
		endTime?: Date | string;
		error?: string;
	};

	function getErrorMessage(e: unknown) {
		if (e instanceof Error) return e.message;
		try {
			return String(e);
		} catch {
			return "Unknown error";
		}
	}

	function getCountFromResData(data: unknown) {
		if (!data || typeof data !== "object") return 0;
		const d = data as Record<string, unknown>;
		const c = d.count;
		if (typeof c === "number") return c;
		if (typeof c === "string") {
			const n = Number(c);
			return Number.isNaN(n) ? 0 : n;
		}
		return 0;
	}

	// Lightweight state using backend's new shapes (uids, image_uid, topic)
	let connected = false;
	let wsClient: WSClient | null = null;
	let lastCursor = 0;

	let stats = {
		activeCount: 0,
		completedCount: 0,
		failedCount: 0,
		totalProcessed: 0
	};

	let activeJobs: UiJob[] = [];
	let completedJobs: UiJob[] = [];
	let failedJobs: UiJob[] = [];

	let runningByTopic: Record<string, number> = {};
	let queuedByTopic: Record<string, number> = {};

	let jobMgmt = {
		loading: false,
		types: [] as JobInfo[],
		concurrency: {} as Record<string, number>
	};

	function showMessage(message: string, type: "success" | "error" | "info" = "info") {
		toastState.addToast({ message, type });
	}

	function getTopicForJobType(jobType: string) {
		const info = jobMgmt.types.find((t) => t.id === jobType);
		if (info && typeof info.topic === "string" && info.topic.length > 0) {
			return info.topic;
		}

		return jobType;
	}

	// Scheduler controls
	async function startScheduler() {
		try {
			const res = await apiStartScheduler();
			if (res.status === 200) showMessage(res.data?.message ?? "Scheduler started", "success");
			else showMessage(`Start failed (${res.status})`, "error");
		} catch (e) {
			showMessage("Start failed: " + getErrorMessage(e), "error");
		}
	}

	async function shutdownScheduler() {
		try {
			const res = await apiShutdownScheduler();
			if (res.status === 200) showMessage(res.data?.message ?? "Scheduler shutdown", "success");
			else showMessage(`Shutdown failed (${res.status})`, "error");
		} catch (e) {
			showMessage("Shutdown failed: " + getErrorMessage(e), "error");
		}
	}

	function connectWS() {
		if (wsClient) {
			return;
		}

		wsClient = createWSConnection(
			async (event: string, data: unknown) => {
				const payload = data && typeof data === "object" ? (data as Record<string, unknown>) : {};
				switch (event) {
					case "connected":
						connected = true;
						if (payload.clientId) {
							showMessage(`Connected (client ${String(payload.clientId).slice(0, 8)})`, "success");
						}

						await bootstrapState();
						break;
					case "job-started": {
						const uid = typeof payload.uid === "string" ? payload.uid : typeof payload.id === "string" ? payload.id : undefined;
						if (!uid) {
							return;
						}

						const progress = typeof payload.progress === "number" ? payload.progress : 0;
						const newJob: UiJob = {
							uid,
							// prefer explicit type/topic when provided, fall back to the other
							type:
								typeof payload.type === "string" ? payload.type : typeof payload.topic === "string" ? payload.topic : "unknown",
							topic:
								typeof payload.topic === "string" ? payload.topic : typeof payload.type === "string" ? payload.type : "unknown",
							status: typeof payload.status === "string" ? payload.status : "running",
							enqueued_at: typeof payload.enqueued_at === "string" ? payload.enqueued_at : new Date().toISOString(),
							image_uid: typeof payload.image_uid === "string" ? payload.image_uid : undefined,
							filename: typeof payload.filename === "string" ? payload.filename : undefined,
							progress,
							startTime: new Date()
						};

						activeJobs = [newJob, ...activeJobs].slice(0, 200);
						stats.activeCount = activeJobs.length;

						const topic = getTopicForJobType(newJob.type || newJob.topic || "unknown");
						runningByTopic[topic] = (runningByTopic[topic] || 0) + 1;
						if ((queuedByTopic[topic] || 0) > 0) {
							queuedByTopic[topic] = Math.max(0, queuedByTopic[topic] - 1);
						}

						break;
					}
					case "job-progress": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						const imageUid = typeof payload.image_uid === "string" ? payload.image_uid : undefined;
						if (!uid && !imageUid) {
							return;
						}

						const progress = typeof payload.progress === "number" ? payload.progress : undefined;
						const step = typeof payload.step === "string" ? payload.step : undefined;

						// I hate everything about this
						activeJobs = activeJobs.map((j) =>
							(j.uid && uid && j.uid === uid) || (j.image_uid && imageUid && j.image_uid === imageUid)
								? { ...j, ...(progress !== undefined ? { progress } : {}), ...(step ? { step } : {}) }
								: j
						);

						break;
					}
					case "job-completed": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						const imageUid = typeof payload.image_uid === "string" ? payload.image_uid : undefined;

						let removed: UiJob | null = null;
						activeJobs = activeJobs.filter((j) => {
							if ((j.uid && uid && j.uid === uid) || (imageUid && j.image_uid === imageUid)) {
								removed = { ...j, ...(payload as Record<string, unknown>), endTime: new Date() } as UiJob;
								return false;
							}
							return true;
						});

						if (!removed) {
							removed = {
								uid: uid ?? String(Math.random()).slice(2),
								type:
									typeof payload.type === "string" ? payload.type : typeof payload.topic === "string" ? payload.topic : "unknown",
								topic:
									typeof payload.topic === "string" ? payload.topic : typeof payload.type === "string" ? payload.type : "unknown",
								status: typeof payload.status === "string" ? payload.status : "completed",
								enqueued_at: typeof payload.enqueued_at === "string" ? payload.enqueued_at : new Date().toISOString(),
								image_uid: imageUid ?? undefined,
								endTime: new Date(),
								startTime: new Date()
							} as UiJob;
						}

						completedJobs = [removed, ...completedJobs].slice(0, 50);
						stats.completedCount++;
						stats.totalProcessed++;
						stats.activeCount = activeJobs.length;

						const topic = getTopicForJobType(removed.type || removed.topic || "unknown");
						runningByTopic[topic] = Math.max(0, (runningByTopic[topic] || 0) - 1);
						break;
					}
					case "job-failed": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						const imageUid = typeof payload.image_uid === "string" ? payload.image_uid : undefined;
						let removed: UiJob | null = null;

						// fuck this
						activeJobs = activeJobs.filter((j) => {
							if ((j.uid && uid && j.uid === uid) || (imageUid && j.image_uid === imageUid)) {
								removed = { ...j, ...(payload as Record<string, unknown>), endTime: new Date() } as UiJob;
								return false;
							}
							return true;
						});

						if (!removed) {
							removed = {
								uid: uid ?? String(Math.random()).slice(2),
								type:
									typeof payload.type === "string" ? payload.type : typeof payload.topic === "string" ? payload.topic : "unknown",
								topic:
									typeof payload.topic === "string" ? payload.topic : typeof payload.type === "string" ? payload.type : "unknown",
								status: typeof payload.status === "string" ? payload.status : "failed",
								enqueued_at: typeof payload.enqueued_at === "string" ? payload.enqueued_at : new Date().toISOString(),
								image_uid: imageUid ?? undefined,
								endTime: new Date(),
								startTime: new Date()
							} as UiJob;
						}

						failedJobs = [removed, ...failedJobs].slice(0, 50);
						stats.failedCount++;
						stats.activeCount = activeJobs.length;

						const topic = getTopicForJobType(removed.type || removed.topic || "unknown");
						runningByTopic[topic] = Math.max(0, (runningByTopic[topic] || 0) - 1);
						break;
					}
				}
			},
			(err) => {
				console.error("WebSocket error:", err);
				connected = false;
				showMessage("WebSocket connection error", "error");
			},
			() => (connected = true),
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

	let statsSyncInterval: ReturnType<typeof setInterval> | null = null;
	fetch("localhost:7770/images?limit=100", {
		method: "GET",
		headers: {
			"Content-Type": "application/json"
		},
		credentials: "include"
	})
		.then((res) => res.json())
		.then((data) => {
			console.log("Fetched images:", data);
		})
		.catch((error) => {
			console.error("Error fetching images:", error);
		});

	onMount(() => {
		connectWS();
		void loadJobStats();
		void fetchJobTypes();
		statsSyncInterval = setInterval(() => void loadJobStats(), 10000);
	});

	onDestroy(() => {
		disconnectWS();
		if (statsSyncInterval) {
			clearInterval(statsSyncInterval);
			statsSyncInterval = null;
		}
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
				const d = res.data as Record<string, unknown> | undefined;
				runningByTopic = d && typeof d.running_by_topic === "object" ? (d.running_by_topic as Record<string, number>) : {};
				queuedByTopic = d && typeof d.queued_by_topic === "object" ? (d.queued_by_topic as Record<string, number>) : {};
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
				const d = res.data;
				jobMgmt.types = Array.isArray(d.items) ? (d.items as JobInfo[]) : [];
				jobMgmt.types.forEach((job) => {
					if (!jobMgmt.concurrency[job.id]) {
						jobMgmt.concurrency[job.id] = 5;
					}
				});
			} else {
				showMessage("Failed to fetch job types", "error");
			}
		} catch (e) {
			showMessage("Error fetching job types: " + getErrorMessage(e), "error");
		} finally {
			jobMgmt.loading = false;
		}
	}

	async function startJobType(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "all" });
			if (res.status === 202) {
				toastState.addToast({
					message: res.data.message || `Job ${jobId} started`,
					type: "success"
				});

				const count = getCountFromResData(res.data);
				if (count > 0) {
					const topic = getTopicForJobType(jobId);
					queuedByTopic[topic] = (queuedByTopic[topic] || 0) + count;
				}
				await fetchJobTypes();
			} else {
				toastState.addToast({ message: `Failed to start job ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error starting job: " + getErrorMessage(e), type: "error" });
		}
	}

	async function _stopJobType(jobId: string) {
		try {
			const res = await stopJobType(jobId);
			if (res.status === 200) {
				toastState.addToast({
					message: res.data.message || `Job type ${jobId} stopped`,
					type: "success"
				});
			} else {
				toastState.addToast({ message: `Failed to stop job type ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error stopping job type: " + getErrorMessage(e), type: "error" });
		}
	}

	async function rescanAll(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "all" });
			if (res.status === 202) {
				toastState.addToast({
					message: res.data.message || `Rescan all for ${jobId} started`,
					type: "success"
				});

				const count = getCountFromResData(res.data);
				if (count > 0) {
					const topic = getTopicForJobType(jobId);
					queuedByTopic[topic] = (queuedByTopic[topic] || 0) + count;
				}
			} else {
				toastState.addToast({ message: `Failed to start rescan all for ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error starting rescan: " + getErrorMessage(e), type: "error" });
		}
	}

	async function rescanMissing(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "missing" });
			if (res.status === 202) {
				toastState.addToast({
					message: res.data.message || `Rescan missing for ${jobId} started`,
					type: "success"
				});
				const count = getCountFromResData(res.data);
				if (count > 0) {
					const topic = getTopicForJobType(jobId);
					queuedByTopic[topic] = (queuedByTopic[topic] || 0) + count;
				}
			} else {
				toastState.addToast({ message: `Failed to start rescan missing for ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error starting rescan: " + getErrorMessage(e), type: "error" });
		}
	}

	async function updateConcurrency(jobId: string, value: number) {
		jobMgmt.concurrency[jobId] = value;
		try {
			const res = await updateJobTypeConcurrency(jobId, { concurrency: value });
			if (res.status === 200) {
				toastState.addToast({
					message: res.data.message || `Concurrency for ${jobId} set to ${value}`,
					type: "success"
				});
			} else {
				toastState.addToast({ message: `Failed to update concurrency for ${jobId}`, type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error updating concurrency: " + getErrorMessage(e), type: "error" });
		}
	}

	// Bootstrap using server snapshot
	async function bootstrapState() {
		const snap = await getJobsSnapshot();
		if (snap.status === 200) {
			const d = snap.data as Record<string, unknown> | undefined;
			runningByTopic = d && typeof d.running_by_topic === "object" ? (d.running_by_topic as Record<string, number>) : {};
			queuedByTopic = d && typeof d.queued_by_topic === "object" ? (d.queued_by_topic as Record<string, number>) : {};
			// fucking hell
			if (Array.isArray(d?.active)) {
				activeJobs = (d!.active as unknown[]).map((a) => {
					if (a && typeof a === "object") {
						const obj = a as Record<string, unknown>;
						return {
							uid: obj.id ? String(obj.id) : String(Math.random()).slice(2),
							type: typeof obj.topic === "string" ? obj.topic : "unknown",
							topic: typeof obj.topic === "string" ? obj.topic : typeof obj.type === "string" ? obj.type : "unknown",
							status: typeof obj.status === "string" ? obj.status : "running",
							enqueued_at: typeof obj.enqueued_at === "string" ? obj.enqueued_at : new Date().toISOString(),
							startTime: new Date()
						} as UiJob;
					}
					return {
						uid: String(a),
						type: String(a),
						topic: String(a),
						status: "running",
						enqueued_at: new Date().toISOString(),
						startTime: new Date()
					} as UiJob;
				});
			} else {
				activeJobs = [];
			}
			stats.activeCount = activeJobs.length;
		}

		// establish simple cursor
		const since = await getEventsSince({ cursor: lastCursor, limit: 1 });
		if (since.status === 200) {
			const nc = Number(since.data.nextCursor) ?? 0;
			if (nc && nc > lastCursor) {
				lastCursor = nc;
			}
		}
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
									<span class="stat-value-small">{runningByTopic[getTopicForJobType(job.id)] || 0}</span>
								</div>
								<div class="stat-item-small">
									<span class="stat-label-small">Waiting:</span>
									<span class="stat-value-small">{queuedByTopic[getTopicForJobType(job.id)] || 0}</span>
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
	{#if activeJobs.length > 0}
		<section class="content-section">
			<div class="section-header">
				<h2>Active Jobs</h2>
				<span class="badge">{activeJobs.length}</span>
			</div>
			<div class="jobs-list">
				{#each activeJobs as job}
					<div class="job-card active">
						<div class="job-header">
							<div class="job-info">
								<MaterialIcon iconName="image" />
								<div>
									<div class="job-title">{job.filename || job.image_uid || job.uid}</div>
									<div class="job-meta">
										{job.type || job.topic} • Started {job.startTime ? new Date(job.startTime).toLocaleTimeString() : ""}
									</div>
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
	{#if completedJobs.length > 0}
		<section class="content-section">
			<div class="section-header">
				<h2>Completed Jobs</h2>
				<span class="badge">{completedJobs.length}</span>
			</div>
			<div class="jobs-list">
				{#each completedJobs.slice(0, 10) as job}
					<div class="job-card completed">
						<div class="job-header">
							<div class="job-info">
								<MaterialIcon iconName="check_circle" />
								<div>
									<div class="job-title">{job.filename || job.image_uid || job.uid}</div>
									<div class="job-meta">
										{job.type || job.topic} • {formatDuration(new Date(job.startTime), new Date(job.endTime ?? job.startTime))}
									</div>
								</div>
							</div>
							<span class="job-time">{job.endTime ? new Date(job.endTime).toLocaleTimeString() : ""}</span>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Failed Jobs -->
	{#if failedJobs.length > 0}
		<section class="content-section">
			<div class="section-header error">
				<h2>Failed Jobs</h2>
				<span class="badge error">{failedJobs.length}</span>
			</div>
			<div class="jobs-list">
				{#each failedJobs.slice(0, 10) as job}
					<div class="job-card failed">
						<div class="job-header">
							<div class="job-info">
								<MaterialIcon iconName="error" />
								<div>
									<div class="job-title">{job.filename || job.image_uid || job.uid}</div>
									<div class="job-meta error">{job.error || job.error_msg}</div>
								</div>
							</div>
							<span class="job-time">{job.endTime ? new Date(job.endTime).toLocaleTimeString() : ""}</span>
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
