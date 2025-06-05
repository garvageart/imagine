<script lang="ts">
	import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
	import Sidebar from "$lib/components/Sidebar.svelte";
	import LoginButtons from "$lib/components/LoginButtons.svelte";
	import hotkeys from "hotkeys-js";
	import { fullscreen } from "$lib/utils";

	hotkeys("f11", () => {
		if (document.fullscreenElement) {
			fullscreen.exit();
		} else {
			fullscreen.enter();
		}
	});
</script>

<main>
	<div id="viz-content">
		<DevWelcomeText></DevWelcomeText>
		<LoginButtons />
	</div>
	<Sidebar></Sidebar>
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
	}
</style>
