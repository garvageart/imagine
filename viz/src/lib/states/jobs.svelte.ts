import {
	getJobStats,
	getJobsSnapshot,
	listAvailableWorkers,
	createJob,
	updateJobTypeConcurrency,
	getEventsSince,
	type WorkerJob,
	type WorkerInfo
} from "$lib/api";
import { createWSConnection, type WSClient } from "$lib/api/websocket";
import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

export type UiJob = WorkerJob & {
	filename?: string;
	progress?: number;
	step?: string;
	startTime: Date;
	endTime?: Date | string;
	error?: string;
};

class JobsState {
	// WebSocket state
	connected = $state(false);
	wsClient: WSClient | null = null;
	lastCursor = 0;

	// Counters & Stats
	stats = $state({
		activeCount: 0,
		completedCount: 0,
		failedCount: 0,
		totalProcessed: 0
	});

	// Job Lists
	activeJobs = $state<UiJob[]>([]);
	completedJobs = $state<UiJob[]>([]);
	failedJobs = $state<UiJob[]>([]);

	// Topic tracking
	runningByTopic = $state<Record<string, number>>({});
	queuedByTopic = $state<Record<string, number>>({});

	// Worker info
	workers = $state({
		loading: false,
		types: [] as WorkerInfo[],
		concurrency: {} as Record<string, number>
	});

	private initialLoadDone = false;

	constructor() {
		// No-op, initialization happens in init()
	}

	async init() {
		if (this.initialLoadDone && this.wsClient) {
			return;
		}

		this.connectWS();

		if (!this.initialLoadDone) {
			await Promise.all([
				this.loadJobStats(),
				this.fetchJobTypes()
			]);
			this.initialLoadDone = true;
		}
	}

