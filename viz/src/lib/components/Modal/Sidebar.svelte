<script lang="ts">
	import SidebarCloseIcon from "../SidebarCloseIcon.svelte";
	import { sidebar } from "$lib/states/index.svelte";

	let sidebarEl: HTMLElement;
	let sidebarOpen = $derived(sidebar.open);
</script>

<div bind:this={sidebarEl} id="viz-sidebar" class="border {sidebarOpen ? '' : 'closed'}">
	<span id="close-sidebar">
		<SidebarCloseIcon
			onclick={() => {
				sidebar.open = !sidebar.open;
			}}
		/>
	</span>
</div>

<style lang="scss">
	#viz-sidebar {
		background-color: var(--imag-100);
		border-right: 1px solid var(--imag-20);
		height: 100%;
		min-width: 15%;
		display: flex;
		position: relative;
		transition: min-width 0.5s;
		transition-timing-function: linear(
			0 0%,
			0 1.8%,
			0.01 3.6%,
			0.03 6.35%,
			0.07 9.1%,
			0.13 11.4%,
			0.19 13.4%,
			0.27 15%,
			0.34 16.1%,
			0.54 18.35%,
			0.66 20.6%,
			0.72 22.4%,
			0.77 24.6%,
			0.81 27.3%,
			0.85 30.4%,
			0.88 35.1%,
			0.92 40.6%,
			0.94 47.2%,
			0.96 55%,
			0.98 64%,
			0.99 74.4%,
			1 86.4%,
			1 100%
		);

		&.closed {
			min-width: 3%;
			justify-content: center;
		}
	}

	#close-sidebar {
		position: absolute;
		right: 0;
		margin: 0.2em;
	}
</style>
