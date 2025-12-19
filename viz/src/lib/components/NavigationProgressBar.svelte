<!-- Adjusted from https://github.com/immich-app/immich/blob/main/web/src/lib/components/shared-components/navigation-loading-bar.svelte -->
<script lang="ts">
	import { onMount } from "svelte";
	import { cubicOut } from "svelte/easing";
	import { Tween } from "svelte/motion";
	import ProgressBar from "./ProgressBar.svelte";

	let showing = $state(false);
	const delay = 100;

	// delay showing any progress for a little bit so very fast loads
	// do not cause flicker
	let progress = new Tween(0, {
		duration: 9000,
		delay: delay,
		easing: cubicOut
	});

	function animate() {
		showing = true;
		void progress.set(90);
	}

	onMount(() => {
		const timer = setTimeout(animate, delay);
		return () => clearTimeout(timer);
	});
</script>

{#if showing}
	<div class="app-progress">
		<ProgressBar variant="medium" bind:width={progress.target} />
	</div>
{/if}

<style lang="scss">
	.app-progress {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		z-index: 9999;
	}
</style>
