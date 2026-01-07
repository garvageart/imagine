<script lang="ts">
	import type { ApiKey, Session, SessionUpdate } from "$lib/api";
	import {
		deleteApiKey,
		listApiKeys,
		revokeApiKey,
		updatePassword,
		getSessions,
		deleteSession,
		deleteSessions,
		updateSession,
		getCurrentSession
	} from "$lib/api";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import CreatedApiKeyModal from "$lib/components/modals/CreatedApiKeyModal.svelte";
	import ModalContainer from "$lib/components/modals/ModalContainer.svelte";
	import { modal } from "$lib/states/index.svelte";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import { DateTime } from "luxon";
	import { onMount } from "svelte";
	import CustomSettingsGroup from "./CustomSettingsGroup.svelte";
	import TextInput from "./inputs/TextInput.svelte";

	// Password state
	let passwordState = $state({
		current: "",
		new: "",
		confirm: ""
	});

	let changingPassword = $state(false);

	// API Keys State
	let apiKeys = $state<ApiKey[]>([]);

	// Session State
	let allSessions = $state<Session[]>([]);
	let currentSession = $state<Session | null>(null);
	let loadingData = $state(false);

	// Create Key Modal State
	let showCreateKeyModal = $state(false);

	// Edit Session Modal State
	let showEditSessionModal = $state(false);
	let editingSessionId = $state<string | null>(null);

	interface SessionFormState {
		clientName: string;
		status?: number;
	}
	let sessionUpdatePayload = $state<SessionFormState>({ clientName: "" });

	async function handleChangePassword() {
		if (
			!passwordState.current ||
			!passwordState.new ||
			!passwordState.confirm
		) {
			toastState.addToast({
				message: "All fields are required",
				type: "error"
			});
			return;
		}

		if (passwordState.new !== passwordState.confirm) {
			toastState.addToast({
				message: "New passwords do not match",
				type: "error"
			});
			return;
		}

		if (passwordState.new.length < 8) {
			toastState.addToast({
				message: "Password must be at least 8 characters long",
				type: "error"
			});
			return;
		}

		changingPassword = true;

		const res = await updatePassword({
			current: passwordState.current,
			new: passwordState.new
		});

		if (res.status === 200) {
			toastState.addToast({
				dismissible: true,
				message: "Password updated successfully",
				type: "success"
			});
			passwordState = { current: "", new: "", confirm: "" };
		} else {
			toastState.addToast({
				dismissible: true,
				message: "Failed to update password",
				type: "error"
			});
		}

		changingPassword = false;
	}

	async function loadData() {
		loadingData = true;
		try {
			const [keysRes, allSessionsRes, currentSessionRes] = await Promise.all([
				listApiKeys(),
				getSessions(),
				getCurrentSession() // Use this to identify the current session
			]);

			if (keysRes.status === 200) {
				apiKeys = keysRes.data.items || [];
			}
			if (allSessionsRes.status === 200) {
				allSessions = allSessionsRes.data || [];
			}
			if (currentSessionRes.status === 200) {
				currentSession = currentSessionRes.data;
			}
		} catch (e) {
			console.error("Failed to load security data", e);
		} finally {
			loadingData = false;
		}
	}

	function handleOpenCreateKeyModal() {
		showCreateKeyModal = true;
		modal.show = true;
	}

	function handleCreateKeyModalClose() {
		showCreateKeyModal = false;
		modal.show = false; // Ensure global modal state is reset
	}

	function handleKeyCreated() {
		loadData();
	}

	async function handleDeleteKey(uid: string) {
		if (
			!confirm(
				"Are you sure you want to delete this API Key? This action cannot be undone."
			)
		)
			return;
		try {
			const res = await deleteApiKey(uid);
			if (res.status === 200) {
				apiKeys = apiKeys.filter((k) => k.uid !== uid);
				toastState.addToast({ message: "API Key deleted", type: "success" });
			} else {
				toastState.addToast({
					message: res.data.error || "Failed to delete key",
					type: "error"
				});
			}
		} catch (e) {
			toastState.addToast({ message: "Failed to delete key", type: "error" });
		}
	}

	async function handleRevokeKey(uid: string) {
		if (
			!confirm(
				"Are you sure you want to revoke this API Key? It will stop working immediately."
			)
		)
			return;
		try {
			const res = await revokeApiKey(uid);
			if (res.status === 200) {
				toastState.addToast({ message: "API Key revoked", type: "success" });
				await loadData();
			} else {
				toastState.addToast({ message: "Failed to revoke key", type: "error" });
			}
		} catch (e) {
			toastState.addToast({ message: "Failed to revoke key", type: "error" });
		}
	}

	async function handleDeleteSession(uid: string) {
		if (
			!confirm(
				"Are you sure you want to delete this session? This action cannot be undone."
			)
		)
			return;
		try {
			const res = await deleteSession(uid);
			if (res.status === 200) {
				toastState.addToast({ message: "Session deleted", type: "success" });
				await loadData(); // Refresh all sessions
				if (currentSession && currentSession.uid === uid) {
					window.location.reload(); // Simple logout for now
				}
			} else {
				toastState.addToast({
					message: "Failed to delete session",
					type: "error"
				});
			}
		} catch (e) {
			toastState.addToast({
				message: "Failed to delete session",
				type: "error"
			});
		}
	}

	async function handleDeleteAllSessions() {
		if (
			!confirm(
				"Are you sure you want to delete ALL your sessions? You will be logged out from all devices, including this one."
			)
		)
			return;
		try {
			const res = await deleteSessions();
			if (res.status === 200) {
				toastState.addToast({
					message: "All sessions deleted. Logging out...",
					type: "success"
				});
				// Since all sessions are deleted, including current, force reload/logout
				window.location.reload();
			} else {
				toastState.addToast({
					message: "Failed to delete all sessions",
					type: "error"
				});
			}
		} catch (e) {
			toastState.addToast({
				message: "Failed to delete all sessions",
				type: "error"
			});
		}
	}

	function handleOpenEditSessionModal(session: Session) {
		editingSessionId = session.uid;
		sessionUpdatePayload = {
			clientName: session.client_name ?? "",
			status: session.status ?? undefined
		};
		showEditSessionModal = true;
		modal.show = true;
	}

	async function handleUpdateSession() {
		if (!editingSessionId) return;

		try {
			const payload: SessionUpdate = {
				clientName: sessionUpdatePayload.clientName,
				status: sessionUpdatePayload.status ?? null
			};
			const res = await updateSession(editingSessionId, payload);
			if (res.status === 200) {
				toastState.addToast({ message: "Session updated", type: "success" });
				showEditSessionModal = false;
				modal.show = false;
				await loadData();
			} else {
				toastState.addToast({
					message: "Failed to update session",
					type: "error"
				});
			}
		} catch (e) {
			toastState.addToast({
				message: "Failed to update session",
				type: "error"
			});
		}
	}

	function handleCloseEditSessionModal() {
		showEditSessionModal = false;
		editingSessionId = null;
		modal.show = false;
	}

	onMount(() => {
		loadData();
	});

	function formatDate(dateStr?: string | null) {
		if (!dateStr) return "Never";
		return DateTime.fromJSDate(new Date(dateStr)).toFormat(
			"dd-MM-yyyy HH:mm:ss"
		);
	}

	function isCurrentSession(sessionItem: Session) {
		return currentSession && sessionItem.uid === currentSession.uid;
	}
