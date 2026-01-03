<script lang="ts">
	import { invalidateAll } from "$app/navigation";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import ProgressBar from "$lib/components/ProgressBar.svelte";
	import AdminRouteShell from "$lib/components/admin/AdminRouteShell.svelte";
	import { formatBytes, formatSeconds } from "$lib/utils/images";
	import { invalidateViz } from "$lib/views/views.svelte.js";
	import { Duration } from "luxon";
	import { onDestroy } from "svelte";

	let { data } = $props();

	let lastUpdated = $state(new Date());

	let systemInfo = $derived({
		version: window.__APP_VERSION__,
		// uptime is now handled by formattedLiveUptime
		activeConnections: data.wsStats?.connectedClients ?? 0,
		goroutines: data.systemStats?.num_goroutine ?? 0,
		allocMemory: data.systemStats?.alloc_memory
			? formatBytes(data.systemStats.alloc_memory)
			: "Unknown",
		sysMemory: data.systemStats?.sys_memory
			? formatBytes(data.systemStats.sys_memory)
			: "Unknown"
	});

	let databaseInfo = $derived({
		connections: data.dbStats?.active_connections ?? 0,
		size: data.dbStats?.db_size_bytes
			? formatBytes(data.dbStats.db_size_bytes)
			: "Unknown",
		users: data.dbStats?.user_count ?? 0,
		images: data.dbStats?.image_count ?? 0
	});

	let storageInfo = $derived({
		totalUsed: data.systemStats?.storage_used_bytes
			? formatBytes(data.systemStats.storage_used_bytes)
			: "Unknown",
		availableSystemSpace: data.systemStats?.total_system_space_bytes
			? formatBytes(data.systemStats.total_system_space_bytes)
			: "Unknown",
		totalSystemSpace: data.systemStats?.total_system_space_bytes
			? formatBytes(data.systemStats.total_system_space_bytes)
			: "Unknown",
		path: data.systemStats?.storage_path ?? "Unknown",
		cacheSize: data.cacheStatus
			? formatBytes(data.cacheStatus.size)
			: "Unknown",
		cacheItems: data.cacheStatus?.items ?? 0
	});

	let liveUptimeSeconds = $derived(data.systemStats?.uptime_seconds || 0);
	let formattedLiveUptime = $derived(formatSeconds(liveUptimeSeconds));

	$effect(() => {
		liveUptimeSeconds = data.systemStats?.uptime_seconds || 0;
	});

	$effect(() => {
		const interval = setInterval(() => {
			liveUptimeSeconds++;
		}, 1000);

		return () => clearInterval(interval);
	});

	$effect(() => {
		const interval = setInterval(
			() => {
				invalidateViz({ delay: 200 }).then(() => {
					lastUpdated = new Date();
				});
			},
			Duration.fromObject({ seconds: 30 }).as("milliseconds")
		);

		return () => clearInterval(interval);
	});
</script>

<svelte:head>
	<title>Admin Dashboard</title>
</svelte:head>

