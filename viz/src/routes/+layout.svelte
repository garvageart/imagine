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

	import { CAN_DEBUG } from "$lib/constants";
	import type { VizConfig } from "$lib/types/config.types";
	import "$lib/styles/scss/main.scss";
	import "@fontsource-variable/manrope";
	import "$lib/stores/appReady";

	window.debug = CAN_DEBUG;

	window.___vizConfig = {
		environment: dev ? "dev" : "prod",
		// @ts-ignore
		version: __APP_VERSION__ as string,
		debug: window.debug ?? false
	};

	let { children } = $props();
</script>

{@render children()}