	connectWS() {
		if (this.wsClient) {
			return;
		}

		this.wsClient = createWSConnection(
			async (event: string, data: unknown) => {
				const payload =
					data && typeof data === "object"
						? (data as Record<string, unknown>)
						: {};
				switch (event) {
					case "connected":
						this.connected = true;
						await this.bootstrapState();
						break;
					case "job-started": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						if (!uid) {
							return;
						}

						const topic = typeof payload.topic === "string" ? payload.topic : "unknown";
						const type = typeof payload.type === "string" ? payload.type : topic;

						const newJob: UiJob = {
							uid,
							type,
							topic,
							status: "running",
							enqueued_at: typeof payload.enqueued_at === "string"
								? payload.enqueued_at
								: new Date().toISOString(),
							image_uid: typeof payload.image_uid === "string"
								? payload.image_uid
								: undefined,
							filename: typeof payload.filename === "string"
								? payload.filename
								: undefined,
							progress: 0,
							startTime: new Date()
						};

						this.activeJobs = [newJob, ...this.activeJobs].slice(0, 200);
						this.stats.activeCount = this.activeJobs.length;

						const displayTopic = this.getTopicForJobType(topic);
						this.runningByTopic[displayTopic] = (this.runningByTopic[displayTopic] || 0) + 1;
						if ((this.queuedByTopic[displayTopic] || 0) > 0) {
							this.queuedByTopic[displayTopic] = Math.max(0, this.queuedByTopic[displayTopic] - 1);
						}
						break;
					}
					case "job-progress": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						if (!uid) {
							return;
						}

						const progress = typeof payload.progress === "number" ? payload.progress : undefined;
						const step = typeof payload.step === "string" ? payload.step : undefined;

						this.activeJobs = this.activeJobs.map((j) =>
							j.uid === uid
								? {
									...j,
									...(progress !== undefined ? { progress } : {}),
									...(step ? { step } : {})
								}
								: j
						);
						break;
					}
					case "job-completed": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						if (!uid) {
							return;
						}

						let removed: UiJob | null = null;
						this.activeJobs = this.activeJobs.filter((j) => {
							if (j.uid === uid) {
								removed = {
									...j,
									status: "completed",
									endTime: new Date()
								};
								return false;
							}
							return true;
						});

						if (!removed) {
							removed = {
								uid,
								type: typeof payload.type === "string" ? payload.type : "unknown",
								topic: typeof payload.topic === "string" ? payload.topic : "unknown",
								status: "completed",
								image_uid: typeof payload.image_uid === "string" ? payload.image_uid : undefined,
								endTime: new Date(),
								startTime: new Date()
							} as UiJob;
						}

						this.completedJobs = [removed, ...this.completedJobs].slice(0, 50);
						this.stats.completedCount++;
						this.stats.totalProcessed++;
						this.stats.activeCount = this.activeJobs.length;

						const displayTopic = this.getTopicForJobType(removed.topic || removed.type || "unknown");
						this.runningByTopic[displayTopic] = Math.max(0, (this.runningByTopic[displayTopic] || 0) - 1);
						break;
					}
					case "job-failed": {
						const uid = typeof payload.uid === "string" ? payload.uid : undefined;
						if (!uid) {
							return;
						}

						let removed: UiJob | null = null;
						this.activeJobs = this.activeJobs.filter((j) => {
							if (j.uid === uid) {
								removed = {
									...j,
									status: "failed",
									endTime: new Date(),
									error: typeof payload.error === "string" ? payload.error : "Unknown error"
								};
								return false;
							}
							return true;
						});

						if (!removed) {
							removed = {
								uid,
								type: typeof payload.type === "string" ? payload.type : "unknown",
								topic: typeof payload.topic === "string" ? payload.topic : "unknown",
								status: "failed",
								image_uid: typeof payload.image_uid === "string" ? payload.image_uid : undefined,
								endTime: new Date(),
								startTime: new Date(),
								error: typeof payload.error === "string" ? payload.error : "Unknown error"
							} as UiJob;
						}

						this.failedJobs = [removed, ...this.failedJobs].slice(0, 50);
						this.stats.failedCount++;
						this.stats.activeCount = this.activeJobs.length;

						const displayTopic = this.getTopicForJobType(removed.topic || removed.type || "unknown");
						this.runningByTopic[displayTopic] = Math.max(0, (this.runningByTopic[displayTopic] || 0) - 1);
						break;
					}
				}
			},
			(err) => {
				console.error("WebSocket error:", err);
				this.connected = false;
				toastState.addToast({ message: "WebSocket connection error", type: "error" });
			},
			() => (this.connected = true),
			(code: number, reason: string) => {
				this.connected = false;
				if (code !== 1000 && code !== 1001 && code !== 1005) {
					toastState.addToast({
						message: `WebSocket disconnected (${code}): ${reason}`,
						type: "error"
					});
				}
			}
		);
	}

	disconnectWS() {
		if (this.wsClient) {
			this.wsClient.close();
			this.wsClient = null;
			this.connected = false;
		}
	}

	async loadJobStats() {
		try {
			const res = await getJobStats();
			if (res.status === 200) {
				const d = res.data;
				this.runningByTopic = d.running_by_topic;
				this.queuedByTopic = d.queued_by_topic;
			}
		} catch (e) {
			console.warn("Failed to load job stats:", e);
		}
	}

	async fetchJobTypes() {
		this.workers.loading = true;
		try {
			const res = await listAvailableWorkers();
			if (res.status === 200) {
				const d = res.data;
				this.workers.types = Array.isArray(d.items) ? (d.items as WorkerInfo[]) : [];
				this.workers.types.forEach((job) => {
					if (!this.workers.concurrency[job.name]) {
						this.workers.concurrency[job.name] = job.concurrency || 5;
					}
				});
			} else {
				toastState.addToast({ message: "Failed to fetch job types", type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Error fetching job types: " + (e instanceof Error ? e.message : String(e)), type: "error" });
		} finally {
			this.workers.loading = false;
		}
	}

	getTopicForJobType(jobType: string) {
		const info = this.workers.types.find((t) => t.name === jobType);
		if (
			info &&
			typeof info.display_name === "string" &&
			info.display_name.length > 0
		) {
			return info.display_name;
		}
		return jobType;
	}

	async bootstrapState() {
		const snap = await getJobsSnapshot();
		if (snap.status === 200) {
			const d = snap.data;
			this.runningByTopic = d.running_by_topic || {};
			this.queuedByTopic = d.queued_by_topic || {};

			if (Array.isArray(d.active)) {
				const newActive = d.active.map((obj) => {
					return {
						uid: obj.uid,
						image_uid: obj.image_uid || undefined,
						type: obj.type,
						topic: obj.topic,
						status: obj.status,
						enqueued_at: obj.enqueued_at,
						startTime: new Date(),
						progress: 0
					} as UiJob;
				});

				// Only update if UIDs changed to avoid flashing
				const oldUids = this.activeJobs.map(j => j.uid).sort().join(',');
				const newUids = newActive.map(j => j.uid).sort().join(',');

				if (oldUids !== newUids) {
					this.activeJobs = newActive;
				}
			} else {
				if (this.activeJobs.length > 0) {
					this.activeJobs = [];
				}
			}
			this.stats.activeCount = this.activeJobs.length;
		}

		const since = await getEventsSince({ cursor: this.lastCursor, limit: 1 });
		if (since.status === 200) {
			const nc = Number(since.data.nextCursor) ?? 0;
			if (nc && nc > this.lastCursor) {
				this.lastCursor = nc;
			}
		}
	}

	async rescanAll(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "all" });
			if (res.status === 202) {
				toastState.addToast({
					message: res.data.message || `Rescan all for ${jobId} started`,
					type: "success"
				});

				const count = this.getCountFromResData(res.data);
				if (count > 0) {
					const topic = this.getTopicForJobType(jobId);
					this.queuedByTopic[topic] = (this.queuedByTopic[topic] || 0) + count;
				}
			} else {
				toastState.addToast({
					message: `Failed to start rescan all for ${jobId}`,
					type: "error"
				});
			}
		} catch (e) {
			toastState.addToast({
				message: "Error starting rescan: " + (e instanceof Error ? e.message : String(e)),
				type: "error"
			});
		}
	}

	async rescanMissing(jobId: string) {
		try {
			const res = await createJob({ type: jobId, command: "missing" });
			if (res.status === 202) {
				toastState.addToast({
					message: res.data.message || `Rescan missing for ${jobId} started`,
					type: "success"
				});

				const count = this.getCountFromResData(res.data);
				if (count > 0) {
					const topic = this.getTopicForJobType(jobId);
					this.queuedByTopic[topic] = (this.queuedByTopic[topic] || 0) + count;
				}
			} else {
				toastState.addToast({
					message: `Failed to start rescan missing for ${jobId}`,
					type: "error"
				});
			}
		} catch (e) {
			toastState.addToast({
				message: "Error starting rescan: " + (e instanceof Error ? e.message : String(e)),
				type: "error"
			});
		}
	}

	async setWorkerConcurrency(jobId: string, value: number) {
		this.workers.concurrency[jobId] = value;
		try {
			const res = await updateJobTypeConcurrency(jobId, { concurrency: value });
			if (res.status !== 200) {
				toastState.addToast({
					message: `Failed to update concurrency for ${jobId}`,
					type: "error"
				});
				await this.fetchJobTypes();
			}
		} catch (e) {
			toastState.addToast({
				message: "Error updating concurrency: " + (e instanceof Error ? e.message : String(e)),
				type: "error"
			});
			await this.fetchJobTypes();
		}
	}

	private getCountFromResData(data: unknown) {
		if (!data || typeof data !== "object") {
			return 0;
		}

		const d = data as Record<string, unknown>;
		const count = d.count;
		if (typeof count === "number") {
			return count;
		}

		if (typeof count === "string") {
			const n = Number(count);
			return Number.isNaN(n) ? 0 : n;
		}

		return 0;
	}
}

export const jobsState = new JobsState();
