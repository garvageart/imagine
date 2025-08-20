<script lang="ts" generics="T extends { id: string } & Record<string, any>">
	import { dev } from "$app/environment";
	import { DateTime } from "luxon";
	import AssetGrid from "./AssetGrid.svelte";
	import AssetToolbar from "./AssetToolbar.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import { getContext, untrack, type ComponentProps, type Snippet } from "svelte";
	import type { HTMLButtonAttributes } from "svelte/elements";
	import { type MaterialSymbol } from "material-symbols";
	import type { IPagination } from "$lib/types/asset";
	import Dropdown, { type DropdownOption } from "./Dropdown.svelte";
	import { sort } from "$lib/states/index.svelte";

	type Props = {
		grid: ComponentProps<typeof AssetGrid<T>>;
		pagination: IPagination;
		children: Snippet;
		selectionToolbarSnippet?: Snippet;
		toolbarSnippet?: Snippet;
		toolbarProps?: Omit<ComponentProps<typeof AssetToolbar>, "children">;
		selectionToolbarProps?: Omit<ComponentProps<typeof AssetToolbar>, "children">;
	};

	type ToolbarButtonProps = {
		iconName: MaterialSymbol;
		iconStyle?: "sharp" | "outlined" | "rounded";
		text: string;
		dropdown?: Omit<ComponentProps<typeof Dropdown>, "title">;
	} & HTMLButtonAttributes;

	let {
		grid = $bindable(),
		pagination = $bindable(),
		children,
		toolbarSnippet,
		toolbarProps,
		selectionToolbarSnippet,
		selectionToolbarProps
	}: Props = $props();

	let assetGridArray: typeof grid.assetGridArray = $state();
	let columnCount: number | undefined = $derived(assetGridArray?.[0]?.length);

	let gridData = $derived.by(() => {
		const dataSlice = grid.data.slice(0, pagination.limit * (pagination.offset === 0 ? 1 : pagination.offset + 1));

		if (columnCount === undefined) {
			return dataSlice;
		}

		// NOTE: in future this might be an option in the settings
		// fill available space in the last row
		const currentRowImageCount = dataSlice.length % columnCount;
		if (currentRowImageCount === 0) {
			return dataSlice;
		}

		dataSlice.push(...grid.data.slice(dataSlice.length, dataSlice.length + (columnCount - currentRowImageCount)));

		return dataSlice;
	});

	// Sorting
	let sortOptions: DropdownOption<string>[] = [
		{
			title: "Name"
		},
		{
			title: "Created At"
		},
		{
			title: "Oldest"
		},
		{
			title: "Most Recent"
		}
	];

	function findCurrentSortOption(options: DropdownOption<string>[]) {
		switch (sort.by) {
			case "name":
				return options.find((o) => o.title === "Name");
			case "created_at":
				return options.find((o) => o.title === "Created At");
			case "most_recent":
				return options.find((o) => o.title === "Most Recent");
			case "oldest":
				return options.find((o) => o.title === "Oldest");
		}
	}

	function printGridAsTable() {
		console.log(
			`%cGrid Array at ${DateTime.now().toFormat("dd.MM.yyyy HH:mm:ss")}`,
			"font-weight: bold; color: var(--imag-100); font-size: 18px;"
		);
		console.table(assetGridArray?.map((i) => i.map((j) => j.asset?.name ?? j.asset?.id)));
	}
</script>

{#snippet toolbarButton(opts: ToolbarButtonProps)}
	{#if opts.dropdown}
		<Dropdown class="toolbar-button" {...opts.dropdown} title={opts.text} icon={opts.iconName} />
	{:else}
		<button class="toolbar-button" {...opts} title={opts.text}>
			<MaterialIcon iconName={opts.iconName} iconStyle={opts.iconStyle} />
			{#if opts.text.trim()}
				<span style="margin: 0em 0.2em;">{opts.text}</span>
			{/if}
		</button>
	{/if}
{/snippet}

{#if grid.selectedAssets.size > 1}
	<AssetToolbar class="selection-toolbar" {...selectionToolbarProps}>
		<button
			id="coll-clear-selection"
			class="toolbar-button"
			title="Clear selection"
			aria-label="Clear selection"
			style="margin-right: 1em;"
			onclick={() => grid.selectedAssets.clear()}
		>
			<MaterialIcon iconName="close" />
		</button>
		<span style="font-weight: 600;">{grid.selectedAssets.size} selected</span>
		{@render selectionToolbarSnippet?.()}
	</AssetToolbar>
{:else}
	<AssetToolbar {...toolbarProps}>
		{@render toolbarSnippet?.()}
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
						} else if (option.title === "Oldest") {
							sort.by = "oldest";
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
	</AssetToolbar>
{/if}

{@render children()}

<AssetGrid {...grid} bind:assetGridArray bind:data={gridData} bind:columnCount />

<style lang="scss">
	#asset-tools {
		display: flex;
		align-items: center;
		font-size: 0.9rem;

		& > button {
			margin: 0em 0.5em;
		}
	}

	:global(.toolbar-button) {
		border-radius: 10em;
		min-width: 2em;
		min-height: 2em;
		padding: 0.5em;
		display: flex;
		align-items: center;
		justify-content: center;
		text-wrap: nowrap;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-primary);
			outline: none;
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