<AdminRouteShell heading="Dashboard" description="System overview and metrics">
	<div class="dashboard-container">
		<div class="stats-info">
			<span>Last updated:</span>
			<span>
				{lastUpdated.toLocaleString()}
			</span>
		</div>

		<!-- System Overview Section -->
		<section class="section">
			<h3 class="section-title">System Overview</h3>
			<div class="stats-grid">
				<!-- System Version -->
				<div class="stat-card">
					<div class="stat-icon version">
						<MaterialIcon iconName="info" />
					</div>
					<div class="stat-content">
						<span class="stat-value">v{systemInfo.version}</span>
						<span class="stat-label">System Version</span>
					</div>
				</div>

				<!-- Uptime -->
				<div class="stat-card">
					<div class="stat-icon uptime">
						<MaterialIcon iconName="schedule" />
					</div>
					<div class="stat-content">
						<span class="stat-value" id="uptime-value"
							>{formattedLiveUptime}</span
						>
						<span class="stat-label">Uptime</span>
					</div>
				</div>

				<!-- Active Connections (WS) -->
				<div class="stat-card">
					<div class="stat-icon connections">
						<MaterialIcon iconName="hub" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{systemInfo.activeConnections}</span>
						<span class="stat-label">Active Clients</span>
					</div>
				</div>

				<!-- Goroutines -->
				<div class="stat-card">
					<div class="stat-icon goroutines">
						<MaterialIcon iconName="compare_arrows" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{systemInfo.goroutines}</span>
						<span class="stat-label">Goroutines</span>
					</div>
				</div>

				<!-- Allocated Memory -->
				<div class="stat-card">
					<div class="stat-icon alloc-memory">
						<MaterialIcon iconName="memory" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{systemInfo.allocMemory}</span>
						<span class="stat-label">Allocated Memory</span>
					</div>
				</div>

				<!-- System Memory -->
				<div class="stat-card">
					<div class="stat-icon sys-memory">
						<MaterialIcon iconName="memory_alt" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{systemInfo.sysMemory}</span>
						<span class="stat-label">System Memory</span>
					</div>
				</div>
			</div>
		</section>

		<!-- Database Section -->
		<section class="section">
			<h3 class="section-title">Database</h3>
			<div class="stats-grid">
				<!-- DB Connections -->
				<div class="stat-card">
					<div class="stat-icon db">
						<MaterialIcon iconName="database" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{databaseInfo.connections}</span>
						<span class="stat-label">Active Connections</span>
					</div>
				</div>

				<!-- DB Size -->
				<div class="stat-card">
					<div class="stat-icon storage">
						<MaterialIcon iconName="hard_drive" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{databaseInfo.size}</span>
						<span class="stat-label">Database Size</span>
					</div>
				</div>

				<!-- Total Users -->
				<div class="stat-card">
					<div class="stat-icon users">
						<MaterialIcon iconName="group" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{databaseInfo.users}</span>
						<span class="stat-label">Total Users</span>
					</div>
				</div>

				<!-- Total Images -->
				<div class="stat-card">
					<div class="stat-icon images">
						<MaterialIcon iconName="image" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{databaseInfo.images}</span>
						<span class="stat-label">Total Images</span>
					</div>
				</div>
			</div>
		</section>

		<!-- Storage Section -->
		<section class="section">
			<h3 class="section-title">Storage</h3>
			<div class="stats-grid">
				<!-- Percentage of Available Space -->
				<div class="stat-card">
					<div class="stat-icon storage">
						<MaterialIcon iconName="hard_drive" />
					</div>
					<div class="stat-content">
						<span class="stat-value">
							{formatBytes(
								(data.systemStats?.total_system_space_bytes ?? 0) -
									(data.systemStats?.total_available_space_bytes ?? 0)
							)} of {storageInfo.totalSystemSpace}
						</span>
						<span class="stat-label">System Storage</span>
						<div class="progress-bar-wrapper">
							<ProgressBar
								colour="secondary"
								width={100 -
									((data.systemStats?.total_available_space_bytes ?? 0) /
										(data.systemStats?.total_system_space_bytes ?? 1)) *
										100}
							/>
						</div>
					</div>
				</div>

				<!-- Total Storage -->
				<div class="stat-card">
					<div class="stat-icon storage">
						<MaterialIcon iconName="hard_drive" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{storageInfo.totalUsed}</span>
						<span class="stat-label">Viz Storage</span>
					</div>
				</div>

				<div class="stat-seperator"></div>

				<!-- Cache Usage -->
				<div class="stat-card">
					<div class="stat-icon cache">
						<MaterialIcon iconName="memory" />
					</div>
					<div class="stat-content">
						<span class="stat-value">{storageInfo.cacheSize}</span>
						<span class="stat-label"
							>Viz Cache ({storageInfo.cacheItems} items)</span
						>
					</div>
				</div>

				<!-- Storage Path -->
				<div class="stat-card wide">
					<div class="stat-icon storage-path">
						<MaterialIcon iconName="folder" />
					</div>
					<div class="stat-content">
						<span class="stat-value path">{storageInfo.path}</span>
						<span class="stat-label">Storage Path</span>
					</div>
				</div>
			</div>
		</section>
	</div>
</AdminRouteShell>

<style lang="scss">
	.dashboard-container {
		display: flex;
		flex-direction: column;
		gap: 2.5rem;
	}

	.stats-info {
		display: flex;

		span {
			font-size: 0.9rem;
			color: var(--imag-40);
			margin-right: 0.5rem;
		}
	}

	.section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.section-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--imag-text-color);
		margin: 0;
		padding-left: 0.5rem;
		border-left: 4px solid var(--imag-primary);
		line-height: 1.2;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
		gap: 1rem;
	}

	.stat-card {
		background: var(--imag-100);
		border: 1px solid var(--imag-90);
		border-radius: 0.75rem;
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: border-color 0.2s;

		&:hover {
			border-color: var(--imag-80);
		}

		&.wide {
			grid-column: 1 / -1;
		}
	}

	.stat-icon {
		width: 3.5rem;
		height: 3.5rem;
		border-radius: 0.75rem;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--imag-bg-color);
		flex-shrink: 0;

		&.version {
			background-color: #3b82f6;
		}
		&.connections {
			background-color: #10b981;
		}
		&.cache {
			background-color: #f59e0b;
		}
		&.uptime {
			background-color: #8b5cf6;
		}
		&.db {
			background-color: #6366f1;
		}
		&.storage {
			background-color: #ec4899;
		}
		&.users {
			background-color: #14b8a6;
		}
		&.images {
			background-color: #f43f5e;
		}
		&.goroutines {
			background-color: #0d9488;
		}
		&.alloc-memory {
			background-color: #d946af;
		}
		&.sys-memory {
			background-color: #a855f7;
		}
		&.storage-path {
			background-color: #22c55e;
		}

		/* Material Icon scaling override */
		:global(.material-icons) {
			font-size: 1.75rem;
		}
	}

	.stat-content {
		display: flex;
		flex-direction: column;
		min-width: 0; /* prevents overflow flex item issues */
		flex: 1;
	}

	.stat-value {
		font-size: 1.2rem;
		font-weight: 700;
		line-height: 1.2;
		color: var(--imag-text-color);

		&.path {
			font-size: 1rem;
			font-family: var(--imag-code-font);
			word-break: break-all;
		}
	}

	#uptime-value {
		min-width: 10rem;
		display: inline-block;
		text-align: left;
	}

	.progress-bar-wrapper {
		position: relative;
		height: 4px;
		width: 100%;
		margin-top: 1rem;
		overflow: hidden;
	}

	.stat-seperator {
		grid-column: 1 / -1;
		height: 1px;
		width: 100%;
		background-color: var(--imag-90);
		margin: 0.5rem 0;
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--imag-40);
		margin-top: 0.25rem;
	}
</style>
