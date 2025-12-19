<script lang="ts">
	import Header from "$lib/components/Header.svelte";
	import UploadPanel from "$lib/components/UploadPanel.svelte";
	import Notifications from "$lib/toast-notifcations/Notifications.svelte";
	import { sortState, upload } from "$lib/states/index.svelte";
	import "$lib/components/panels/viz-panel.scss";
	import { registerReady } from "$lib/stores/appReady";
	import { loadRuntimeConfig } from "$lib/runtime-config";
	import { onMount } from "svelte";

	let { children } = $props();

	$effect(() => {
		sortState.save();
	});

	// Fetch runtime config early and have the app wait for it before marking ready
	if (typeof window !== "undefined") {
		onMount(() => {
			const p = loadRuntimeConfig();
			registerReady(p);
			p.finally(() => {
				// warn if still using localhost fallback (client exports this helper)
				import("$lib/api")
					.then((m) => {
						try {
							m.warnIfLocalhostFallback();
						} catch (e) {}
					})
					.catch(() => {});
			});
		});
	}
</script>

<div class="viz-app-layout">
	<Header />
	<Notifications />
	{#if upload.files.length > 0}
		<UploadPanel />
	{/if}
	<main class="viz-content">
		{@render children()}
	</main>
</div>

<style>
	.viz-app-layout {
		display: flex;
		flex-direction: column;
		height: 100vh;
		width: 100vw;
		overflow: hidden;
		position: relative;
	}

	.viz-content {
		display: flex;
		flex: 1;
		overflow: hidden;
		width: 100%;
		position: relative;
	}
</style>
