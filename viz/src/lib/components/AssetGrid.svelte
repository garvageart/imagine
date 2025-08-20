<script lang="ts" generics="T extends { id: string } & Record<string, any>">
	import { dev } from "$app/environment";
	import { page } from "$app/state";
	import { buildGridArray } from "$lib/utils/dom";
	import { SvelteSet } from "svelte/reactivity";
	import hotkeys from "hotkeys-js";
	import type { AssetGridArray, AssetSortBy } from "$lib/types/asset";
	import { type Snippet } from "svelte";
	import { sort } from "$lib/states/index.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";

	interface Props {
		data: T[];
		assetSnippet: Snippet<[T]>;
		singleSelectedAsset?: T;
		selectedAssets: SvelteSet<T>;
		assetGridArray?: AssetGridArray<T>;
		columnCount?: number;
		searchValue?: string;
		noAssetsMessage?: string;
		assetDblClick?: (
			e: MouseEvent & {
				currentTarget: EventTarget & (HTMLDivElement | HTMLTableRowElement);
			},
			asset: T
		) => void;
	}

	let {
		data = $bindable(),
		assetSnippet,
		assetGridArray = $bindable(),
		columnCount = $bindable(),
		searchValue = $bindable(""),
		noAssetsMessage = "No assets found",
		singleSelectedAsset = $bindable(),
		selectedAssets = $bindable(new SvelteSet<T>()),
		assetDblClick
	}: Props = $props();

	// HTML Elements
	let assetGridEl: HTMLDivElement | undefined = $state();

	let allAssetsData = $derived.by(() => {
		return data;
	});

	// Inspecting/Debugging
	if (window.debug) {
		$inspect("selected asset", singleSelectedAsset);
	}

	$effect(() => {
		if (!assetGridEl) {
			return;
		}

		assetGridArray = buildAssetGridArray(assetGridEl);
	});

	function handleImageCardSelect(asset: T, e: MouseEvent) {
		if (e.shiftKey) {
			selectedAssets.clear();

			const ids = allAssetsData.map((i: T) => i.id);
			let startIndex = 0;
			const endIndex = ids.indexOf(asset.id);

			if (singleSelectedAsset) {
				startIndex = ids.indexOf(singleSelectedAsset.id);
			}

			const start = Math.min(startIndex, endIndex);
			const end = Math.max(startIndex, endIndex);

			for (let i = start; i <= end; i++) {
				selectedAssets.add(allAssetsData[i]);
			}
		} else if (e.ctrlKey) {
			if (selectedAssets.has(asset)) {
				selectedAssets.delete(asset);
			} else {
				selectedAssets.add(asset);
			}
		} else {
			selectedAssets.clear();
			selectedAssets.add(asset);
			singleSelectedAsset = asset;
		}
	}

	function handleKeydownCardSelect(asset: T, e: KeyboardEvent) {
		if (!assetGridArray) {
			return;
		}

		if (
			e.key !== "ArrowLeft" &&
			e.key !== "ArrowRight" &&
			e.key !== "ArrowUp" &&
			e.key !== "ArrowDown" &&
			e.key !== "Tab" &&
			!e.shiftKey &&
			!e.metaKey
		) {
			return;
		}

		const imageInGridArray = assetGridArray
			.find((i) => i.find((j) => j.asset.id === asset.id))
			?.find((j) => j.asset.id === asset.id);

		if (!imageInGridArray) {
			if (dev) {
				console.warn(`Can't find asset ${asset.id} in grid array`);
			}

			return;
		}

		const columnCount = assetGridArray[0].length;
		const positionIndexInGrid = imageInGridArray.row * columnCount + imageInGridArray.column;
		const imageGridChildren = assetGridEl?.children;

		// Mimic click since we already have a handler for that in `handleImageCardSelect()`
		const focusAndSelectElement = (element: HTMLElement | undefined) => {
			// out of bounds
			if (!element) {
				return;
			}

			// maybe unnessary to blur but i wanna make sure lmao
			(imageGridChildren?.item(positionIndexInGrid) as HTMLElement).blur();
			element.focus();
			element.click();
		};

		switch (e.key) {
			case "ArrowRight":
				const elementRight = imageGridChildren?.item(positionIndexInGrid + 1) as HTMLElement;
				focusAndSelectElement(elementRight);
				break;
			case "ArrowLeft":
				const elementLeft = imageGridChildren?.item(positionIndexInGrid - 1) as HTMLElement;
				focusAndSelectElement(elementLeft);
				break;
			case "ArrowUp":
				const elementUp = imageGridChildren?.item(positionIndexInGrid - columnCount) as HTMLElement;
				focusAndSelectElement(elementUp);
				break;
			case "ArrowDown":
				const elementDown = imageGridChildren?.item(positionIndexInGrid + columnCount) as HTMLElement;
				focusAndSelectElement(elementDown);
				break;
			case "Tab":
				// to break out of the grid by tabbing and focusing we need to let
				// the browser handle the tabbing if we are at the edge of the grid boundary (first and last elements)
				if (e.shiftKey) {
					if (positionIndexInGrid > 0) {
						e.preventDefault();
					}
					focusAndSelectElement(imageGridChildren?.item(positionIndexInGrid - 1) as HTMLElement);
				} else {
					if (positionIndexInGrid < imageGridChildren?.length! - 1) {
						e.preventDefault();
					}
					focusAndSelectElement(imageGridChildren?.item(positionIndexInGrid + 1) as HTMLElement);
				}
				break;
		}
	}

	function buildAssetGridArray(element: HTMLElement) {
		const array = buildGridArray(element).map((i) => {
			return i.map((j) => {
				const assetId = j.element?.firstElementChild?.getAttribute("data-asset-id")!;
				const asset = allAssetsData.find((i: T) => i.id === assetId)!;

				return {
					asset,
					row: j.row,
					column: j.column,
					columnSize: j.columnSize,
					size: j.size,
					rowSize: j.rowSize
				};
			});
		});

		return array;
	}

	function unselectImagesOnClickOutsideAssetContainer(element: HTMLElement) {
		const clickHandler = (e: MouseEvent) => {
			const target = e.target as HTMLElement;
			const selectionToolbar = target.closest(".selection-toolbar") as HTMLElement | undefined;

			// ignore the selection toolbar since this is what we use do actions
			if (target === selectionToolbar || selectionToolbar?.contains(target)) {
				return;
			}

			const imageCard = target.closest(".image-card") as HTMLElement | undefined;
			const isGridButNotImageCard = target === element && !imageCard;
			if (!element.contains(target) || isGridButNotImageCard) {
				singleSelectedAsset = undefined;
				selectedAssets.clear();
			}
		};

		$effect(() => {
			document.addEventListener("click", clickHandler);

			return () => {
				document.removeEventListener("click", clickHandler);
			};
		});
	}

	hotkeys("meta+a", (e) => {
		e.preventDefault();
		selectedAssets.clear();
		allAssetsData.forEach((i) => selectedAssets.add(i));
	});

	hotkeys("esc", (e) => {
		e.preventDefault();
		selectedAssets.clear();
	});
