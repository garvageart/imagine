<script lang="ts" module>
	export type DropdownOption = {
		title: string;
		icon?: MaterialSymbol;
		disabled?: boolean;
	};
</script>

<script lang="ts" generics="T">
	import { isEqual } from "lodash-es";
	import MaterialIcon from "./MaterialIcon.svelte";
	import type { MaterialSymbol } from "material-symbols";
	import ContextMenu, {
		type MenuItem
	} from "$lib/context-menu/ContextMenu.svelte";
	import IconButton from "./IconButton.svelte";
	import Button from "./Button.svelte";

	interface Props {
		class?: string;
		options: DropdownOption[];
		/** If provided, shows this as the current selection and displays a check icon next to it in the menu */
		selectedOption?: DropdownOption;
		showMenu?: boolean;
		/** Button text when no selection (or for action menus that don't track selection) */
		title?: string;
		/** Icon to show on the button */
		icon?: MaterialSymbol;
		/** Called when an option is selected */
		onSelect?: (option: DropdownOption) => void;
		/** If true, show check icons for selected option. Set to false for action menus. Default: true */
		showSelectionIndicator?: boolean;
		/** Horizontal alignment of the menu relative to the button: 'left' or 'right' */
		align?: "left" | "right";
		/** Debug: forward to ContextMenu to render overlays and logs */
		debug?: boolean;
	}

	let {
		options,
		selectedOption = $bindable(),
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

    let currentIcon: MaterialSymbol | undefined = $derived(selectedOption?.icon ?? icon);

	// Build MenuItem array for ContextMenu
	function buildMenuItems(): MenuItem[] {
		return options.map((opt, idx) => ({
			id: `opt-${idx}`,
			label: opt.title,
			// Show option's icon if provided, or check mark if this is the selected option (when showSelectionIndicator is true)
			icon:
				opt.icon ??
				(showSelectionIndicator && isEqual(selectedOption, opt)
					? ("check" as any)
					: undefined),
			disabled: opt.disabled,
			action: (e) => handleOptionSelect(opt as any)
		}));
	}

	let menuItems: MenuItem[] = $state([]);

	function handleOptionSelect(option: DropdownOption) {
		onSelect?.(option);
		if (showSelectionIndicator) {
			selectedOption = option;
		}
		showMenu = false;
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
	{#if selectedOption}
		<p class="viz-dropdown-title">
			{selectedOption.title}
		</p>
	{:else}
		{#if title}
			<p class="viz-dropdown-title">
				{title}
			</p>
		{/if}
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
