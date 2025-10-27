<script lang="ts">
	import type { MenuItem } from "./ContextMenu.svelte";

	interface Props {
		item: MenuItem;
		index?: number;
		active?: boolean;
		onselect?: (detail: { item: MenuItem; index: number; event: MouseEvent }) => void;
	}

	let { item, index = 0, active = false, onselect }: Props = $props();

	function onClick(e: MouseEvent) {
		if (item.disabled || item.separator) return;
		onselect?.({ item, index, event: e });
	}
</script>

<li role="none">
	<button
		role="menuitem"
		aria-disabled={item.disabled ? "true" : undefined}
		class:disabled={!!item.disabled}
		data-index={index}
		tabindex={active ? 0 : -1}
		onclick={onClick}
	>
		{#if item.icon}
			<span class="icon" aria-hidden="true">{item.icon}</span>
		{/if}
		<span class="label">{item.label}</span>
		{#if item.shortcut}
			<span class="shortcut" aria-hidden="true">{item.shortcut}</span>
		{/if}
	</button>
	{#if item.danger}
		<!-- Optional visual hook for danger styling via parent CSS -->
	{/if}
	{#if item.separator}
		<!-- Not rendered here; separators handled by parent list -->
	{/if}
</li>

<style>
	li {
		display: flex;
		list-style-type: none;
		width: 100%;
	}

	li button:hover {
		background-color: #3f3f3f;
	}

	li > button {
		display: grid;
		grid-template-columns: auto 1fr auto;
		gap: 0.5rem;
		align-items: center;
		font-size: 0.9rem;
		font-weight: 500;
		font-family: "Switzer-Variable", system-ui, sans-serif;
		padding: 0.5rem 0.6rem;
		text-align: left;
		width: 100%;
		border: 0px;
		color: var(--almost-white);
		background-color: var(--almost-black);
		cursor: pointer;
		border-radius: 6px;
	}

	li > button.disabled {
		color: #7a7a7a;
		cursor: default;
	}

	li > button.disabled:hover {
		background-color: var(--almost-black);
	}

	.icon {
		opacity: 0.85;
		width: 1rem;
		text-align: center;
	}

	.shortcut {
		opacity: 0.6;
		font-size: 0.8em;
	}
</style>
