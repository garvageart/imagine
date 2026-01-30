<script module lang="ts">
	declare global {
		interface Window {
			___viewfinderConfig?: ViewfinderConfig;
			resetAndReloadLayout?: () => void;
			__APP_VERSION__: string;
			__RUNTIME_CONFIG__: {
				[key: string]: string;
			};
			__VIZ_CONFIG__?: VizConfig;
		}
	}
</script>

<script>
	import { dev } from "$app/environment";
	import { afterNavigate, beforeNavigate } from "$app/navigation";

	import NavigationProgressBar from "$lib/components/NavigationProgressBar.svelte";
	import { historyState } from "$lib/states/history.svelte";
	import { debugState, themeState } from "$lib/states/index.svelte";
	import { loadingState } from "$lib/states/loading.svelte";
	import "$lib/stores/appReady";
	import "$lib/styles/scss/main.scss";
	import type { VizConfig, ViewfinderConfig } from "$lib/types/config.types";
	import { toggleFullscreen } from "$lib/utils/misc";
	import "@fontsource-variable/manrope/index.css";
	// import "@fontsource-variable/public-sans/index.css";
	// import "@fontsource/clear-sans/index.css";
	import "@fontsource-variable/roboto-mono/index.css";
	import hotkeys from "hotkeys-js";

	historyState.init();

	window.___viewfinderConfig = {
		environment: dev ? "dev" : "prod",
		// @ts-ignore
		version: window.__APP_VERSION__
	};

	let { children } = $props();

	$effect(() => {
		const themeScript = document.getElementById("theme-ready-script");
		if (themeScript) {
			themeScript.remove();
		}
	});

	$effect(() => {
		debugState.storage.set(debugState.value);
	});

	$effect(() => {
		themeState.ls.set(themeState.value);
		if (typeof document !== "undefined") {
			document.documentElement.setAttribute("data-theme", themeState.resolved);
		}
	});

	hotkeys("shift+f", (e) => {
		e.preventDefault();
		toggleFullscreen();
	});

	let showNavProgress = $state(false);

	beforeNavigate(() => {
		showNavProgress = true;
		loadingState.startNavigation();
	});

	afterNavigate(() => {
		showNavProgress = false;
		loadingState.endNavigation();
	});
</script>

{#if showNavProgress}
	<NavigationProgressBar />
{/if}
{@render children()}
