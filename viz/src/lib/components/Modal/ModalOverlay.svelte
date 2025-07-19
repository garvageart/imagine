<script lang="ts">
	import { modal } from "$lib/states/index.svelte";
	import Modal from "./Modal.svelte";
	let { children } = $props();
	let modalOverlayEl: HTMLElement | undefined = $state();

	if (window.debug) {
		$effect(() => {
			if (modal.show) {
				console.log("modal is showing");
			} else {
				console.log("modal is not showing");
			}
		});
	}
</script>

<svelte:window
	on:keydown={(e) => {
		if (e.key === "Escape") {
			modal.show = false;
		}
	}}
/>

<svelte:document
	on:click={(e) => {
		if (e.target === modalOverlayEl) {
			modal.show = false;
		}
	}}
/>

{#if modal.show}
	<div id="viz-modal-overlay" bind:this={modalOverlayEl}>
		<Modal>
			{@render children()}
		</Modal>
	</div>
{/if}

<style lang="scss">
	#viz-modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background-color: rgba(0, 0, 0, 0.5);
		z-index: 9998;
		display: flex;
		justify-content: center;
		align-items: center;
	}
</style>
