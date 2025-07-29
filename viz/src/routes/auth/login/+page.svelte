<script lang="ts">
	import LoginButtons from "$lib/components/LoginButtons.svelte";
	import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
</script>

<main>
	<DevWelcomeText />
	<LoginButtons />
</main>

<style lang="scss">
	main {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 100%;
	}
</style>
