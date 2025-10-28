<script lang="ts">
	import { dev } from "$app/environment";
	import { page } from "$app/state";
	import { CLIENT_IS_PRODUCTION } from "$lib/constants";
	import { performSearch } from "$lib/search/execute";
	import { search, user } from "$lib/states/index.svelte";
	import { VizLocalStorage } from "$lib/utils/misc";
	import hotkeys from "hotkeys-js";
	import { onMount } from "svelte";
	import type { SvelteHTMLElements } from "svelte/elements";
	import MaterialIcon from "./MaterialIcon.svelte";
	import SearchInput from "./SearchInput.svelte";
	import UploadManager from "$lib/upload/manager.svelte";
	import { SUPPORTED_IMAGE_TYPES, SUPPORTED_RAW_FILES, type SupportedImageTypes } from "$lib/types/images";

	let { ...props }: SvelteHTMLElements["header"] = $props();

	// eventually this will move to a different page with a different way of enabling, this is just temporary
	const storeDebug = new VizLocalStorage<boolean>("debugMode");
	let devEnabled = $state(storeDebug.get() === null ? false : storeDebug.get()) as boolean;

	// i'd probably forget to remove this in prod setting so just check lmao
	if (dev || !CLIENT_IS_PRODUCTION) {
		$effect(() => {
			console.log("Debug mode is", devEnabled ? "enabled" : "disabled");
			storeDebug.set(devEnabled!);
			if (window.debug !== devEnabled) {
				window.debug = devEnabled;
			}
		});
	}

	onMount(() => {
		if (page.url.pathname !== "/search") {
			return;
		}

		// If URL has search params, perform search automatically
		const urlParams = new URLSearchParams(window.location.search);
		const q = urlParams.get("q");
		if (q) {
			search.value = q;
		}
	});

	let searchInputHasFocus = $state(search.element && document.activeElement === search.element);

	// If URL has search params, perform search automatically (only on client)
	if (page.url.pathname === "/search") {
		const urlParams = new URLSearchParams(window.location.search);
		const q = urlParams.get("q");
		if (q) {
			search.value = q;
		}
	}

	// Register hotkeys only in the browser/runtime. Keep hotkeys-js as requested.
	// Ctrl/Cmd+K toggles focus on the search input.
	hotkeys("ctrl+k, command+k", (e) => {
		e.preventDefault();
		if (!searchInputHasFocus) {
			search.element?.focus();
		} else {
			search.element?.blur();
		}
	});

	// Escape blurs the search input only when it is focused.
	hotkeys("esc, escape", (e) => {
		if (searchInputHasFocus) {
			e.preventDefault();
			search.element?.blur();
		}
	});

	function handleUpload(e: MouseEvent) {
		e.preventDefault();
		// allowed image types will come from the config but for now just hardcode
		const controller = new UploadManager([...SUPPORTED_RAW_FILES, ...SUPPORTED_IMAGE_TYPES] as SupportedImageTypes[]);
		controller.openFileHolder();
		controller.uploadImage();
	}
</script>

<header {...props} class="{props.class} no-select">
	<a id="viz-title" href="/">viz</a>
	<SearchInput
		placeholder="Search (Ctrl/Cmd + K)"
		bind:loading={search.loading}
		bind:value={search.value}
		bind:element={search.element}
		{performSearch}
		style="width: 30%;"
	/>
	<div class="header-button-container">
		<button id="upload-button" class="header-button" aria-label="Upload" onclick={handleUpload}>
			<MaterialIcon iconName="upload" iconStyle="sharp" />
			<span style="font-size: 0.9rem; font-weight: 500;"> Upload </span>
		</button>
		{#if dev || !CLIENT_IS_PRODUCTION}
			<button class="header-button" aria-label="Toggle Debug Mode" onclick={() => (devEnabled = !devEnabled)}>
				{#if devEnabled}
					<span class="debug-mode-text">ON</span>
				{:else}
					<span class="debug-mode-text">OFF</span>
				{/if}
				<MaterialIcon iconName="bug_report" />
			</button>
		{/if}
		<button
			id="account-button"
			class="header-button"
			aria-label="Account"
			title={user.data?.username ? `${user.data.username} (${user.data.email})` : "Account"}
		>
			<figure style="height: 100%; display: flex; align-items: center; justify-content: center;">
				<span style="font-weight: 700; font-size: 1em;">{user.data ? user.data.username[0] : "?"}</span>
			</figure>
		</button>
	</div>
</header>

<style lang="scss">
	header {
		background-color: var(--imag-bg-color);
		max-height: 3em;
		padding: 1em 1em;
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
		font-size: 1.2em;
		position: absolute;
		left: 1em;
	}

	#account-button {
		height: 2rem;
		width: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
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
		margin-right: 2em;
		color: var(--imag-text-color);
		font-size: 0.9rem;
		padding: 0.3em 0.6em;
	}

	.header-button-container {
		position: absolute;
		right: 1em;
		display: flex;
		align-items: center;
	}

	:global(.header-button) {
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: var(--imag-100);
		border-radius: 4em;
		padding: 0.2em 0.5em;
		font-size: 1rem;
		color: var(--imag-text-color);
		margin-right: 1em;
		cursor: pointer;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-primary);
			outline: none;
			background-color: var(--imag-80);
		}

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	.debug-mode-text {
		margin-right: 0.5em;
		font-family: var(--imag-code-font);
		font-weight: bold;
		font-size: 0.9em;
		color: var(--imag-text-color);
	}
</style>
