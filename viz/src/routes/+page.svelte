<script lang="ts">
	import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
	import LoginButtons from "$lib/components/LoginButtons.svelte";
	import VizPanel from "$lib/components/panels/VizPanel.svelte";
	import { login } from "$lib/states/index.svelte";
	
	let vizContentContainer: HTMLDivElement | undefined = $state();
	const loginState = login.state;
</script>

<svelte:head>
	<title>viz</title>
</svelte:head>

<main>
	<div class="viz-content-container" bind:this={vizContentContainer}>
		{#if !loginState}
			<DevWelcomeText />
			<LoginButtons />
		{:else}
			<VizPanel id="viz-content" />
		{/if}
	</div>
</main>

<style>
	main {
		display: flex;
		flex-direction: row;
		align-items: center;
		height: 100%;
		width: 100%;
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
