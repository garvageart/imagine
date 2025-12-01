<script module lang="ts">
	declare global {
		interface Window {
			___vizConfig?: VizConfig;
			resetAndReloadLayout?: () => void;
			__APP_VERSION__: string;
			__RUNTIME_CONFIG__: {
				[key: string]: string;
			};
		}
	}
</script>

<script>
	import { dev } from "$app/environment";
	import { debugState, themeState } from "$lib/states/index.svelte";
	import type { VizConfig } from "$lib/types/config.types";
	import "@fontsource-variable/manrope";
	import "$lib/styles/scss/main.scss";
	import "$lib/stores/appReady";

	window.___vizConfig = {
		environment: dev ? "dev" : "prod",
		// @ts-ignore
		version: __APP_VERSION__ as string
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
</script>

{@render children()}
