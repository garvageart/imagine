<script lang="ts">
	import { goto } from "$app/navigation";
	import {
		system,
		user,
		toggleTheme,
		getTheme
	} from "$lib/states/index.svelte";
	import {
		setupSuperadmin,
		doUserOnboarding,
		getUserSettings,
		type UserSetting
	} from "$lib/api";
	import InputText from "$lib/components/dom/InputText.svelte";
	import InputSelect from "$lib/components/dom/InputSelect.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import { onMount } from "svelte";
	import { fade } from "svelte/transition";
	import { formatLabel } from "$lib/settings/utils";
	import ProgressBar from "$lib/components/ProgressBar.svelte";

	let isLoading = $state(false);
	let currentStep = $state(0);

	// --- Superadmin State ---
	let superadminForm = $state({
		username: "",
		email: "",
		password: "",
		confirmPassword: "",
		firstName: "",
		lastName: ""
	});

	// --- User Onboarding State ---
	let userForm = $state({
		firstName: "",
		lastName: ""
	});

	let userSettings = $state<UserSetting[]>([]);
	let userSettingsValues = $state<Record<string, string>>({});

	// Group settings by their 'group' field
	let settingsGroups = $derived.by(() => {
		const groups: Record<string, UserSetting[]> = {};
		for (const s of userSettings) {
			const g = s.group || "General";
			if (!groups[g]) {
				groups[g] = [];
			}

			groups[g].push(s);
		}
		return groups;
	});

	let groupNames = $derived(Object.keys(settingsGroups).sort());

	// --- Logic ---

	async function loadUserSettings() {
		if (system.data?.user_onboarding_required) {
			isLoading = true;
			try {
				const res = await getUserSettings();
				if (res.status === 200) {
					userSettings = res.data.filter((s) => s.is_user_editable);

					// Pre-fill values
					userSettings.forEach((s) => {
						// Special handling for theme setting if it exists to match current UI state
						if (s.name === "theme" || s.name === "default_theme") {
							// If the user hasn't explicitly set it yet (it's using default),
							// we might want to propose the one they are currently viewing.
							// However, s.value might already be set from DB default.
							userSettingsValues[s.name] = s.value || s.default_value;
						} else {
							userSettingsValues[s.name] = s.value || s.default_value;
						}
					});

					if (user.data) {
						userForm.firstName = user.data.first_name || "";
						userForm.lastName = user.data.last_name || "";
					}
				} else {
					toastState.addToast({
						message: "Failed to load settings.",
						type: "error"
					});
				}
			} catch (e) {
				console.error(e);
			} finally {
				isLoading = false;
			}
		}
	}

	function handleThemeToggle() {
		toggleTheme();
		// Sync with form if 'theme' setting exists
		const newTheme = getTheme(); // 'light' or 'dark' (or 'system' if we could query that state, but getTheme resolves it)

		// Look for a theme setting key
		for (const key of Object.keys(userSettingsValues)) {
			if (key === "theme" || key === "default_theme") {
				userSettingsValues[key] = newTheme;
			}
		}
	}

	onMount(() => {
		if (
			system.data?.user_onboarding_required &&
			!system.data?.needs_superadmin
		) {
			loadUserSettings();
		}
	});

	// --- Navigation Helpers ---
	function nextStep() {
		currentStep++;
	}

	function prevStep() {
		if (currentStep > 0) currentStep--;
	}

	// --- Actions ---

	async function handleSuperadminSubmit() {
		if (superadminForm.password !== superadminForm.confirmPassword) {
			toastState.addToast({
				message: "Passwords do not match.",
				type: "error"
			});
			return;
		}

		isLoading = true;
		try {
			const res = await setupSuperadmin({
				username: superadminForm.username,
				email: superadminForm.email,
				password: superadminForm.password,
				firstName: superadminForm.firstName,
				lastName: superadminForm.lastName
			});

			if (res.status === 201) {
				toastState.addToast({
					message: "Superadmin setup complete!",
					type: "success"
				});

				// Invalidate system state to force re-fetch of status flags
				system.fetched = false;
				system.data = null;

				window.location.href = "/";
			} else {
				toastState.addToast({
					message: res.data.error || "Setup failed.",
					type: "error"
				});
			}
		} catch (err) {
			console.error(err);
			toastState.addToast({
				message: "An unexpected error occurred.",
				type: "error"
			});
		} finally {
			isLoading = false;
		}
	}

	async function handleUserOnboardingSubmit() {
		isLoading = true;
		try {
			const res = await doUserOnboarding({
				first_name: userForm.firstName,
				last_name: userForm.lastName,
				settings: userSettings.map((setting) => ({
					name: setting.name,
					display_name: setting.display_name,
					value: userSettingsValues[setting.name] || setting.default_value,
					default_value: setting.default_value,
					value_type: setting.value_type,
					allowed_values: setting.allowed_values,
					is_user_editable: setting.is_user_editable,
					group: setting.group,
					description: setting.description
				}))
			});

			if (res.status === 200) {
				toastState.addToast({ message: "Welcome aboard!", type: "success" });

				// Invalidate system state to force re-fetch of status flags
				// This ensures the router knows onboarding is complete
				system.fetched = false;
				system.data = null;

				user.data = res.data;
				goto("/");
			} else {
				toastState.addToast({
					message: res.data.error || "Onboarding failed.",
					type: "error"
				});
			}
		} catch (err) {
			console.error(err);
			toastState.addToast({
				message: "An unexpected error occurred.",
				type: "error"
			});
		} finally {
			isLoading = false;
		}
	}

	let totalSteps = $derived(2 + groupNames.length);
	let progressBarWidth = $derived(((currentStep + 1) / totalSteps) * 100);
