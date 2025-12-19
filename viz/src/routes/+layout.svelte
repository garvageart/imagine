<script module lang="ts">
	declare global {
		interface Window {
			___vizConfig?: VizConfig;
			resetAndReloadLayout?: () => void;
			__APP_VERSION__: string;
			__RUNTIME_CONFIG__: {
				[key: string]: string;
			};
			__IMAGINE_CONFIG__?: ImagineConfig;
		}
	}
</script>

<script>
	import { dev } from "$app/environment";
	import { debugState, themeState } from "$lib/states/index.svelte";
	import { historyState } from "$lib/states/history.svelte";
	import { toggleFullscreen } from "$lib/utils/misc";
	import type { VizConfig, ImagineConfig } from "$lib/types/config.types";
	import hotkeys from "hotkeys-js";
	import "@fontsource-variable/manrope/index.css";
	import "@fontsource-variable/roboto-mono/index.css";
	import "$lib/styles/scss/main.scss";
	import "$lib/stores/appReady";
	import { beforeNavigate, afterNavigate } from "$app/navigation";
	import NavigationProgressBar from "$lib/components/NavigationProgressBar.svelte";

	historyState.init();

	window.___vizConfig = {
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
	});

	afterNavigate(() => {
		showNavProgress = false;
	});
</script>

{#if showNavProgress}
	<NavigationProgressBar />
{/if}
{@render children()}
