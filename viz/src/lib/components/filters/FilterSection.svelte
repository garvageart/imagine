<script lang="ts">
	import type { Snippet } from "svelte";
	import { slide } from "svelte/transition";
	import { toggleSection } from "./ImageFilter.svelte";

	interface Props {
		heading: string;
		expanded: boolean;
		header: Snippet;
		content: Snippet;
		onToggle: (expanded: boolean) => void;
	}

	let {
		heading,
		expanded = $bindable(),
		header,
		content,
		onToggle
	}: Props = $props();
</script>

<div class="filter-section">
	<button class="section-header" onclick={() => onToggle(expanded)}>
		<span>{heading}</span>
		{@render header()}
	</button>
	{#if expanded}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			{@render content()}
		</div>
	{/if}
</div>

<style lang="scss">
	.filter-section {
		border-bottom: 1px solid var(--imag-80);

		&:last-child {
			border-bottom: none;
		}
	}

	.section-header {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: transparent;
		border: none;
		padding: 0.2rem;
		cursor: pointer;
		color: var(--imag-text-color);
		font-weight: 600;
		font-size: 0.8rem;
		text-align: left;
		transition: color 0.2s;

		&:hover {
			color: var(--imag-20);
			background-color: var(--imag-90);
		}
	}

	.section-content {
		padding: 0.5rem 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
</style>
