<script lang="ts">
	import Header from "$lib/components/Header.svelte";
	import "$lib/components/panels/viz-panel.scss";
	import UploadPanel from "$lib/components/UploadPanel.svelte";
	import { upload } from "$lib/states/index.svelte";

	let { children } = $props();
</script>

<Header />
{#if upload.files.length > 0}
	<UploadPanel />
{/if}
{@render children()}
