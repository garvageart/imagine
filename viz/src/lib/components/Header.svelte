<script lang="ts">
	import { dev } from "$app/environment";
	import { page } from "$app/state";
	import { CLIENT_IS_PRODUCTION } from "$lib/constants";
	import { performSearch } from "$lib/search/execute";
	import { debugState, getTheme, search, toggleTheme, user } from "$lib/states/index.svelte";
	import { SUPPORTED_IMAGE_TYPES, SUPPORTED_RAW_FILES, type SupportedImageTypes } from "$lib/types/images";
	import UploadManager from "$lib/upload/manager.svelte";
	import hotkeys from "hotkeys-js";
	import type { SvelteHTMLElements } from "svelte/elements";
	import OpenAccountPanel from "./AccountPanel.svelte";
	import AppMenu from "./AppMenu.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import SearchInput from "./SearchInput.svelte";

	let { ...props }: SvelteHTMLElements["header"] = $props();

	let searchElement = $state<HTMLInputElement | undefined>();
	let searchInputHasFocus = $derived(searchElement && document.activeElement === searchElement);

	$effect(() => {
		if (page.url.pathname === "/search") {
			const q = page.url.searchParams.get("q");
			// Only update search.value if 'q' is present and different from current search.value
			// This prevents unnecessary updates and potential infinite loops if search.value also affects the URL
			if (q && q !== search.value) {
				search.value = q;
			}
		}
	});

	// Ctrl/Cmd+I toggles dev mode
	hotkeys("ctrl+i, command+i", (e) => {
		e.preventDefault();
		debugState.toggle();
	});

	// Ctrl/Cmd+K toggles focus on the search input.
	hotkeys("ctrl+k, command+k", (e) => {
		e.preventDefault();
		if (!searchInputHasFocus) {
			searchElement?.focus();
		} else {
			searchElement?.blur();
		}
	});

	function handleUpload(e: MouseEvent) {
		e.preventDefault();
		// allowed image types will come from the config but for now just hardcode
		const manager = new UploadManager([...SUPPORTED_RAW_FILES, ...SUPPORTED_IMAGE_TYPES] as SupportedImageTypes[]);

		// Use new API: open picker and upload (fire and forget - panel will show progress)
		manager.openPickerAndUpload();
	}

	let openAccPanel = $state(false);
	let openAppMenu = $state(false);
	let appMenuButton: HTMLButtonElement | undefined = $state();
</script>

<svelte:window
	onkeydown={(e) => {
		if (e.key !== "Escape") {
			return;
		}

		if (searchInputHasFocus) {
			console.log("Escape key pressed, blurring search input");
			searchElement?.blur();
			return;
		}

		if (openAccPanel) {
			openAccPanel = false;
		}

		if (openAppMenu) {
			openAppMenu = false;
		}
	}}
	onclick={(e) => {
		if (openAccPanel && !(e.target as HTMLElement).closest("#account-container")) {
			openAccPanel = false;
		}
	}}
/>

<header {...props} class="{props.class} no-select">
	<div id="app-menu-container">
		<button
			bind:this={appMenuButton}
			id="viz-title"
			onclick={() => (openAppMenu = !openAppMenu)}
			aria-label="App Menu"
			title="App Menu"
		>
			viz
			<MaterialIcon iconName="arrow_drop_down" weight={300} style="font-size: 1.2em; margin-left: 0.15em;" />
		</button>
		<AppMenu bind:isOpen={openAppMenu} bind:anchor={appMenuButton} />
	</div>
	<SearchInput
		inputId="header-search"
		placeholder="Search (Ctrl/Cmd + K)"
		bind:searchInputHasFocus
		bind:loading={search.loading}
		bind:value={search.value}
		bind:element={searchElement}
		{performSearch}
		style="width: 30%; border-color: var(--imag-80); height: 1.5em; font-size: 0.9em;"
	/>
	<div class="header-button-container">
		<button
			id="theme-toggle"
			class="header-button theme-toggle"
			title="Toggle theme"
			aria-label="Toggle Theme"
			onclick={() => toggleTheme()}
		>
			<MaterialIcon weight={300} iconName={getTheme() === "dark" ? "dark_mode" : "light_mode"} />
		</button>
		<button id="upload-button" class="header-button" aria-label="Upload" onclick={handleUpload}>
			<MaterialIcon iconName="upload" iconStyle="sharp" />
			<span style="font-size: 0.75rem; font-weight: 500;"> Upload </span>
		</button>
		{#if dev || !CLIENT_IS_PRODUCTION}
			<button
				id="debug-button"
				class="header-button"
				aria-label="Toggle Debug Mode"
				onclick={() => debugState.toggle()}
				title="Toggle Debug Mode"
			>
				<span class="debug-mode-text">{debugState.value ? "ON" : "OFF"}</span>
				<MaterialIcon iconName="bug_report" />
			</button>
		{/if}
		<div id="account-container">
			<button
				id="account-button"
				class="header-button"
				aria-label="Account"
				onclick={() => (openAccPanel = !openAccPanel)}
				title={user.data?.username ? `${user.data.username} (${user.data.email})` : "Account"}
			>
				<figure style="height: 100%; display: flex; align-items: center; justify-content: center;">
					<span style="font-weight: 700; font-size: 0.8em;">{user.data ? user.data.username[0] : "?"}</span>
				</figure>
			</button>
			{#if openAccPanel}
				<OpenAccountPanel bind:openAccPanel />
			{/if}
		</div>
	</div>
</header>

<style lang="scss">
	header {
		background-color: var(--imag-bg-color);
		max-height: 2em;
		padding: 0.3em 0.8em;
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--imag-60);
		position: relative;
		justify-content: center;
		flex-direction: row;
	}

	#viz-title {
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 1em;
		display: flex;
		align-items: center;
		gap: 0.1em;
		background: transparent;
		border: none;
		color: var(--imag-text-color);
		cursor: pointer;
		padding: 0.1em 0.5em;
		border-radius: 0.5rem;
		transition: background-color 0.1s ease;

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	#app-menu-container {
		position: absolute;
		left: 0.5rem;
		border-radius: 0.25rem;
		z-index: 300;
		display: flex;
		flex-direction: column;
	}

	#account-container {
		position: relative;
	}

	#account-button {
		height: 1.4rem;
		width: 1.4rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10em;
		outline: 1px solid var(--imag-60);
		background-color: var(--imag-80);
	}

	figure {
		display: block;
		margin-block-start: 0em;
		margin-block-end: 0em;
		margin-inline-start: 0px;
		margin-inline-end: 0px;
		unicode-bidi: isolate;
	}

	#upload-button {
		margin-right: 1.5em;
		color: var(--imag-text-color);
		font-size: 0.8rem;
		padding: 0.25em 0.5em;
	}

	.header-button-container {
		position: absolute;
		right: 0.8em;
		display: flex;
		align-items: center;
	}

	#theme-toggle {
		color: var(--imag-text-color);
		margin: auto 1.5rem;
	}

	:global(.header-button) {
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10em;
		padding: 0.15em 0.4em;
		font-size: 0.85rem;
		color: var(--imag-text-color);
		margin-right: 0.6em;
		cursor: pointer;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-primary);
			outline: none;
			background-color: var(--imag-90);
			border-radius: 4em;
		}

		&:hover {
			background-color: var(--imag-100);
			// outline: 1px solid var(--imag-60);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	.debug-mode-text {
		margin-right: 0.4em;
		font-family: var(--imag-code-font);
		font-weight: bold;
		font-size: 0.8em;
		color: var(--imag-text-color);
	}
</style>
