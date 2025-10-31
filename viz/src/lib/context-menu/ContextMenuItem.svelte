<script lang="ts">
	import type { MenuItem } from "./ContextMenu.svelte";
	import MaterialIcon from "../components/MaterialIcon.svelte";

	interface Props {
		item: MenuItem;
		index?: number;
		active?: boolean;
		onselect?: (detail: { item: MenuItem; index: number; event: MouseEvent }) => void;
	}

	let { item, index = 0, active = false, onselect }: Props = $props();

	function onClick(e: MouseEvent) {
		if (item.disabled || item.separator) {
			return;
		}

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
			<MaterialIcon class="icon" iconName={item.icon} weight={300} />
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
		background-color: var(--imag-60);
	}

	li > button {
		display: grid;
		grid-template-columns: auto 1fr auto;
		gap: 0.5rem;
		align-items: center;
		font-size: 0.9rem;
		font-weight: 500;
		padding: 0.1rem 0.6rem;
		text-align: left;
		width: 100%;
		border: 0px;
		color: var(--imag-text-color);
		background-color: var(--imag-100);
		cursor: pointer;
	}

	li > button.disabled {
		color: var(--imag-70);
		cursor: default;
	}

	li > button.disabled:hover {
		background-color: var(--imag-20);
	}

	.shortcut {
		opacity: 0.6;
		font-size: 0.8em;
	}
</style>
