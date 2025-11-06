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

	// submenu visibility on hover/focus
	let showSubmenu = $state(false);

	function onClick(e: MouseEvent) {
		if (item.disabled || item.separator) {
			return;
		}

		// If the item has children, clicking the parent shouldn't immediately activate
		if (item.children && item.children.length > 0) {
			showSubmenu = true;
			return;
		}

		onselect?.({ item, index, event: e });
	}

	function onChildClick(child: MenuItem, childIndex: number, e: MouseEvent) {
		if (child.disabled || child.separator) return;
		child.action?.(e);
		onselect?.({ item: child, index: childIndex, event: e });
		showSubmenu = false;
	}
</script>

<li role="none" onmouseenter={() => (showSubmenu = true)} onmouseleave={() => (showSubmenu = false)}>
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
		{#if item.children}
			<span class="submenu-arrow" aria-hidden="true">▸</span>
		{/if}
	</button>
	{#if item.children && item.children.length > 0}
		{#if showSubmenu}
			<div class="submenu" role="menu">
				<ul>
					{#each item.children as child, ci}
						{#if child.separator}
							<li class="separator" role="separator" aria-hidden="true"></li>
						{:else}
							<li role="none">
								<button
									role="menuitem"
									aria-disabled={child.disabled ? "true" : undefined}
									class:disabled={!!child.disabled}
									tabindex={-1}
									onclick={(e) => onChildClick(child, ci, e)}
								>
									{#if child.icon}
										<MaterialIcon class="icon" iconName={child.icon} weight={300} />
									{/if}
									<span class="label">{child.label}</span>
									{#if child.shortcut}
										<span class="shortcut" aria-hidden="true">{child.shortcut}</span>
									{/if}
								</button>
							</li>
						{/if}
					{/each}
				</ul>
			</div>
		{/if}
	{/if}
</li>

<style>
	li {
		display: flex;
		list-style-type: none;
		width: 100%;
		position: relative;
	}

	li > button {
		display: grid;
		grid-template-columns: auto 1fr auto;
		gap: 0.5rem;
		align-items: center;
		font-size: 0.9rem;
		font-weight: 500;
		padding: 0.25rem 0.75rem;
		text-align: left;
		width: 100%;
		border: 0px;
		color: var(--imag-text-color);
		background-color: var(--imag-100);
		cursor: pointer;
		transition: background-color 0.1s ease;
	}

	li > button:hover {
		background-color: var(--imag-90);
	}

	li > button.disabled {
		color: var(--imag-70);
		cursor: default;
		opacity: 0.5;
	}

	li > button.disabled:hover {
		background-color: var(--imag-100);
	}

	.shortcut {
		opacity: 0.6;
		font-size: 0.8em;
		margin-left: auto;
	}

	.submenu {
		position: absolute;
		/* overlap slightly with parent to avoid hover gap */
		left: calc(100% - 6px);
		top: 0.15rem;
		background: var(--imag-100);
		box-shadow:
			0 10px 30px rgba(0, 0, 0, 0.35),
			0 2px 8px rgba(0, 0, 0, 0.3);
		border-radius: 0.5rem;
		overflow: hidden;
		z-index: 995;
		box-sizing: border-box;
		min-width: 10rem;
	}

	.submenu ul {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.submenu ul li {
		width: 100%;
		list-style: none;
	}

	.submenu ul li > button {
		display: grid;
		grid-template-columns: auto 1fr auto;
		gap: 0.5rem;
		align-items: center;
		font-size: 0.9rem;
		font-weight: 500;
		padding: 0.25rem 0.75rem;
		text-align: left;
		width: 100%;
		border: 0px;
		color: var(--imag-text-color);
		background-color: var(--imag-100);
		cursor: pointer;
		box-sizing: border-box;
		transition: background-color 0.1s ease;
	}

	.submenu ul li > button:hover {
		background-color: var(--imag-90);
	}

	.submenu ul li > button.disabled {
		color: var(--imag-70);
		cursor: default;
		opacity: 0.5;
	}

	.submenu ul li > button.disabled:hover {
		background-color: var(--imag-100);
	}

	.submenu-arrow {
		opacity: 0.7;
		margin-left: 0.5rem;
		font-size: 0.9em;
	}
</style>