</script>

<div class="onboarding-container">
	<button
		class="theme-toggle"
		onclick={handleThemeToggle}
		title="Toggle Theme"
		aria-label="Toggle Theme"
	>
		<MaterialIcon
			weight={300}
			iconName={getTheme() === "dark" ? "dark_mode" : "light_mode"}
			style="font-size: 1.5rem;"
		/>
	</button>

	<div class="card">
		{#if system.loading}
			<div class="loading">Loading...</div>
		{:else if system.data?.needs_superadmin}
			<!-- SUPERADMIN FLOW -->
			<div class="step-container" in:fade={{ duration: 200 }}>
				<!-- Step 0: Welcome -->
				{#if currentStep === 0}
					<div class="step-content">
						<h1>Welcome to Imagine</h1>
						<div class="info-box">
							<p>
								<strong>You are the first user!</strong>
							</p>
							<p>
								As the first user, you will be granted <strong
									>Superadmin</strong
								>
								privileges. This gives you full control over the system, including
								managing other users, system settings, and more.
							</p>
							<p>Let's get your account set up.</p>
						</div>
					</div>
					<div class="actions">
						<Button onclick={nextStep}>Get Started</Button>
					</div>

					<!-- Step 1: Account Info -->
				{:else if currentStep === 1}
					<div class="step-content">
						<h2>Account Details</h2>
						<p class="subtitle">Set your login credentials.</p>
						<form onsubmit={(e) => e.preventDefault()}>
							<InputText
								label="Username"
								bind:value={superadminForm.username}
								required
							/>
							<InputText
								label="Email"
								type="email"
								bind:value={superadminForm.email}
								required
							/>
							<InputText
								label="Password"
								type="password"
								bind:value={superadminForm.password}
								required
							/>
							<InputText
								label="Confirm Password"
								type="password"
								bind:value={superadminForm.confirmPassword}
								required
							/>
						</form>
					</div>
					<div class="actions">
						<Button onclick={prevStep}>Back</Button>
						<Button
							disabled={!superadminForm.username ||
								!superadminForm.email ||
								!superadminForm.password}
							onclick={nextStep}>Next</Button
						>
					</div>

					<!-- Step 2: Profile Info -->
				{:else if currentStep === 2}
					<div class="step-content">
						<h2>Your Profile</h2>
						<p class="subtitle">Tell us a bit about yourself.</p>
						<form onsubmit={(e) => e.preventDefault()}>
							<InputText
								label="First Name"
								bind:value={superadminForm.firstName}
							/>
							<InputText
								label="Last Name"
								bind:value={superadminForm.lastName}
							/>
						</form>
					</div>
					<div class="actions">
						<Button onclick={prevStep}>Back</Button>
						<Button onclick={handleSuperadminSubmit} disabled={isLoading}>
							{isLoading ? "Creating Account..." : "Finish Setup"}
						</Button>
					</div>
				{/if}
			</div>
		{:else if system.data?.user_onboarding_required}
			<!-- USER ONBOARDING FLOW -->
			<!-- 0:Intro, 1:Profile, 2..N:Settings -->

			<div class="step-container" in:fade={{ duration: 200 }}>
				<ProgressBar bind:width={progressBarWidth} />

				{#if currentStep === 0}
					<div class="step-content center-text">
						<h1>Welcome, {user.data?.username || "Traveler"}!</h1>
						<p>We're glad you're here.</p>
						<p>
							Before you dive in, we need to gather a few details to personalize
							your experience.
						</p>
					</div>
					<div class="actions centered">
						<Button onclick={nextStep}>Let's Go</Button>
					</div>
				{:else if currentStep === 1}
					<div class="step-content">
						<h2>Personal Details</h2>
						<p class="subtitle">How should we address you?</p>
						<form onsubmit={(e) => e.preventDefault()}>
							<InputText
								label="First Name"
								bind:value={userForm.firstName}
								required
							/>
							<InputText
								label="Last Name"
								bind:value={userForm.lastName}
								required
							/>
						</form>
					</div>
					<div class="actions">
						<Button onclick={prevStep}>Back</Button>
						<Button
							disabled={!userForm.firstName || !userForm.lastName}
							onclick={nextStep}>Next</Button
						>
					</div>

					<!-- Steps 2...N: Settings Groups -->
				{:else if currentStep >= 2 && currentStep < 2 + groupNames.length}
					{@const groupIndex = currentStep - 2}
					{@const groupName = groupNames[groupIndex]}
					{@const settings = settingsGroups[groupName]}

					<div class="step-content">
						<h2>{groupName.replace(/_/g, " ")} Settings</h2>
						<p class="subtitle">Customize your experience.</p>

						<div class="settings-list">
							{#each settings as setting}
								<div class="setting-item">
									{#if setting.value_type === "enum" && setting.allowed_values}
										<InputSelect
											label={setting.name.trim()
												? setting.name
												: formatLabel(setting.name)}
											description={setting.description}
											bind:value={userSettingsValues[setting.name]}
										>
											{#each setting.allowed_values as option}
												<option value={option}>{option}</option>
											{/each}
										</InputSelect>
									{:else if setting.value_type === "boolean"}
										<InputSelect
											label={setting.name.trim()
												? setting.name
												: formatLabel(setting.name)}
											description={setting.description}
											bind:value={userSettingsValues[setting.name]}
										>
											<option value="true">Yes</option>
											<option value="false">No</option>
										</InputSelect>
									{:else}
										<InputText
											label={setting.name.trim()
												? setting.name
												: formatLabel(setting.name)}
											description={setting.description}
											bind:value={userSettingsValues[setting.name]}
										/>
									{/if}
								</div>
							{/each}
						</div>
					</div>
					<div class="actions">
						<Button onclick={prevStep}>Back</Button>
						{#if currentStep === totalSteps - 1}
							<Button onclick={handleUserOnboardingSubmit} disabled={isLoading}>
								{isLoading ? "Finishing..." : "Complete Setup"}
							</Button>
						{:else}
							<Button onclick={nextStep}>Next</Button>
						{/if}
					</div>
				{/if}
			</div>
		{:else}
			<div class="loading">Redirecting...</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.onboarding-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 100vh;
		background-color: var(--imag-bg-color);
		padding: 1rem;
		position: relative;
	}

	.theme-toggle {
		position: absolute;
		top: 1.5rem;
		right: 1.5rem;
		background: transparent;
		border: none;
		color: var(--imag-text-color);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background-color 0.2s;

		&:hover {
			background-color: var(--imag-90);
		}
	}

	.card {
		background: var(--imag-100);
		border-radius: 1rem;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
		width: 100%;
		max-width: 40%;
		border: 1px solid var(--imag-60);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		min-height: 50%;
		position: relative;
	}

	.step-container {
		padding: 2.5rem;
		display: flex;
		flex-direction: column;
		flex: 1;
		height: 100%;
	}

	.step-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;

		form {
			display: flex;
			flex-direction: column;
			gap: 1rem;
		}

		&.center-text {
			text-align: center;
			justify-content: center;
		}
	}

	h1 {
		font-size: 2rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
		color: var(--imag-text-color);
	}

	h2 {
		font-size: 1.5rem;
		color: var(--imag-text-color);
		margin-bottom: 0.25rem;
		text-transform: capitalize;
	}

	.subtitle {
		color: var(--imag-40);
		margin-bottom: 1rem;
	}

	.info-box {
		background: var(--imag-90);
		border: 1px solid var(--imag-primary);
		padding: 1rem;
		border-radius: 0.5rem;
		color: var(--imag-text-color);
		font-size: 0.95rem;
		line-height: 1.5;

		p {
			margin-bottom: 0.5rem;
			&:last-child {
				margin-bottom: 0;
			}
		}
	}

	.actions {
		display: flex;
		justify-content: space-between;
		margin-top: 2rem;
		padding-top: 1rem;

		&.centered {
			justify-content: center;
			border-top: none;
		}
	}

	.settings-list {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		max-height: 350px;
		overflow-y: auto;
		padding-right: 0.5rem;
	}

	.setting-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.loading {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		color: var(--imag-40);
	}
</style>
