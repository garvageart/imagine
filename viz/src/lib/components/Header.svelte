<script lang="ts">
	import { dev } from "$app/environment";
	import { CLIENT_IS_PRODUCTION } from "$lib/constants";
	import { VizLocalStorage } from "$lib/utils";

	// eventually this will move to a different page with a different way of enabling, this is just temporary
	let devEnabled = $state(false);
	let devEnableClickCounter = $state(0);
	const devEnabledValue = 10;
	const storeDebug = new VizLocalStorage<boolean>("debugMode");

	// i'd probably forget to remove this in prod setting so just check lmao
	if (dev || !CLIENT_IS_PRODUCTION) {
		$effect(() => {
			if (devEnableClickCounter === devEnabledValue && !devEnabled) {
				devEnabled = true;
				devEnableClickCounter = 0;
				window.debug = true;
				storeDebug.set(true);

				console.log("Debug mode enabled through header click");
			} else if (devEnableClickCounter === devEnabledValue && devEnabled) {
				devEnabled = false;
				devEnableClickCounter = 0;
				window.debug = false;
				storeDebug.set(false);

				console.log("Debug mode disabled");
			}
		});
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<header role="button" tabindex="0" ontouchstart={() => devEnableClickCounter++} onclick={() => devEnableClickCounter++}>
	<a id="viz-title" href="/">viz</a>
</header>

<style>
	header {
		background-color: var(--imag-blue-100);
		height: 1.5em;
		padding: 0.3em 1em;
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--imag-blue-60);
		position: relative;
	}

	#viz-title {
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 1.2em;
	}
</style>
