<script lang="ts">
	import type { MaterialSymbol } from "material-symbols";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import IconButton from "./IconButton.svelte";
	import Button from "./Button.svelte";
	import type { MenuItem } from "$lib/context-menu/types";
	
	interface Props {
		class?: string;
		/** Menu items to render directly in ContextMenu */
		items: MenuItem[];
		/** If provided, shows this id as the current selection and displays a check icon next to it in the menu */
		selectedItemId?: string;
		showMenu?: boolean;
		/** Button text when no selection (or for action menus that don't track selection) */
		title?: string;
		/** Icon to show on the button */
		icon?: MaterialSymbol;
		/** Called when an item is selected */
		onSelect?: (item: MenuItem) => void;
		/** If true, show check icons for selected item. Set to false for action menus. Default: true */
		showSelectionIndicator?: boolean;
		/** Horizontal alignment of the menu relative to the button: 'left' or 'right' */
		align?: "left" | "right";
		/** Debug: forward to ContextMenu to render overlays and logs */
		debug?: boolean;
	}

	let {
		items,
		selectedItemId = $bindable(),
		showMenu = $bindable(false),
		title,
		icon,
		onSelect,
		showSelectionIndicator = true,
		align = "left",
		debug = false,
		class: className
	}: Props = $props();

	let buttonEl: HTMLButtonElement | undefined = $state(undefined);
	let containerEl: HTMLElement | null = $state(null);

	// Derived selected item from items by id
	let selectedItem: MenuItem | undefined = $derived(
		items?.find((i) => i.id === selectedItemId)
	);

	let currentIcon: MaterialSymbol | undefined = $derived(
		(selectedItem?.icon as MaterialSymbol | undefined) ?? icon
	);

	let menuItems: MenuItem[] = $state([]);

	function buildMenuItems(): MenuItem[] {
		return items.map((it) => ({
			...it,
			icon:
				it.icon ??
				(showSelectionIndicator && selectedItemId === it.id
					? "check"
					: undefined),
			// wrap existing action so dropdown selection handling runs first
			action: (e) => handleItemSelect(it, e)
		}));
	}

	function handleItemSelect(item: MenuItem, e: MouseEvent | KeyboardEvent) {
		onSelect?.(item);
		if (showSelectionIndicator) {
			selectedItemId = item.id;
		}
		showMenu = false;
		// Call item's own action if present (forward the event)
		item.action?.(e);
	}

	function toggleMenu() {
		menuItems = buildMenuItems();
		showMenu = !showMenu;
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if (e.key === "Escape") {
			showMenu = false;
		}
	}}
	onclick={(e) => {
		if (!containerEl?.contains(e.target as Node)) {
			showMenu = false;
		}
	}}
	onresize={() => {
		/* ContextMenu handles collisions */
	}}
/>

{#snippet buttonContent()}
	{#if selectedItem}
		<p class="viz-dropdown-title">{selectedItem.label}</p>
	{:else if title}
		<p class="viz-dropdown-title">{title}</p>
	{/if}
{/snippet}

<div class="viz-dropdown-container" bind:this={containerEl}>
	{#if currentIcon}
		<IconButton
			class="viz-dropdown-button {className}"
			weight={300}
			iconName={currentIcon}
			bind:element={buttonEl}
			onclick={toggleMenu}
		>
			{@render buttonContent()}
		</IconButton>
	{:else}
		<Button
			class="viz-dropdown-button {className}"
			bind:element={buttonEl}
			onclick={toggleMenu}
		>
			{@render buttonContent()}
		</Button>
	{/if}

	<!-- Render menu without a positioned wrapper; ContextMenu uses fixed coords anchored to button -->
	<ContextMenu
		bind:showMenu
		items={menuItems}
		anchor={buttonEl as HTMLElement}
		offsetY={0}
		{align}
		{debug}
	/>
</div>

<style lang="scss">
	.viz-dropdown-container {
		position: relative;
		display: inline-block;
	}

	.viz-dropdown-title {
		margin: 0em 0.2rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	:global(.viz-dropdown-button) {
		display: flex;
		align-items: center;
		border-radius: 10em;
		white-space: nowrap;
		cursor: pointer;

		&:focus {
			outline: 2px solid var(--imag-60);
			background-color: var(--imag-80);
		}

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}
</style>
