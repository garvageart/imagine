<script lang="ts">
	import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
	import LoginButtons from "$lib/components/LoginButtons.svelte";
	import VizPanel from "$lib/components/panels/VizPanel.svelte";

	let vizContent: HTMLDivElement;
</script>

<svelte:head>
	<title>viz</title>
</svelte:head>

<main>
	<div bind:this={vizContent} id="viz-content">
		<!-- <DevWelcomeText></DevWelcomeText> -->
		<!-- <LoginButtons /> -->
		<VizPanel />
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

	#viz-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		width: 100%;
		height: 100%;
		overflow: auto;
	}
</style>
