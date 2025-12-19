<script lang="ts">
	import VizPanel from "$lib/components/panels/VizPanel.svelte";
	import { search } from "$lib/states/index.svelte";
	import SearchPage from "./search/+page.svelte";
	import { onMount } from "svelte";

	let vizContentContainer: HTMLDivElement | undefined = $state();
</script>

<svelte:head>
	<title>viz</title>
</svelte:head>

<a href="#main" class="skip-to-main" aria-label="Skip to main content">
	<p>Skip to main content</p>
</a>

<main id="main" class="no-select">
	{#if search.value !== "" && search.enableHomePageSearch}
		<SearchPage />
	{:else}
		<div class="viz-content-container" bind:this={vizContentContainer}>
			<VizPanel id="viz-content" />
		</div>
	{/if}
</main>

<style lang="scss">
	main {
		display: flex;
		flex-direction: row;
		align-items: center;
		width: 100%;
		height: 100%;
	}

	.skip-to-main {
		left: -999px;
		position: absolute;
		top: 0;
		width: 1px;
		height: 1px;
		overflow: hidden;
		z-index: -999;
		display: block;

		&:focus,
		&:focus-visible {
			color: var(--imag-text-color);
			background: var(--imag-primary);
			left: auto;
			height: auto;
			width: auto;
			padding: 0.2em 3em;
			outline: var(--imag-text-color) dashed 2px;
			outline-offset: -2px;
			z-index: 999;
		}
	}

	.viz-content-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		width: 100%;
		height: 100%;
		overflow: auto;
	}
</style>
