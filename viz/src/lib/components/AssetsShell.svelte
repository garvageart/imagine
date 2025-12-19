<script lang="ts" generics="T extends { uid: string } & Record<string, any>">
	import { dev } from "$app/environment";
	import { DateTime } from "luxon";
	import AssetGrid from "./AssetGrid.svelte";
	import PhotoAssetGrid from "./PhotoAssetGrid.svelte";
	import AssetToolbar from "./AssetToolbar.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import { type Component, type ComponentProps, type Snippet } from "svelte";
	import type { HTMLButtonAttributes } from "svelte/elements";
	import { type MaterialSymbol } from "material-symbols";
	import type { IPagination } from "$lib/types/asset";
	import Dropdown, { type DropdownOption } from "./Dropdown.svelte";
	import { sort } from "$lib/states/index.svelte";
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
		toolbarProps,
		selectionToolbarSnippet,
		selectionToolbarProps
	}: Props = $props();

	let assetGridArray: typeof grid.assetGridArray = $state();
	let columnCount: number | undefined = $derived(assetGridArray?.[0]?.length);

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

	// Sorting
	let sortOptions: DropdownOption[] = [
		{
			title: "Name"
		},
		{
			title: "Created At"
		},
		{
			title: "Updated At"
		},
		{
			title: "Oldest"
		},
		{
			title: "Most Recent"
		}
	];

	function findCurrentSortOption(options: DropdownOption[]) {
		switch (sort.by) {
			case "name":
				return options.find((o) => o.title === "Name");
			case "created_at":
				return options.find((o) => o.title === "Created At");
			case "updated_at":
				return options.find((o) => o.title === "Updated At");
			case "oldest":
				return options.find((o) => o.title === "Oldest");
			case "most_recent":
				return options.find((o) => o.title === "Most Recent");
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

{#if gridData.length > 0}
	{#if grid.selectedAssets && grid.selectedAssets.size > 1}
		<AssetToolbar class="selection-toolbar" {...selectionToolbarProps}>
			<button
				id="coll-clear-selection"
				class="toolbar-button"
				title="Clear selection"
				aria-label="Clear selection"
				style="margin-right: 1em;"
				onclick={() => grid.selectedAssets?.clear()}
			>
				<MaterialIcon iconName="close" />
			</button>
			<span style="font-weight: 600;">{grid.selectedAssets.size} selected</span>
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
						options: sortOptions,
						selectedOption: findCurrentSortOption(sortOptions),
						onSelect: (option) => {
							if (option.title === "Name") {
								sort.by = "name";
							} else if (option.title === "Created At") {
								sort.by = "created_at";
							} else if (option.title === "Updated At") {
								sort.by = "updated_at";
							} else if (option.title === "Oldest") {
								sort.by = "oldest";
							} else if (option.title === "Most Recent") {
								sort.by = "most_recent";
							}
						}
					}
				})}
				{#if dev}
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

	#viz-no_assets {
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}
</style>
