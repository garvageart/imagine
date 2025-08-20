<script lang="ts">
	import { lightbox } from "$lib/states/index.svelte";

	let lightboxEl: HTMLElement | undefined = $state();
	let {
		children,
		onclick,
		lightboxElement = $bindable()
	}: {
		children: () => any;
		onclick?: (
			e: MouseEvent & {
				currentTarget: EventTarget & Document;
			}
		) => void;
		lightboxElement?: HTMLElement | undefined;
	} = $props();

	$effect(() => {
		lightboxElement = lightboxEl;
	});

	if (window.debug) {
		$effect(() => {
			if (lightbox.show) {
				console.log("lightbox is showing");
			} else {
				console.log("lightbox is not showing");
			}
		});
	}
</script>

<svelte:window
	on:keydown={(e) => {
		if (e.key === "Escape") {
			lightbox.show = false;
		}
	}}
/>

<svelte:document
	on:click={(e) => {
		if (e.target === lightboxEl) {
			lightbox.show = false;
			onclick?.(e);
		}
	}}
/>

{#if lightbox.show}
	<div id="viz-lightbox-overlay" bind:this={lightboxEl}>
		{@render children()}
	</div>
{/if}

<style lang="scss">
	#viz-lightbox-overlay {
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
