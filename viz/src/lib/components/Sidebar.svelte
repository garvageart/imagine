<script lang="ts">
	import { slide } from "svelte/transition";
	import MaterialIcon from "./MaterialIcon.svelte";

	interface Props {
		open?: boolean;
		children?: import("svelte").Snippet;
	}

	let { open = $bindable(true), children }: Props = $props();

	let sidebarEl: HTMLElement;
</script>

<nav bind:this={sidebarEl} id="viz-sidebar" class="border {open ? '' : 'closed'}">
	{#if open}
		<button class="close-sidebar-button" title="Close Settings Sidebar" onclick={() => (open = !open)}>
			<MaterialIcon iconName="close" />
		</button>
		<div class="sidebar-content">
			{@render children?.()}
		</div>
	{:else}
		<button
			id="open-sidebar-button"
			title="Open Settings Sidebar"
			onclick={() => (open = true)}
			out:slide={{ axis: "x", duration: 300 }}
		>
			<MaterialIcon iconName="arrow_right" />
		</button>
	{/if}
</nav>

<style lang="scss">
	#viz-sidebar {
		background-color: var(--imag-100);
		border-right: 1px solid var(--imag-60);
		height: 100%;
		min-width: 20%;
		display: flex;
		position: relative;
		transition: min-width 0.3s ease;

		&.closed {
			min-width: 3%;
			justify-content: center;
		}
	}

	.sidebar-content {
		width: 100%;
		height: 100%;
		overflow-y: auto;
	}

	#open-sidebar-button {
		position: absolute;
		height: 2em;
		min-width: 2em;
		width: 100%;
		top: 2em;
		background-color: var(--imag-80);
		border-top-right-radius: 3em;
		border-bottom-right-radius: 3em;
	}

	.close-sidebar-button {
		position: absolute;
		left: 0;
		margin: 0.5em;
		z-index: 10;
	}
</style>
