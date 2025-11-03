<script lang="ts">
	import { user } from "$lib/states/index.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { startScheduler as apiStartScheduler, shutdownScheduler as apiShutdownScheduler } from "$lib/api";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

	function showMessage(message: string, type: "success" | "error" | "info" = "info") {
		toastState.addToast({ message, type });
	}

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
</script>

<svelte:head>
	<title>Admin</title>
</svelte:head>

<div class="admin-page">
	<header class="page-header">
		<div>
			<h1>Admin Panel</h1>
			<p class="subtitle">Manage system settings and monitor operations</p>
		</div>
	</header>

	<section class="content-section">
		<div class="section-header">
			<MaterialIcon iconName="settings" />
			<h2>Scheduler Controls</h2>
		</div>
		<div class="controls-grid">
			<Button onclick={startScheduler}>
				<MaterialIcon iconName="play_arrow" />
				Start Scheduler
			</Button>
			<Button onclick={shutdownScheduler}>
				<MaterialIcon iconName="stop" />
				Shutdown Scheduler
			</Button>
		</div>
	</section>

	<section class="content-section">
		<div class="section-header">
			<MaterialIcon iconName="dashboard" />
			<h2>Quick Navigation</h2>
		</div>
		<div class="nav-grid">
			<a href="/admin/jobs" class="nav-card">
				<div class="nav-icon">
					<MaterialIcon iconName="work" />
				</div>
				<div class="nav-content">
					<h3>Job Manager</h3>
					<p>Monitor and manage background jobs</p>
				</div>
				<MaterialIcon iconName="arrow_forward" />
			</a>
			<a href="/admin/events" class="nav-card">
				<div class="nav-icon">
					<MaterialIcon iconName="analytics" />
				</div>
				<div class="nav-content">
					<h3>Event Monitor</h3>
					<p>View SSE metrics and event history</p>
				</div>
				<MaterialIcon iconName="arrow_forward" />
			</a>
		</div>
	</section>

	<section class="content-section">
		<div class="section-header">
			<MaterialIcon iconName="info" />
			<h2>System Information</h2>
		</div>
		<div class="info-grid">
			<div class="info-item">
				<span class="info-label">Username:</span>
				<span class="info-value">{user.data?.username || "Unknown"}</span>
			</div>
			<div class="info-item">
				<span class="info-label">User Role:</span>
				<span class="info-value">{user.data?.role || "Unknown"}</span>
			</div>
		</div>
	</section>
</div>

<style lang="scss">
	:global(.admin-page) {
		padding: 2rem;
		overflow-y: auto;
	}

	.page-header {
		margin-bottom: 2rem;

		h1 {
			margin: 0;
			font-size: 2rem;
			font-weight: 600;
		}

		.subtitle {
			margin: 0.5rem 0 0 0;
			color: var(--imag-40);
			font-size: 1rem;
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

	.controls-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12em, 1fr));
		gap: 1rem;
	}

	:global(.controls-grid button) {
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

	.nav-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1rem;
	}

	.nav-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: var(--imag-90);
		border-radius: 1em;
		border: 2px solid transparent;
		text-decoration: none;
		color: var(--imag-text-color);
		transition:
			border-color 0.2s,
			background-color 0.2s;

		&:hover {
			border-color: var(--imag-60);
			background-color: var(--imag-90);
		}

		.nav-icon {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 3em;
			height: 3em;
			background: var(--imag-60);
			color: white;
			border-radius: 12px;
		}

		.nav-content {
			flex: 1;

			h3 {
				margin: 0 0 0.25rem 0;
				font-size: 1.125rem;
				font-weight: 600;
			}

			p {
				margin: 0;
				color: var(--imag-40);
				font-size: 0.875rem;
			}
		}
	}

	.info-grid {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.info-item {
		display: flex;
		justify-content: left;
		padding: 0.75rem;
		background: var(--imag-90);
		border-radius: 0.5rem;
		gap: 0.3rem;

		.info-label {
			font-weight: 500;
			color: var(--imag-40);
		}

		.info-value {
			font-weight: 600;
		}
	}
</style>