</script>

{#if showCreateKeyModal && modal.show}
	<CreatedApiKeyModal
		onClose={handleCreateKeyModalClose}
		onSuccess={handleKeyCreated}
	/>
{/if}

{#if showEditSessionModal && modal.show}
	<ModalContainer>
		<div class="edit-session-modal-content">
			<h3>Edit Session</h3>
			<TextInput
				label="Client Name"
				bind:value={sessionUpdatePayload.clientName}
			/>
			<!-- Add status update here if needed -->
			<div class="modal-actions">
				<Button onclick={handleUpdateSession}>Update</Button>
				<Button
					hoverColor="var(--imag-alert-color)"
					onclick={handleCloseEditSessionModal}>Cancel</Button
				>
			</div>
		</div>
	</ModalContainer>
{/if}
{#snippet changePasswordAction()}
	<Button
		onclick={handleChangePassword}
		disabled={changingPassword ||
			!passwordState.current ||
			!passwordState.new ||
			!passwordState.confirm}
	>
		{changingPassword ? "Updating..." : "Update Password"}
	</Button>
{/snippet}

<div class="security-settings">
	<CustomSettingsGroup
		title="Password"
		description="Change your account password."
	>
		{#if passwordState.new && passwordState.confirm}
			{@render changePasswordAction()}
		{/if}

		<TextInput
			label="Current Password"
			type="password"
			bind:value={passwordState.current}
			disabled={changingPassword}
		/>
		<TextInput
			label="New Password"
			type="password"
			bind:value={passwordState.new}
			disabled={changingPassword}
		/>
		<TextInput
			label="Confirm Password"
			type="password"
			bind:value={passwordState.confirm}
			disabled={changingPassword}
		/>
	</CustomSettingsGroup>

	<CustomSettingsGroup
		title="API Keys"
		description="Manage API keys for accessing the Imagine API."
	>
		{#snippet actions()}
			<Button onclick={handleOpenCreateKeyModal}>
				<MaterialIcon iconName="add" />
				<span>Create Key</span>
			</Button>
		{/snippet}

		<div class="keys-list">
			{#each apiKeys as key}
				<div class="key-item" class:revoked={key.revoked}>
					<div class="key-info">
						<div class="key-header">
							<span class="key-name">{key.name}</span>
							{#if key.revoked}
								<span class="status-badge revoked">Revoked</span>
							{:else}
								<span class="status-badge active">Active</span>
							{/if}
						</div>
						<div class="key-meta">
							<span>Created: {formatDate(key.created_at)}</span>
							{#if key.last_used_at}
								<span>Last used: {formatDate(key.last_used_at)}</span>
							{/if}
							<span class="key-desc">{key.description || "No description"}</span
							>
						</div>
					</div>
					<div class="key-actions">
						{#if !key.revoked}
							<Button
								onclick={() => handleRevokeKey(key.uid)}
								title="Revoke Key"
							>
								<MaterialIcon iconName="block" />
							</Button>
						{/if}
						<Button
							onclick={() => handleDeleteKey(key.uid)}
							title="Delete Key"
							hoverColor="var(--imag-alert-color)"
						>
							<MaterialIcon iconName="delete" />
						</Button>
					</div>
				</div>
			{:else}
				{#if !loadingData}
					<p class="empty-state">No API keys found</p>
				{/if}
			{/each}
		</div>
	</CustomSettingsGroup>

	<CustomSettingsGroup
		title="Sessions"
		description="Manage your active sessions across devices."
	>
		{#snippet actions()}
			<Button
				onclick={handleDeleteAllSessions}
				hoverColor="var(--imag-alert-color)"
			>
				<MaterialIcon iconName="logout" />
				<span>Log out All</span>
			</Button>
		{/snippet}

		<div class="sessions-list">
			{#each allSessions as session_item}
				<div
					class="session-item"
					class:current={isCurrentSession(session_item)}
				>
					{#if isCurrentSession(session_item)}
						<span class="current-badge">Current Session</span>
					{/if}
					<div class="session-details">
						<div class="session-row">
							<span class="label">Client:</span>
							<span class="value"
								>{session_item.client_name ||
									session_item.user_agent ||
									"Unknown Client"}</span
							>
						</div>
						<div class="session-row">
							<span class="label">Session ID:</span>
							<span class="value">{session_item.uid}</span>
						</div>
						<div class="session-row">
							<span class="label">Logged In:</span>
							<span class="value">{formatDate(session_item.login_at)}</span>
						</div>
						<div class="session-row">
							<span class="label">Last Active:</span>
							<span class="value">{formatDate(session_item.last_active)}</span>
						</div>
						<div class="session-row">
							<span class="label">Expires:</span>
							<span class="value">{formatDate(session_item.expires_at)}</span>
						</div>
						{#if session_item.client_ip}
							<div class="session-row">
								<span class="label">IP:</span>
								<span class="value">{session_item.client_ip}</span>
							</div>
						{/if}
					</div>
					<div class="session-actions">
						<Button
							onclick={() => handleOpenEditSessionModal(session_item)}
							title="Edit Session"
						>
							<MaterialIcon iconName="edit" />
						</Button>
						<Button
							onclick={() => handleDeleteSession(session_item.uid)}
							title="Delete Session"
							hoverColor="var(--imag-alert-color)"
						>
							<MaterialIcon iconName="delete" />
						</Button>
					</div>
				</div>
			{:else}
				{#if !loadingData}
					<p class="empty-state">No active sessions found.</p>
				{/if}
			{/each}
			{#if loadingData}
				<p class="loading-state">Loading sessions...</p>
			{/if}
		</div>
	</CustomSettingsGroup>
</div>

<style lang="scss">
	.security-settings {
		display: flex;
		flex-direction: column;
		width: 100%;
		height: 100%;
		gap: 3rem;
	}

	.keys-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.key-item,
	.session-item {
		background-color: var(--imag-100);
		border: 1px solid var(--imag-80);
		border-radius: 0.5rem;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.key-item.revoked {
		opacity: 0.6;
	}

	.key-info,
	.session-details {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.key-header {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.session-row {
		display: flex;
		gap: 0.5rem;
	}

	.label {
		font-weight: 500;
		color: var(--imag-text-color);
		min-width: 5rem;
	}

	.value {
		color: var(--imag-text-color);
	}

	.key-name {
		font-weight: 600;
		font-size: 1rem;
	}

	.status-badge,
	.current-badge {
		font-size: 0.75rem;
		padding: 0.2rem 0.5rem;
		border-radius: 1rem;
		font-weight: 500;
		border: 1px solid var(--imag-80);
		color: var(--imag-text-color);
		background-color: var(--imag-90);
	}

	.status-badge.active {
		color: var(--imag-primary);
	}

	.status-badge.revoked {
		color: var(--imag-alert-color);
	}

	.current-badge {
		border-color: var(--imag-primary);
		// 	color: var(--imag-text-color);
	}

	.key-meta {
		display: flex;
		flex-direction: column;
		font-size: 0.75rem;
		color: var(--imag-text-color);
		gap: 0.25rem;
	}

	.key-actions,
	.session-actions {
		display: flex;
		gap: 0.5rem;
	}

	.sessions-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.edit-session-modal-content {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		padding: 1rem;
		min-width: 400px;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 1rem;
	}
</style>
