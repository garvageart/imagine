<script lang="ts">
	import type { UserSetting } from "$lib/api/client.gen";
	import { updateUserSetting } from "$lib/api/client.gen";
	import SliderToggleInput from "./inputs/SliderToggleInput.svelte";
	import SelectInput from "./inputs/SelectInput.svelte";
	import TextInput from "./inputs/TextInput.svelte";
	import JsonInput from "./inputs/JsonInput.svelte";
	import { fade } from "svelte/transition";

	interface Props {
		settings?: UserSetting[];
		title: string;
		description?: string;
	}

	let { settings = $bindable([]), title, description = "" }: Props = $props();

	// Track modified settings: map of setting name -> new value
	let dirtySettings: Record<string, string> = $state({});
	let saving = $state(false);
	let saveStatus: "idle" | "success" | "error" = $state("idle");
	let errorMessage = $state("");

	let hasChanges = $derived(Object.keys(dirtySettings).length > 0);

	function handleSettingChange(setting: UserSetting, newValue: any) {
		// If the new value matches the original value, remove from dirty list
		if (String(newValue) === setting.value) {
			const newDirty = { ...dirtySettings };
			delete newDirty[setting.name];
			dirtySettings = newDirty;
		} else {
			dirtySettings = {
				...dirtySettings,
				[setting.name]: String(newValue)
			};
		}
		saveStatus = "idle";
	}

	async function saveChanges() {
		if (!hasChanges) {
			return;
		}

		saving = true;
		saveStatus = "idle";
		errorMessage = "";

		try {
			const updates = Object.entries(dirtySettings).map(([name, value]) => updateUserSetting(name, { value }));

			await Promise.all(updates);

			// Update local state to reflect saved changes
			settings = settings.map((s) => ({
				...s,
				value: dirtySettings[s.name] ?? s.value
			}));

			dirtySettings = {};
			saveStatus = "success";

			setTimeout(() => {
				saveStatus = "idle";
			}, 3000);
		} catch (e) {
			console.error("Failed to save settings", e);
			saveStatus = "error";
			errorMessage = "Failed to save changes. Please try again.";
		} finally {
			saving = false;
		}
	}

	function formatLabel(name: string): string {
		return name
			.replace(/^[a-z]+_/, "")
			.replace(/_/g, " ")
			.replace(/\b\w/g, (l) => l.toUpperCase());
	}

	function getToggleValue(settingName: string, originalValue: string): "on" | "off" {
		const val = dirtySettings[settingName] ?? originalValue;
		return val === "true" ? "on" : "off";
	}
</script>

<div class="settings-group">
	<header>
		<div>
			<h2>{title}</h2>
			{#if description}
				<p class="group-description">{description}</p>
			{/if}
		</div>

		{#if hasChanges || saveStatus === "success"}
			<div class="actions" transition:fade>
				{#if saveStatus === "success"}
					<span class="success-msg">Saved!</span>
				{/if}
				{#if hasChanges}
					<button class="btn-save" disabled={saving} onclick={saveChanges}>
						{saving ? "Saving..." : "Save Changes"}
					</button>
				{/if}
			</div>
		{/if}
	</header>

	{#if errorMessage}
		<div class="error-banner" transition:fade>
			{errorMessage}
		</div>
	{/if}

	<div class="settings-list">
		{#each settings as setting (setting.name)}
			<div class="setting-item">
				{#if setting.value_type === "boolean"}
					{@const currentVal = getToggleValue(setting.name, setting.value)}
					<SliderToggleInput
						label={formatLabel(setting.name)}
						description={setting.description}
						value={currentVal}
						disabled={!setting.is_user_editable || saving}
						onchange={(val) => {
							// val comes from SliderToggleInput which binds to SliderToggle which uses "on"/"off"
							// We need to store "true"/"false" in dirtySettings for the API
							const newVal = val === "on" ? "true" : "false";
							handleSettingChange(setting, newVal);
						}}
					/>
				{:else if setting.value_type === "enum"}
					<SelectInput
						label={formatLabel(setting.name)}
						description={setting.description}
						value={dirtySettings[setting.name] ?? setting.value}
						options={setting.allowed_values || []}
						disabled={!setting.is_user_editable || saving}
						onchange={(val) => handleSettingChange(setting, val)}
					/>
				{:else if setting.value_type === "integer"}
					<TextInput
						type="number"
						label={formatLabel(setting.name)}
						description={setting.description}
						value={dirtySettings[setting.name] ?? setting.value}
						disabled={!setting.is_user_editable || saving}
						onchange={(val) => handleSettingChange(setting, val)}
					/>
				{:else if setting.value_type === "json"}
					<JsonInput
						label={formatLabel(setting.name)}
						description={setting.description}
						value={dirtySettings[setting.name] ?? setting.value}
						disabled={!setting.is_user_editable || saving}
						onchange={(val) => handleSettingChange(setting, val)}
					/>
				{:else}
					<TextInput
						type="text"
						label={formatLabel(setting.name)}
						description={setting.description}
						value={dirtySettings[setting.name] ?? setting.value}
						disabled={!setting.is_user_editable || saving}
						onchange={(val) => handleSettingChange(setting, val)}
					/>
				{/if}
			</div>
		{/each}

		{#if settings.length === 0}
			<div class="empty-state">No settings available in this group.</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.settings-group {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		max-width: 800px;
	}

	header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--imag-80);

		h2 {
			font-size: 1.5rem;
			font-weight: 600;
			color: var(--imag-text-color);
			margin: 0 0 0.5rem 0;
		}

		.group-description {
			color: var(--imag-40);
			margin: 0;
		}
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.success-msg {
		color: var(--imag-success-color, #10b981);
		font-weight: 500;
	}

	.btn-save {
		background-color: var(--imag-primary);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-weight: 500;
		cursor: pointer;
		transition: opacity 0.2s;

		&:hover:not(:disabled) {
			opacity: 0.9;
		}

		&:disabled {
			opacity: 0.7;
			cursor: not-allowed;
		}
	}

	.error-banner {
		background-color: rgba(239, 68, 68, 0.1);
		color: var(--imag-error-color, #ef4444);
		padding: 0.75rem;
		border-radius: 0.375rem;
		border: 1px solid var(--imag-error-color, #ef4444);
	}

	.settings-list {
		display: flex;
		flex-direction: column;
	}

	.empty-state {
		padding: 2rem;
		text-align: center;
		color: var(--imag-40);
		background-color: var(--imag-100);
		border-radius: 0.5rem;
	}
</style>
