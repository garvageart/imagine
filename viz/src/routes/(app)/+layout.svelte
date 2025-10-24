<script lang="ts">
	import Header from "$lib/components/Header.svelte";
	import "$lib/components/panels/viz-panel.scss";
	import UploadPanel from "$lib/components/UploadPanel.svelte";
	import { upload } from "$lib/states/index.svelte";
	import Notifications from "$lib/toast-notifcations/Notifications.svelte";

	let { children } = $props();
</script>

<Header />
<Notifications />
{#if upload.files.length > 0}
	<UploadPanel />
{/if}
{@render children()}
