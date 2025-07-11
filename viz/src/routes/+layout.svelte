<script module lang="ts">
	declare global {
		interface Window {
			debug?: boolean;
			___vizConfig?: VizConfig;
			resetAndReloadLayout?: () => void;
			__APP_VERSION__: string;
		}
	}
</script>

<script>
	import { dev } from "$app/environment";

	import { CAN_DEBUG, DEFAULT_THEME } from "$lib/constants";
	import type { VizConfig } from "$lib/types/config.types";
	import Header from "$lib/components/Header.svelte";
	import "$lib/styles/scss/main.scss";
	import "@fontsource-variable/manrope";

	window.debug = CAN_DEBUG
	
	window.___vizConfig = {
		environment: dev ? "dev" : "prod",
		// @ts-ignore
		version: __APP_VERSION__ as string,
		debug: window.debug ?? false,
		theme: DEFAULT_THEME
	};

	let { children } = $props();
</script>

<Header />

{@render children()}
