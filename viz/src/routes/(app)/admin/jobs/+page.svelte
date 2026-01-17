<script lang="ts">
	import { onMount } from "svelte";
	import { fade } from "svelte/transition";
	import AdminRouteShell from "$lib/components/admin/AdminRouteShell.svelte";
	import IconButton from "$lib/components/IconButton.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { jobsState } from "$lib/states/jobs.svelte";

	function formatDuration(start: Date, end: Date) {
		const ms = end.getTime() - start.getTime();
		if (ms < 1000) return `${ms}ms`;
		if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
		return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
	}

	onMount(() => {
		void jobsState.init();
	});
</script>

<svelte:head>
	<title>Jobs - Admin</title>
</svelte:head>

<AdminRouteShell
	heading="Job Manager"
	description="Monitor and manage background jobs"
>
	{#snippet actions()}
		<div class="connection-status" class:connected={jobsState.connected}>
			<span class="status-dot"></span>
			<span class="status-text">
				{jobsState.connected ? "WebSocket Connected" : "WebSocket Disconnected"}
			</span>

			<div class="status-action">
				{#if jobsState.connected}
					<IconButton
						iconName="link_off"
						variant="small"
						onclick={() => jobsState.disconnectWS()}
						class="btn-header-action disconnect"
					>
						Go Offline
					</IconButton>
				{:else}
					<IconButton
						iconName="link"
						variant="small"
						onclick={() => jobsState.connectWS()}
						class="btn-header-action connect"
					>
						Go Online
					</IconButton>
				{/if}
			</div>
		</div>
	{/snippet}

	<div class="jobs-dashboard">
		<div class="side-column">
			<!-- Job Types Management -->
			<section class="dashboard-section workers-section">
				<div class="section-header-compact">
					<h3>Available Workers</h3>
					<IconButton
						iconName="refresh"
						variant="small"
						onclick={() => jobsState.fetchJobTypes()}
						disabled={jobsState.workers.loading}
					></IconButton>
				</div>

				{#if jobsState.workers.loading}
					<div class="side-loading">
						<div class="spinner-small"></div>
						<span>Updating registry...</span>
					</div>
				{:else}
					<div class="job-types-list">
						{#each jobsState.workers.types as job}
							<div class="worker-card">
								<div class="worker-header">
									<div class="worker-id">
										<span class="worker-name"
											>{jobsState.getTopicForJobType(job.name)}</span
										>
										<span
											class="worker-dot {(jobsState.runningByTopic[
												jobsState.getTopicForJobType(job.name)
											] || 0) > 0
												? 'active'
												: 'idle'}"
										></span>
									</div>
									<div class="worker-stats">
										<span class="stat-badge running" title="Active">
											{jobsState.runningByTopic[
												jobsState.getTopicForJobType(job.name)
											] || 0}
										</span>
										<span class="stat-badge queued" title="Queued">
											{jobsState.queuedByTopic[
												jobsState.getTopicForJobType(job.name)
											] || 0}
										</span>
									</div>
								</div>

								<div class="worker-actions">
									<IconButton
										iconName="refresh"
										variant="small"
										onclick={() => jobsState.rescanAll(job.name)}
										title="Rescan All"
									></IconButton>
									<IconButton
										iconName="search"
										variant="small"
										onclick={() => jobsState.rescanMissing(job.name)}
										title="Rescan Missing"
									></IconButton>
								</div>

								<div class="concurrency-row">
									<span class="concurrency-label">Concurrency</span>
									<div class="concurrency-input">
										<button
											class="step-btn"
											onclick={() =>
												jobsState.setWorkerConcurrency(
													job.name,
													Math.max(
														1,
														(jobsState.workers.concurrency[job.name] || 5) - 1
													)
												)}>-</button
										>
										<span class="step-value"
											>{jobsState.workers.concurrency[job.name] || 5}</span
										>
										<button
											class="step-btn"
											onclick={() =>
												jobsState.setWorkerConcurrency(
													job.name,
													Math.min(
														50,
														(jobsState.workers.concurrency[job.name] || 5) + 1
													)
												)}>+</button
										>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</section>
		</div>

		<div class="main-column">
			<!-- Statistics -->
			<section class="dashboard-section stats-section">
				<div class="stats-grid">
					<div class="stat-card active" in:fade>
						<div class="stat-icon">
							<MaterialIcon iconName="pending" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{jobsState.stats.activeCount}</span>
							<span class="stat-label">Active Jobs</span>
						</div>
					</div>
					<div class="stat-card completed" in:fade={{ delay: 100 }}>
						<div class="stat-icon">
							<MaterialIcon iconName="check_circle" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{jobsState.stats.completedCount}</span>
							<span class="stat-label">Completed</span>
						</div>
					</div>
					<div class="stat-card failed" in:fade={{ delay: 200 }}>
						<div class="stat-icon">
							<MaterialIcon iconName="error" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{jobsState.stats.failedCount}</span>
							<span class="stat-label">Failed</span>
						</div>
					</div>
					<div class="stat-card total" in:fade={{ delay: 300 }}>
						<div class="stat-icon">
							<MaterialIcon iconName="analytics" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{jobsState.stats.totalProcessed}</span>
							<span class="stat-label">Total Processed</span>
						</div>
					</div>
				</div>
			</section>

			<!-- Active Jobs -->
			<section class="dashboard-section prominent">
				<div class="section-header">
					<div class="header-title">
						<MaterialIcon iconName="bolt" class="title-icon highlight" />
						<h2>Realtime Processing</h2>
					</div>
					<span class="badge highlight">{jobsState.activeJobs.length}</span>
				</div>

				<div class="jobs-list active-list">
					{#each jobsState.activeJobs as job (job.uid)}
						<div class="job-card active">
							<div class="job-card-main">
								<div class="job-info">
									<div class="job-icon-wrapper">
										<MaterialIcon iconName="image" />
									</div>
									<div class="job-details">
										<div
											class="job-title"
											title={job.filename || job.image_uid || job.uid}
										>
											{job.filename || job.image_uid || job.uid}
										</div>
										<div class="job-meta">
											<span class="job-type-tag">{job.type || job.topic}</span>
											<span class="separator">•</span>
											<span class="job-time">
												Started {job.startTime
													? job.startTime.toLocaleTimeString()
													: "just now"}
											</span>
										</div>
									</div>
								</div>
								<div class="job-progress-section">
									<div class="progress-info">
										<span class="job-step">{job.step || "Initializing..."}</span
										>
										<span class="progress-value">{job.progress || 0}%</span>
									</div>
									<div class="progress-bar-container">
										<div
											class="progress-bar-fill"
											style="width: {job.progress || 0}%"
										>
											<div class="progress-shimmer"></div>
										</div>
									</div>
								</div>
							</div>
						</div>
					{/each}

					{#if jobsState.activeJobs.length === 0}
						<div class="empty-state-placeholder">
							<MaterialIcon iconName="magic_button" class="empty-icon" />
							<p>No active jobs at the moment</p>
						</div>
					{/if}
				</div>
			</section>

			<div class="history-grid">
				<!-- Completed Jobs -->
				<section class="dashboard-section history-section">
					<div class="section-header">
						<div class="header-title">
							<MaterialIcon iconName="history" />
							<h3>Recent Success</h3>
						</div>
					</div>
					<div class="jobs-list mini-list">
						{#each jobsState.completedJobs.slice(0, 8) as job (job.uid)}
							<div class="mini-job-card success">
								<MaterialIcon iconName="check_circle" class="status-icon" />
								<div class="mini-job-info">
									<span class="mini-job-title"
										>{job.filename || job.image_uid || job.uid}</span
									>
									<span class="mini-job-meta">
										{formatDuration(
											new Date(job.startTime),
											new Date(job.endTime ?? job.startTime)
										)}
									</span>
								</div>
							</div>
						{/each}
						{#if jobsState.completedJobs.length === 0}
							<div class="mini-empty">No history yet</div>
						{/if}
					</div>
				</section>

				<!-- Failed Jobs -->
				<section class="dashboard-section history-section">
					<div class="section-header">
						<div class="header-title">
							<MaterialIcon iconName="running_with_errors" class="error-text" />
							<h3>Recent Failures</h3>
						</div>
					</div>
					<div class="jobs-list mini-list">
						{#each jobsState.failedJobs.slice(0, 8) as job (job.uid)}
							<div class="mini-job-card failure">
								<MaterialIcon iconName="error" class="status-icon" />
								<div class="mini-job-info">
									<span class="mini-job-title"
										>{job.filename || job.image_uid || job.uid}</span
									>
									<span class="mini-job-error" title={job.error}
										>{job.error || "Unknown error"}</span
									>
								</div>
							</div>
						{/each}
						{#if jobsState.failedJobs.length === 0}
							<div class="mini-empty">System stable</div>
						{/if}
					</div>
				</section>
			</div>
		</div>
	</div>
</AdminRouteShell>

<style lang="scss">
	.jobs-dashboard {
		display: grid;
		grid-template-columns: 20rem 1fr;
		gap: 1.5rem;
		align-items: start;

		@media (max-width: 64rem) {
			grid-template-columns: 1fr;
		}
	}

	.main-column {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.side-column {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		position: sticky;
		top: 1.5rem;
	}

	.dashboard-section {
		background: var(--imag-100);
		border-radius: 1rem;
		padding: 1.5rem;
		border: 1px solid var(--imag-90);
		contain: layout;

		&.stats-section {
			min-height: 7rem;
		}

		&.prominent {
			height: 35rem; // Fixed height to prevent layout shifts
			display: flex;
			flex-direction: column;
		}

		&.history-section {
			height: 28rem; // Fixed height
			display: flex;
			flex-direction: column;
		}

		&.workers-section {
			min-height: 20rem;
		}
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1.5rem;
		flex-shrink: 0;

		.header-title {
			display: flex;
			align-items: center;
			gap: 0.75rem;

			h2 {
				margin: 0;
				font-size: 1.25rem;
				font-weight: 700;
				color: var(--imag-text-color);
			}

			:global(.title-icon) {
				font-size: 1.5rem;
			}
		}
	}

	.section-header-compact {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;

		h3 {
			margin: 0;
			font-size: 1rem;
			font-weight: 600;
			color: var(--imag-text-color);
		}
	}

	.badge {
		padding: 0.25rem 0.75rem;
		background: var(--imag-80);
		border-radius: 1.25rem;
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--imag-text-color);

		&.highlight {
			background: var(--imag-primary);
			color: white;
		}
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
		gap: 1rem;
	}

	.stat-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: var(--imag-90);
		border-radius: 0.75rem;
		border: 1px solid var(--imag-80);
		transition:
			border-color 0.2s,
			background-color 0.2s;

		&:hover {
			border-color: var(--imag-70);
			background-color: var(--imag-80);
		}

		.stat-icon {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 3rem;
			height: 3rem;
			border-radius: 0.75rem;
			background: var(--imag-80);
			color: var(--imag-text-color);
		}

		.stat-content {
			display: flex;
			flex-direction: column;
		}

		.stat-value {
			font-size: 1.5rem;
			font-weight: 800;
			line-height: 1.2;
		}

		.stat-label {
			font-size: 0.75rem;
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 0.05em;
			color: var(--imag-40);
		}

		&.active {
			background: linear-gradient(
				135deg,
				rgba(var(--imag-primary-rgb), 0.1) 0%,
				transparent 100%
			);
			border: 1px solid rgba(var(--imag-primary-rgb), 0.2);

			&:hover {
				background: rgba(var(--imag-primary-rgb), 0.15);
				border-color: var(--imag-primary);
			}

			.stat-icon {
				background: var(--imag-primary);
				color: white;
			}
			.stat-value {
				color: var(--imag-primary);
			}
		}

		&.completed {
			&:hover {
				background: rgba(16, 185, 129, 0.05);
				border-color: #10b981;
			}
			.stat-icon {
				background: #10b981;
				color: white;
			}
			.stat-value {
				color: #10b981;
			}
		}

		&.failed {
			&:hover {
				background: rgba(220, 38, 38, 0.05);
				border-color: var(--imag-error-color);
			}
			.stat-icon {
				background: var(--imag-error-color);
				color: white;
			}
			.stat-value {
				color: var(--imag-error-color);
			}
		}
	}

	.jobs-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		flex: 1;
		overflow-y: auto;
		padding-right: 0.5rem;
		min-height: 0; // Crucial for flexbox children with overflow
	}

	.job-card {
		background: var(--imag-90);
		border-radius: 0.75rem;
		overflow: hidden;
		border: 1px solid var(--imag-80);
		flex-shrink: 0;

		&.active {
			border-left: 0.25rem solid var(--imag-primary);
		}

		.job-card-main {
			padding: 1.25rem;
		}
	}

	.job-info {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1rem;

		.job-icon-wrapper {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 2.5rem;
			height: 2.5rem;
			background: var(--imag-80);
			border-radius: 0.625rem;
			color: var(--imag-primary);
		}

		.job-details {
			flex: 1;
			min-width: 0;

			.job-title {
				font-weight: 700;
				font-size: 1rem;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
				color: var(--imag-text-color);
			}

			.job-meta {
				display: flex;
				align-items: center;
				gap: 0.5rem;
				font-size: 0.75rem;
				color: var(--imag-40);

				.job-type-tag {
					background: var(--imag-80);
					padding: 0.125rem 0.5rem;
					border-radius: 0.25rem;
					font-weight: 600;
					text-transform: uppercase;
				}
			}
		}
	}

	.job-progress-section {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;

		.progress-info {
			display: flex;
			justify-content: space-between;
			align-items: center;
			font-size: 0.875rem;

			.job-step {
				font-weight: 600;
				color: var(--imag-text-color);
			}

			.progress-value {
				font-weight: 800;
				color: var(--imag-primary);
			}
		}
	}

	.progress-bar-container {
		height: 0.625rem;
		background: var(--imag-80);
		border-radius: 0.3125rem;
		overflow: hidden;
		position: relative;
	}

	.progress-bar-fill {
		height: 100%;
		background: var(--imag-primary);
		border-radius: 0.3125rem;
		transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		position: relative;
		overflow: hidden;
	}

	.progress-shimmer {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(
			90deg,
			transparent 0%,
			rgba(255, 255, 255, 0.2) 50%,
			transparent 100%
		);
		animation: shimmer 2s infinite linear;
	}

	@keyframes shimmer {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	.history-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
		contain: layout;

		@media (max-width: 48rem) {
			grid-template-columns: 1fr;
		}
	}

	.history-section {
		h3 {
			margin: 0;
			font-size: 0.9rem;
			font-weight: 700;
			text-transform: uppercase;
			letter-spacing: 0.05em;
			color: var(--imag-40);
			flex-shrink: 0;
		}
	}

	.mini-job-card {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: var(--imag-90);
		border-radius: 0.625rem;
		border: 1px solid var(--imag-80);
		flex-shrink: 0;
		transition:
			border-color 0.2s,
			background-color 0.2s;

		&:hover {
			border-color: var(--imag-70);
			background-color: var(--imag-80);
		}

		:global(.status-icon) {
			font-size: 1.25rem;
		}

		.mini-job-info {
			display: flex;
			flex-direction: column;
			min-width: 0;

			.mini-job-title {
				font-size: 0.8125rem;
				font-weight: 600;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}

			.mini-job-meta,
			.mini-job-error {
				font-size: 0.75rem;
				color: var(--imag-40);
			}

			.mini-job-error {
				color: var(--imag-error-color);
				font-style: italic;
			}
		}

		&.success {
			:global(.status-icon) {
				color: #10b981;
			}
		}

		&.failure {
			:global(.status-icon) {
				color: var(--imag-error-color);
			}
		}
	}

	:global(.btn-connect),
	:global(.btn-disconnect) {
		width: auto;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		font-weight: 700;
	}

	:global(.btn-connect) {
		background: var(--imag-primary) !important;
		color: white !important;
	}
	:global(.btn-disconnect) {
		background: var(--imag-80) !important;
		color: var(--imag-text-color) !important;
	}

	.job-types-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.worker-card {
		background: var(--imag-90);
		border-radius: 0.75rem;
		padding: 1rem;
		border: 1px solid var(--imag-80);
		transition: border-color 0.2s;

		&:hover {
			border-color: var(--imag-70);
		}

		.worker-header {
			display: flex;
			justify-content: space-between;
			align-items: flex-start;
			margin-bottom: 1rem;

			.worker-id {
				display: flex;
				align-items: center;
				gap: 0.5rem;

				.worker-name {
					font-weight: 700;
					font-size: 0.9rem;
				}

				.worker-dot {
					width: 0.5rem;
					height: 0.5rem;
					border-radius: 50%;

					&.active {
						background: #10b981;
					}
					&.idle {
						background: var(--imag-60);
					}
				}
			}

			.worker-stats {
				display: flex;
				gap: 0.25rem;

				.stat-badge {
					padding: 0.125rem 0.375rem;
					border-radius: 0.25rem;
					font-size: 0.75rem;
					font-weight: 700;

					&.running {
						background: rgba(var(--imag-primary-rgb), 0.2);
						color: var(--imag-primary);
					}
					&.queued {
						background: var(--imag-80);
						color: var(--imag-40);
					}
				}
			}
		}

		.worker-actions {
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 0.5rem;
			margin-bottom: 1rem;
		}

		.concurrency-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding-top: 0.75rem;
			border-top: 1px solid var(--imag-80);

			.concurrency-label {
				font-size: 0.75rem;
				font-weight: 600;
				color: var(--imag-40);
			}

			.concurrency-input {
				display: flex;
				align-items: center;
				background: var(--imag-80);
				border-radius: 0.375rem;
				overflow: hidden;

				.step-btn {
					width: 1.5rem;
					height: 1.5rem;
					border: none;
					background: transparent;
					color: var(--imag-text-color);
					cursor: pointer;
					&:hover {
						background: var(--imag-70);
					}
				}

				.step-value {
					width: 2rem;
					text-align: center;
					font-size: 0.875rem;
					font-weight: 700;
				}
			}
		}
	}

	.connection-status {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.25rem 0.25rem 0.25rem 1rem;
		border-radius: 1.875rem;
		font-size: 0.75rem;
		font-weight: 700;
		background: var(--imag-90);
		color: var(--imag-40);
		border: 1px solid var(--imag-80);

		.status-dot {
			width: 0.5rem;
			height: 0.5rem;
			border-radius: 50%;
			background: var(--imag-60);
		}

		.status-text {
			flex: 1;
		}

		.status-action {
			display: flex;
		}

		&.connected {
			background: rgba(16, 185, 129, 0.1);
			color: #10b981;
			border-color: rgba(16, 185, 129, 0.3);

			.status-dot {
				background: #10b981;
				animation: pulse 2s infinite;
			}
		}
	}

	@keyframes pulse {
		0% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
		100% {
			opacity: 1;
		}
	}

	.empty-state-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex: 1;
		color: var(--imag-40);
		text-align: center;
		padding: 2rem;
		border: 2px dashed var(--imag-80);
		border-radius: 0.75rem;
		margin: 1rem 0;

		:global(.empty-icon) {
			font-size: 3rem;
			margin-bottom: 1rem;
			opacity: 0.3;
		}

		p {
			font-weight: 600;
			margin: 0;
			opacity: 0.5;
		}
	}

	.mini-empty {
		padding: 2rem;
		text-align: center;
		font-size: 0.75rem;
		color: var(--imag-60);
		font-style: italic;
		border: 1px dashed var(--imag-80);
		border-radius: 0.5rem;
		margin-top: 0.5rem;
	}

	.side-loading {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 2rem;
		color: var(--imag-40);
		font-size: 0.875rem;
	}

	.spinner-small {
		width: 1rem;
		height: 1rem;
		border: 2px solid var(--imag-80);
		border-top-color: var(--imag-primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.separator {
		margin: 0 0.25rem;
	}
</style>
