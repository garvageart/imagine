<script lang="ts" generics="T extends { uid: string } & Record<string, any>">
	import { dev } from "$app/environment";
	import { DateTime } from "luxon";
	import AssetGrid from "./AssetGrid.svelte";
	import PhotoAssetGrid from "./PhotoAssetGrid.svelte";
	import AssetToolbar from "./AssetToolbar.svelte";
	import { type Component, type ComponentProps, type Snippet } from "svelte";
	import type { HTMLButtonAttributes } from "svelte/elements";
	import { type MaterialSymbol } from "material-symbols";
	import type { IPagination } from "$lib/types/asset";
	import Dropdown from "./Dropdown.svelte";
	import type { MenuItem } from "$lib/context-menu/types";
	import { sort } from "$lib/states/index.svelte";
	import { selectionManager } from "$lib/states/selection.svelte";
	import IconButton from "./IconButton.svelte";

	type Props = {
		grid:
			| ComponentProps<typeof AssetGrid<T>>
			| ComponentProps<typeof PhotoAssetGrid>;
		gridComponent?: Component<any>;
		pagination?: IPagination;
		children?: Snippet;
		selectionToolbarSnippet?: Snippet;
		toolbarSnippet?: Snippet;
		noAssetsSnippet?: Snippet;
		showToolbars?: boolean;
		toolbarProps?: Omit<ComponentProps<typeof AssetToolbar>, "children">;
		selectionToolbarProps?: Omit<
			ComponentProps<typeof AssetToolbar>,
			"children"
		>;
	};

	type ToolbarButtonProps = {
		iconName: MaterialSymbol;
		iconStyle?: "sharp" | "outlined" | "rounded";
		text: string;
		dropdown?: Omit<ComponentProps<typeof Dropdown>, "title">;
	} & HTMLButtonAttributes;

	let {
		grid = $bindable(),
		gridComponent = AssetGrid,
		pagination = $bindable({
			limit: 25,
			page: 0
		}),
		children,
		toolbarSnippet,
		noAssetsSnippet,
		showToolbars = $bindable(true),
		toolbarProps,
		selectionToolbarSnippet,
		selectionToolbarProps
	}: Props = $props();

	let assetGridArray: typeof grid.assetGridArray = $state();
	let columnCount: number | undefined = $derived(assetGridArray?.[0]?.length);

	let selectionScope = $derived(
		grid.scopeId ? selectionManager.getScope(grid.scopeId) : null
	);

	let gridData = $derived.by(() => {
		const dataSlice = grid.data.slice(
			0,
			pagination.limit * (pagination.page === 0 ? 1 : pagination.page + 1)
		);

		if (columnCount === undefined) {
			return dataSlice;
		}

		// NOTE: in future this might be an option in the settings
		// fill available space in the last row
		const currentRowImageCount = dataSlice.length % columnCount;
		if (currentRowImageCount === 0) {
			return dataSlice;
		}

		const fillItems = grid.data.slice(
			dataSlice.length,
			dataSlice.length + (columnCount - currentRowImageCount)
		);
		return [...dataSlice, ...fillItems] as typeof dataSlice;
	});

	// Sorting (MenuItem[] for Dropdown)
	let sortOptions: MenuItem[] = [
		{ id: "sort-name", label: "Name" },
		{ id: "sort-created_at", label: "Created At" },
		{ id: "sort-updated_at", label: "Updated At" },
		{ id: "sort-oldest", label: "Oldest" },
		{ id: "sort-most_recent", label: "Most Recent" }
	];

	function currentSortId() {
		switch (sort.by) {
			case "name":
				return "sort-name";
			case "created_at":
				return "sort-created_at";
			case "updated_at":
				return "sort-updated_at";
			case "oldest":
				return "sort-oldest";
			case "most_recent":
				return "sort-most_recent";
		}
	}

	function printGridAsTable() {
		console.log(
			`%cGrid Array at ${DateTime.now().toFormat("dd.MM.yyyy HH:mm:ss")}`,
			"font-weight: bold; color: var(--imag-100); font-size: 18px;"
		);
		console.table(
			assetGridArray?.map((i) => i.map((j) => j.asset?.name ?? j.asset?.uid))
		);
	}
</script>

{#snippet toolbarButton(opts: ToolbarButtonProps)}
	{#if opts.dropdown}
		<Dropdown
			class="toolbar-button"
			{...opts.dropdown}
			title={opts.text}
			icon={opts.iconName}
		/>
	{:else}
		<IconButton
			{...opts}
			iconName={opts.iconName}
			iconStyle={opts.iconStyle}
			class="toolbar-button"
			title={opts.text}
		>
			{#if opts.text.trim()}
				<span style="margin: 0em 0.2em;">{opts.text}</span>
			{/if}
		</IconButton>
	{/if}
{/snippet}

{#if showToolbars}
	{#if selectionScope?.selected && selectionScope.selected.size > 0}
		<AssetToolbar class="selection-toolbar" {...selectionToolbarProps}>
			<div class="selection-info">
				<IconButton
					iconName="close"
					class="toolbar-button"
					title="Clear selection"
					aria-label="Clear selection"
					style="margin-right: 1em;"
					onclick={() => selectionScope.clear()}
				/>
				<span style="font-weight: 600;"
					>{selectionScope.selected.size} selected</span
				>
			</div>
			{@render selectionToolbarSnippet?.()}
		</AssetToolbar>
	{:else}
		<AssetToolbar {...toolbarProps}>
			<div id="asset-tools">
				{@render toolbarButton({
					iconName: "sort",
					text: "Sort by",
					title: "Sort by",
					dropdown: {
						items: sortOptions,
						selectedItemId: currentSortId(),
						onSelect: (item) => {
							switch (item.id) {
								case "sort-name":
									sort.by = "name";
									break;
								case "sort-created_at":
									sort.by = "created_at";
									break;
								case "sort-updated_at":
									sort.by = "updated_at";
									break;
								case "sort-oldest":
									sort.by = "oldest";
									break;
								case "sort-most_recent":
									sort.by = "most_recent";
									break;
							}
						}
					}
				})}
				{#if dev && grid.view === "cards"}
					{@render toolbarButton({
						iconName: "grid_view",
						text: "Print Grid",
						title: "Print Grid to Console",
						onclick: printGridAsTable
					})}
				{/if}
			</div>
			{@render toolbarSnippet?.()}
		</AssetToolbar>
	{/if}
{/if}

{@render children?.()}

{#if gridData.length === 0}
	<div id="viz-no_assets">
		{#if noAssetsSnippet}
			{@render noAssetsSnippet()}
		{:else}
			<p style="text-align: center; margin: 2em; color: var(--imag-10);">
				No assets to display.
			</p>
		{/if}
	</div>
{:else}
	{@const GridComp = gridComponent}
	<GridComp
		{...grid}
		bind:assetGridArray
		bind:data={gridData}
		bind:columnCount
	/>
{/if}

<style lang="scss">
	#asset-tools {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.selection-info {
		display: flex;
		align-items: center;
	}

	:global(.toolbar-button) {
		border-radius: 10em;
		padding: 0.1em 0.3em;
		display: flex;
		align-items: center;
		justify-content: center;
		white-space: nowrap;

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	#viz-no_assets {
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}
</style>