</script>

{#snippet assetComponentCard(assetData: T)}
	{@const isSelected = selectedAssets.values().some((i) => i.id === assetData.id) || singleSelectedAsset === assetData}
	<div
		class="asset-card"
		class:max-width-column={columnCount !== undefined && columnCount > 1}
		class:selected-card={isSelected}
		role="button"
		tabindex="0"
		onfocus={(e) => {
			if (!e.currentTarget.classList.contains("selected-card")) {
				e.currentTarget.classList.add("selected-card");
			}
		}}
		onblur={(e) => {
			if (e.currentTarget.classList.contains("selected-card") && !isSelected) {
				e.currentTarget.classList.remove("selected-card");
			}
		}}
		onclick={(e) => {
			e.preventDefault();
			handleImageCardSelect(assetData, e);
		}}
		onkeydown={(e) => {
			e.preventDefault();
			handleKeydownCardSelect(assetData, e);
		}}
		ondblclick={(e) => {
			if (e.ctrlKey) {
				e.preventDefault();
				return;
			}

			assetDblClick?.(e, assetData);
		}}
	>
		{@render assetSnippet(assetData)}
	</div>
{/snippet}

{#snippet assetComponentListOption(assetData: T)}
	{@const isSelected = selectedAssets.values().some((i) => i.id === assetData.id) || singleSelectedAsset === assetData}
	<tr
		class="asset-card"
		class:selected-card={isSelected}
		role="button"
		tabindex="0"
		onfocus={(e) => {
			if (!e.currentTarget.classList.contains("selected-card")) {
				e.currentTarget.classList.add("selected-card");
			}
		}}
		onblur={(e) => {
			if (e.currentTarget.classList.contains("selected-card") && !isSelected) {
				e.currentTarget.classList.remove("selected-card");
			}
		}}
		onclick={(e) => {
			handleImageCardSelect(assetData, e);
		}}
		onkeydown={(e) => {
			handleKeydownCardSelect(assetData, e);
		}}
		ondblclick={(e) => {
			if (e.ctrlKey) {
				e.preventDefault();
				return;
			}

			assetDblClick?.(e, assetData);
		}}
	>
		<td>
			{@render assetSnippet(assetData)}
		</td>
	</tr>
{/snippet}

{#snippet assetTable()}
	<div class="viz-asset-table-container">
		<table>
			<thead>
				<tr>
					{#each Object.keys(allAssetsData[0]) as key}
						<th>
							<button
								onclick={() => {
									if (sort.by === key) {
										sort.order = sort.order === "asc" ? "desc" : "asc";
									} else {
										sort.by = key as AssetSortBy;
									}
								}}
							>
								<MaterialIcon iconName="arrow_{sort.by === key && sort.order === 'asc' ? 'upward' : 'downward'}" />
								{key}
							</button>
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each allAssetsData as asset}
					{@render assetComponentListOption(asset)}
				{/each}
			</tbody>
		</table>
	</div>
{/snippet}

{#if allAssetsData.length === 0}
	{#if searchValue}
		<div class="no-results">
			<p>No results found for "{searchValue}"</p>
		</div>
	{:else}
		<div>
			<p>{noAssetsMessage}</p>
		</div>
	{/if}
{:else if sort.display === "list"}
	{@render assetTable()}
{:else}
	<div
		bind:this={assetGridEl}
		class="viz-asset-grid-container"
		style="padding: 0em {page.url.pathname === '/' ? '1em' : '2em'}; {columnCount !== undefined && columnCount > 1
			? `grid-template-columns: repeat(${columnCount}, minmax(15em, 1fr));`
			: ''}"
		use:unselectImagesOnClickOutsideAssetContainer
	>
		{#each allAssetsData as asset}
			{@render assetComponentCard(asset)}
		{/each}
	</div>
{/if}

<style lang="scss">
	.viz-asset-grid-container {
		box-sizing: border-box;
		margin: 2em 0em;
		display: grid;
		gap: 1em;
		width: 100%;
		max-width: 100%;
		text-overflow: clip;
		justify-content: center;
		grid-template-columns: repeat(auto-fit, minmax(15em, 1fr));
	}

	.asset-card {
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		border-radius: 0.5em;
		overflow: hidden;
	}

	.asset-card.max-width-column {
		max-width: 15em;
	}

	.selected-card {
		outline: 2px solid var(--imag-primary);
	}
</style>
