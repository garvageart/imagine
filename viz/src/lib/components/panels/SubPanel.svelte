<script module lang="ts">
	import Splitpanes from "$lib/third-party/svelte-splitpanes/Splitpanes.svelte";

	export type InternalSubPanelContainer = Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">;
	export type InternalPanelContainer = Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
	export type Content = Omit<VizSubPanel, "childs" | "id"> & { id?: string; views: VizView[] };
	export type SubPanelChilds = {
		internalSubPanelContainer: InternalSubPanelContainer;
		internalPanelContainer: InternalPanelContainer;
		content: Content[];
	};

	export type VizSubPanel = Props &
		ComponentProps<typeof Pane> & {
			childs: SubPanelChilds;
		};
</script>

<script lang="ts">
	import { setContext, untrack, type ComponentProps, type Snippet } from "svelte";
	import { Pane } from "$lib/third-party/svelte-splitpanes";
	import MaterialIcon from "../MaterialIcon.svelte";
	import { dev } from "$app/environment";
	import type { TabData } from "$lib/views/tabs.svelte";
	import TabOps from "$lib/views/tabs.svelte";
	import VizView from "$lib/views/views.svelte";
	import { measureComponentRenderTimes, resetAndReloadLayout } from "$lib/dev/components.svelte";
	import { views } from "$lib/layouts/views";
	import LoadingContainer from "../LoadingContainer.svelte";
	import { isElementScrollable } from "$lib/utils/dom";
	import { findSubPanel, generateKeyId, isVizSubPanelData } from "$lib/utils/layout";
	import { goto } from "$app/navigation";
	import ContextMenu, { type MenuItem } from "$lib/context-menu/ContextMenu.svelte";
	import { layoutState, layoutTree } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import VizSubPanelData, { Content as ContentClass } from "$lib/layouts/subpanel.svelte";
	import {
		cleanupEmptyPanels,
		duplicateView,
		normalizePanelSizes,
		removeEmptyContent,
		splitPanelHorizontally,
		splitPanelVertically
	} from "$lib/layouts/panel-operations";

	if (dev) {
		window.resetAndReloadLayout = resetAndReloadLayout;
	}

	if (dev || window.debug) {
		measureComponentRenderTimes();
	}

	interface Props {
		id: string;
		header?: boolean;
		views: VizView[];
		children?: Snippet;
	}

	// CSS was a mistake (or I'm an idiot)
	let mainHeaderHeight = $state(document.querySelector("header")?.clientHeight ?? 0);

	const defaultClass = "viz-panel";
	let className: string = $state(defaultClass);

	const allProps: Props & ComponentProps<typeof Pane> = $props();

	let id = allProps.id;
	let header = allProps.header ?? false;

	const children = allProps.children;
	const keyId = allProps.paneKeyId ?? generateKeyId();
	const minSize = allProps.minSize ?? 10;

	// pane size overrides when locked to prevent resizing
	let paneMinSize: number | undefined = $state(minSize);
	let paneMaxSize: number | undefined = $state(allProps.maxSize);

	$effect(() => {
		const res = findSubPanel("paneKeyId", keyId);
		if (res && isVizSubPanelData(res.subPanel) && res.subPanel.locked) {
			// Prefer the internalSubPanelContainer.size, fallback to panel size, then minSize
			const lockedSize = res.subPanel.childs?.internalSubPanelContainer?.size ?? res.subPanel.size ?? minSize;
			paneMinSize = lockedSize ?? minSize;
			paneMaxSize = lockedSize ?? paneMinSize;
			console.debug("[Viz] panel", keyId, "locked -> size", paneMinSize);
		} else {
			paneMinSize = minSize;
			paneMaxSize = allProps.maxSize;
			console.debug("[Viz] panel", keyId, "unlocked -> min", paneMinSize, "max", paneMaxSize);
		}
	});

	// Helper: match svelte-kit style dynamic routes like "/collections/[uid]" to concrete paths
	function pathMatches(pattern: string | undefined, actual: string | undefined): boolean {
		if (!pattern || !actual) {
			return false;
		}

		if (pattern === actual) {
			return true;
		}

		// Escape regex specials, then turn dynamic segments \[param\] into [^/]+
		const escaped = pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&").replace(/\\\[[^\]]+\\\]/g, "[^/]+");
		const re = new RegExp("^" + escaped + "$");
		return re.test(actual);
	}

	// construct the views from the stored data
	const initialViews = allProps.views ?? [];
	let panelViews = $state(initialViews);

	for (let i = 0; i < initialViews.length; i++) {
		const storedView = initialViews[i];

		// If already a VizView instance (from testLayout), skip reconstruction
		if (storedView instanceof VizView) {
			storedView.parent = keyId;
			continue;
		}

		// Otherwise, it's serialized data from localStorage - hydrate it
		const { id: panelViewId, name: panelViewName, path: panelViewPath } = storedView;

		// Try to find by path (supports dynamic segments), then by name, then by id
		const matchedView = views.find((view) => {
			if (panelViewPath && view.path && pathMatches(view.path, panelViewPath)) return true;
			return view.name === panelViewName || view.id === panelViewId;
		});

		if (!matchedView?.component) {
			console.warn(`Could not find component for view: ${panelViewName} (id: ${panelViewId}, path: ${panelViewPath})`);
			continue;
		}

		// Reconstruct VizView instance from serialized data
		const v = VizView.fromJSON(storedView as any, matchedView.component);
		v.parent = keyId;

		initialViews[i] = v;
	}

	if (allProps.class) {
		className = allProps.class;
	}

	if (initialViews.length > 0) {
		header = true;
	}

	if (header === true && initialViews.length === 0) {
		throw new Error("Viz: Header is showing, but no tabs are provided for: " + keyId);
	}

	const storedActiveView = $derived(panelViews.find((view) => view.isActive === true));
	let activeView = $derived(storedActiveView ?? panelViews[0]);
	let panelData = $derived(activeView.viewData ?? activeView?.getComponentData());

	function isPanelLocked(): boolean {
		const result = findSubPanel("paneKeyId", keyId)?.subPanel;
		if (!result) return false;
		if (isVizSubPanelData(result)) return result.locked;
		return false;
	}

	let subPanelContentElement: HTMLDivElement | undefined = $state();
	let subPanelContentFocused = $state(false);

	let showContextMenu = $state(false);
	let contextMenuItems = $state<MenuItem[]>([]);
	let contextMenuAnchor = $state<{ x: number; y: number } | null>(null);
	let contextMenuTargetView: VizView | null = $state(null);

	if (window.debug === true) {
		$inspect("active view", keyId, activeView);
		if (initialViews.length) {
			$effect(() => {
				(async () => {
					const data = await panelData;
					console.log("panel data", keyId, $state.snapshot(data));
				})();
			});
		}
		$inspect("panel views", keyId, panelViews);
	}

	let tabDropper: TabOps;

	if (initialViews.length) {
		tabDropper = new TabOps(initialViews);
		setContext<Content>("content", {
			paneKeyId: keyId,
			views: initialViews
		});
	}

	$effect(() => {
		if (panelViews.length) {
			const element = subPanelContentElement;
			if (!element) {
				return;
			}

			const lastChild = element.lastElementChild as HTMLElement;
			if (!lastChild) {
				return;
			}

			if (subPanelContentFocused) {
				if (isElementScrollable(lastChild)) {
					element.classList.add("with__scrollbar");
				}
				element.classList.add("splitpanes__pane__active");
			} else {
				element.classList.remove("with__scrollbar");
				element.classList.remove("splitpanes__pane__active");
			}
		}
	});

	// make the last view in the panel active if the current active view is removed
	$effect(() => {
		if (!panelViews.find((view) => view.id === activeView?.id)) {
			activeView = panelViews[panelViews.length - 1];
		}
	});

	$effect(() => {
		if (tabDropper?.activeView) {
			activeView = tabDropper?.activeView;
		}
	});

	$effect(() => {
		if (activeView) {
			// will loop endlessly without it
			untrack(() => {
				updateSubPanelActiveView(activeView);
			});
		}
	});

	function tabDragable(node: HTMLElement, data: TabData) {
		// Prevent tab dragging when panel is locked
		if (isPanelLocked()) {
			return { destroy: () => {} };
		}

		return tabDropper.draggable(node, data);
	}

	function onDropOver(event: DragEvent) {
		return tabDropper.onDropOver(event);
	}

	function tabDrop(node: HTMLElement) {
		return tabDropper.tabDrop(node);
	}

	function headerDraggable(node: HTMLElement) {
		if (isPanelLocked()) {
			return { destroy: () => {} };
		}

		// If header dragging is implemented later, return its destroy here.
		return { destroy: () => {} };
	}

	function subPanelDrop(node: HTMLElement, data: TabData) {
		return tabDropper.subPanelDropInside(node, data);
	}

	function makeViewActive(view: VizView) {
		if (view.id === activeView.id) {
			return;
		}

		activeView.isActive = false;
		view.isActive = true;

		updateSubPanelActiveView(view);
	}

	/**
	 * Updates the active view of a subpanel based on the given key ID.
	 *
	 * Finds the subpanel associated with the provided key ID
	 * and sets the specified view as the active view. If there is a current
	 * active view, it is deactivated before activating the new view.
	 * The views array in the subpanel is updated to ensure that the new
	 * active view is correctly reflected.
	 *
	 * @param view - The view to be set as the active view.
	 */
	function updateSubPanelActiveView(view: VizView) {
		const subPanel = findSubPanel("paneKeyId", keyId)?.subPanel;

		if (!subPanel) {
			if (dev) {
				throw new Error("Viz: Subpanel not found");
			}

			console.error("Viz: Subpanel not found");
			return;
		}

		subPanel.views.splice(
			subPanel.views.findIndex((spview) => spview.id === view.id),
			1,
			view
		);
	}

	/**
	 * Closes a specific tab/view
	 */
	function closeTab(view: VizView) {
		if (isPanelLocked()) return;
		const index = panelViews.findIndex((v) => v.id === view.id);
		if (index === -1) {
			return;
		}

		panelViews.splice(index, 1);

		// Find the parent panel and content group
		const result = findSubPanel("paneKeyId", keyId);
		if (!result) return;

		const { parentIndex, childIndex, isChild } = result;
		const currentPanel = layoutState.tree[parentIndex];

		if (isChild && currentPanel) {
			const currentContent = currentPanel.childs.content[childIndex];
			if (currentContent) {
				currentContent.views = panelViews;

				// Clean up empty content and panels
				removeEmptyContent(currentPanel, childIndex);
				currentPanel.views = currentPanel.childs.content.flatMap((c) => c.views);

				if (currentPanel.views.length === 0) {
					layoutState.tree.splice(parentIndex, 1);
					normalizePanelSizes(layoutState.tree);
				}
			}
		}

		// If we closed the active tab, activate another one
		if (activeView.id === view.id && panelViews.length > 0) {
			// Activate the tab to the left, or the first tab if we closed index 0
			const newActiveIndex = Math.max(0, index - 1);
			makeViewActive(panelViews[newActiveIndex]);
		}
	}

	/**
	 * Closes all tabs except the specified one
	 */
	function closeOtherTabs(exceptView: VizView) {
		if (isPanelLocked()) return;
		panelViews = panelViews.filter((v) => v.id === exceptView.id);

		const subPanel = findSubPanel("paneKeyId", keyId)?.subPanel;
		if (subPanel) {
			subPanel.views = panelViews;
		}

		if (activeView.id !== exceptView.id) {
			makeViewActive(exceptView);
		}
	}

	/**
	 * Closes all tabs to the right of the specified tab
	 */
	function closeTabsToRight(view: VizView) {
		if (isPanelLocked()) return;
		const index = panelViews.findIndex((v) => v.id === view.id);
		if (index === -1 || index === panelViews.length - 1) return;

		const viewsToKeep = panelViews.slice(0, index + 1);
		const closedActiveView = !viewsToKeep.some((v) => v.id === activeView.id);

		panelViews = viewsToKeep;

		const subPanel = findSubPanel("paneKeyId", keyId)?.subPanel;
		if (subPanel) {
			subPanel.views = panelViews;
		}

		// If active view was closed, activate the rightmost remaining tab
		if (closedActiveView && panelViews.length) {
			makeViewActive(panelViews[panelViews.length - 1]);
		}
	}

	/**
	 * Closes all tabs in this panel
	 */
	function closeAllTabs() {
		if (isPanelLocked()) return;
		panelViews = [];

		const subPanel = findSubPanel("paneKeyId", keyId)?.subPanel;
		if (subPanel) {
			subPanel.views = panelViews;
		}
	}

	/**
	 * Splits the current panel and moves a view to a new panel on the right
	 */
	function splitRight(view: VizView) {
		if (isPanelLocked()) return;
		const result = findSubPanel("paneKeyId", keyId);
		if (!result) return;

		const { parentIndex } = result;

		// Create a new view instance and split the panel
		const newView = duplicateView(view);
		const newPanel = splitPanelVertically(layoutState.tree, parentIndex, newView);

		if (newPanel) {
			newView.setActive(true);
		}
	}

	/**
	 * Splits the current panel and moves a view to a new content group below within the same parent
	 */
	function splitDown(view: VizView) {
		if (isPanelLocked()) return;
		const result = findSubPanel("paneKeyId", keyId);
		if (!result) return;

		const { parentIndex } = result;
		const currentPanel = layoutState.tree[parentIndex];

		// Create a new view instance and split horizontally
		const newView = duplicateView(view);
		const newContent = splitPanelHorizontally(currentPanel, newView);

		if (newContent) {
			newView.setActive(true);
		}
	}

	/**
	 * Moves a view to an existing panel group
	 */
	function moveToPanel(view: VizView, direction: "left" | "right" | "up" | "down") {
		if (isPanelLocked()) return;
		// Resolve the location of the view by its parent id to be precise
		const viewParentId = view.parent ?? keyId;
		const result = findSubPanel("paneKeyId", viewParentId);
		if (!result) return;

		if (dev || window.debug) {
			const layoutSummary = layoutState.tree.map((p) => ({
				paneKeyId: p.paneKeyId,
				contentCount: p.childs?.content?.length ?? 0
			}));
			console.debug("[Viz] moveToPanel start", { viewId: view.id, viewParent: view.parent, result, layoutSummary, direction });
		}

		let { parentIndex, isChild, childIndex } = result;
		const currentPanel = layoutState.tree[parentIndex];

		// If findSubPanel returned the top-level panel (isChild === false), determine which content group contains this view
		if (!isChild) {
			const foundIdx = currentPanel.childs?.content?.findIndex((c) => c.views.some((v) => v.id === view.id)) ?? -1;
			if (foundIdx === -1) {
				// Nothing to move
				return;
			}
			childIndex = foundIdx;
			isChild = true;
		}

		const currentContent = currentPanel.childs.content[childIndex];

		// Remove the view from current content group
		if (currentContent) {
			const viewIdx = currentContent.views.findIndex((v) => v.id === view.id);
			if (viewIdx !== -1) {
				currentContent.views.splice(viewIdx, 1);
			}
		}

		if (dev || window.debug) {
			const layoutSummary2 = layoutState.tree.map((p) => ({
				paneKeyId: p.paneKeyId,
				contentCount: p.childs?.content?.length ?? 0
			}));
			console.debug("[Viz] moveToPanel after removal", {
				viewId: view.id,
				parentIndex,
				childIndex,
				isChild,
				currentContentCount: currentPanel.childs.content.length,
				layoutSummary2
			});
		}

		// Determine target panel/content group
		let targetPanelIndex = parentIndex;
		let targetContentIndex = childIndex;

		const resolvedChildIndex = isChild
			? childIndex
			: (currentPanel.childs?.content?.findIndex((c) => c.views.some((v) => v.id === view.id)) ?? -1);

		if (direction === "left") {
			if (resolvedChildIndex > 0) {
				targetPanelIndex = parentIndex;
				targetContentIndex = resolvedChildIndex - 1;
			} else if (parentIndex > 0) {
				// move into the previous panel's rightmost content group
				targetPanelIndex = parentIndex - 1;
				const prevPanel = layoutState.tree[targetPanelIndex];
				targetContentIndex = Math.max(0, (prevPanel.childs?.content?.length ?? 1) - 1);
			}
		} else if (direction === "right") {
			if (resolvedChildIndex !== -1 && resolvedChildIndex < (currentPanel.childs?.content?.length ?? 0) - 1) {
				targetPanelIndex = parentIndex;
				targetContentIndex = resolvedChildIndex + 1;
			} else if (parentIndex < layoutState.tree.length - 1) {
				// move into the next panel's leftmost content group
				targetPanelIndex = parentIndex + 1;
				targetContentIndex = 0;
			}
		} else if (direction === "up" && parentIndex > 0) {
			targetPanelIndex = parentIndex - 1;
			targetContentIndex = 0;
		} else if (direction === "down" && parentIndex < layoutState.tree.length - 1) {
			targetPanelIndex = parentIndex + 1;
			targetContentIndex = 0;
		}

		// Add view to target content group
		const targetPanel = layoutState.tree[targetPanelIndex];
		const targetContent = targetPanel?.childs?.content?.[targetContentIndex];
		if (targetContent) {
			targetContent.views.push(view);
			view.parent = targetContent.paneKeyId;
			targetPanel.views = targetPanel.childs.content.flatMap((c) => c.views);
			view.setActive(true);
		}

		cleanupEmptyPanels(layoutState.tree);
	}

	/**
	 * Shows context menu for a tab
	 */
	function showTabContextMenu(event: MouseEvent, view: VizView) {
		event.preventDefault();
		event.stopPropagation();

		contextMenuTargetView = view;
		contextMenuAnchor = { x: event.clientX, y: event.clientY };

		const viewIndex = panelViews.findIndex((v) => v.id === view.id);
		const isLastTab = viewIndex === panelViews.length - 1;
		const isOnlyTab = panelViews.length === 1;

		// Determine move availability by resolving the view's actual parent
		const viewParentId = view.parent ?? keyId;
		const viewLoc = findSubPanel("paneKeyId", viewParentId);
		let canMoveLeft = false,
			canMoveRight = false,
			canMoveUp = false,
			canMoveDown = false;

		if (viewLoc) {
			const { parentIndex, childIndex, isChild } = viewLoc as any;
			const currentPanel = layoutState.tree[parentIndex];

			// Use the isChild flag: when isChild is true, childIndex is the content index.
			// When isChild is false, the paneKey matched a top-level panel and we must search its content groups.
			const resolvedChildIndex = isChild
				? childIndex
				: (currentPanel.childs?.content?.findIndex((c) => c.views.some((v) => v.id === view.id)) ?? -1);

			canMoveLeft = resolvedChildIndex > 0;
			// Allow moving left/right into adjacent panels as well as sibling content groups
			canMoveRight = resolvedChildIndex !== -1 && resolvedChildIndex < (currentPanel.childs?.content?.length ?? 0) - 1;
			// If there is no sibling to the left/right, but there is a panel to the left/right,
			// treat that as a valid move target too.
			if (!canMoveLeft && parentIndex > 0) {
				canMoveLeft = true;
			}
			if (!canMoveRight && parentIndex < layoutState.tree.length - 1) {
				canMoveRight = true;
			}
			// Respect top-level split orientation: only allow up/down when layoutTree.horizontal is true
			const panelsAreStacked = !!layoutTree.horizontal;
			canMoveUp = panelsAreStacked && parentIndex > 0;
			canMoveDown = panelsAreStacked && parentIndex < layoutState.tree.length - 1;
		}

		contextMenuItems = [
			{
				id: isPanelLocked() ? "unlock-panel" : "lock-panel",
				label: isPanelLocked() ? "Unlock Panel" : "Lock Panel",
				action: () => {
					const result = findSubPanel("paneKeyId", keyId)?.subPanel;
					if (result && isVizSubPanelData(result)) {
						result.locked = !result.locked;
						console.debug("[Viz] toggled lock for", keyId, "now", result.locked);
					}
					showContextMenu = false;
				},
				icon: isPanelLocked() ? "lock_open" : "lock",
				danger: false
			},
			{
				id: "separator-lock",
				label: "",
				separator: true
			},
			{
				id: "close",
				label: "Close Tab",
				action: () => closeTab(view),
				icon: "close",
				shortcut: "Ctrl+W"
			},
			{
				id: "close-others",
				label: "Close Other Tabs",
				action: () => closeOtherTabs(view),
				icon: "tab_close",
				disabled: isOnlyTab
			},
			{
				id: "close-right",
				label: "Close Tabs to the Right",
				action: () => closeTabsToRight(view),
				icon: "close_fullscreen",
				disabled: isLastTab || isOnlyTab
			},
			{
				id: "separator1",
				label: "",
				separator: true
			},
			{
				id: "split-right",
				label: "Split Right",
				action: () => splitRight(view),
				icon: "vertical_split"
			},
			{
				id: "split-down",
				label: "Split Down",
				action: () => splitDown(view),
				icon: "horizontal_split"
			},
			{
				id: "separator2",
				label: "",
				separator: true
			},
			{
				id: "move-left",
				label: "Move to Left Group",
				action: () => moveToPanel(view, "left"),
				icon: "arrow_back",
				disabled: !canMoveLeft
			},
			{
				id: "move-right",
				label: "Move to Right Group",
				action: () => moveToPanel(view, "right"),
				icon: "arrow_forward",
				disabled: !canMoveRight
			},
			{
				id: "move-up",
				label: "Move to Above Group",
				action: () => moveToPanel(view, "up"),
				icon: "arrow_upward",
				disabled: !canMoveUp
			},
			{
				id: "move-down",
				label: "Move to Below Group",
				action: () => moveToPanel(view, "down"),
				icon: "arrow_downward",
				disabled: !canMoveDown
			},
			{
				id: "separator3",
				label: "",
				separator: true
			},
			{
				id: "close-all",
				label: "Close All Tabs",
				action: () => closeAllTabs(),
				icon: "cancel_presentation",
				danger: true
			}
		];

		showContextMenu = true;
	}
</script>

<svelte:document
	on:click={(event) => {
		const target = event.target as HTMLElement;
		const element = subPanelContentElement;

		if (!element) {
			return;
		}

		if (!element.contains(target)) {
			subPanelContentFocused = false;
		}
	}}
/>

<ContextMenu bind:showMenu={showContextMenu} items={contextMenuItems} anchor={contextMenuAnchor} />

<Pane class={className} {...allProps} {id} paneKeyId={keyId} minSize={paneMinSize} maxSize={paneMaxSize}>
	<!--
TODO:
Make the header draggable too. Use the same drag functions. If we're dragging
a header into a different panel, place that panel in place and update the state
for Splitpanes
	-->
	{#if panelViews.length > 0}
		<div
			class="viz-sub_panel-header {isPanelLocked() ? 'locked' : ''}"
			role="tablist"
			tabindex="0"
			use:headerDraggable
			use:tabDrop
			ondragover={(event) => onDropOver(event)}
		>
			{#each panelViews as view, i}
				{#if view.name && view.name.trim() != ""}
					{@const tabNameId = view.name.toLowerCase().replaceAll(" ", "-")}
					{@const data = { index: i, view: view }}
					<button
						id={tabNameId + "-tab"}
						class="viz-tab-button {activeView.id === view.id ? 'active-tab' : ''}"
						data-tab-id={view.id}
						role="tab"
						title={view.name}
						aria-label={view.name}
						onclick={async () => {
							if (dev) {
								if (activeView.id === view.id) {
									if (view.path) {
										await goto(view.path);
										return;
									}
								}
							}

							if (activeView.id === view.id) {
								// show context menu or maybe add an onclick
								// property to the class
							}

							makeViewActive(view instanceof VizView ? view : new VizView(view));
						}}
						oncontextmenu={(e) => showTabContextMenu(e, view)}
						use:tabDragable={data}
						use:tabDrop
						ondragover={(event) => onDropOver(event)}
					>
						<!--
					Every tab name needs to manually align itself with the icon
					Translate is used instead of margin or position is used to avoid
					shifting the layout  
					-->
						<MaterialIcon style={`transform: translateY(${view.opticalCenterFix}px);`} weight={200} iconName="menu" />
						<span class="viz-sub_panel-name">{view.name}</span>
					</button>
				{/if}
			{/each}
			{#if dev}
				<button
					id="viz-debug-button"
					class="viz-tab-button"
					aria-label="Reset and Reload"
					title="Reset and Reload"
					onclick={() => resetAndReloadLayout()}
				>
					<span class="viz-sub_panel-name">Reset Layout</span>
					<MaterialIcon iconName="refresh" />
				</button>
			{/if}
			{#if isPanelLocked()}
				<button class="viz-lock-indicator" aria-label="Panel locked" title="Panel is locked">
					<MaterialIcon iconName="lock" />
				</button>
			{/if}
		</div>
	{/if}
	{#if activeView?.component}
		{@const Comp = activeView.component}
		{@const data = { index: panelViews.findIndex((view) => view.id === activeView.id), view: activeView }}
		<div
			role="none"
			class="viz-sub_panel-content"
			style="height: calc((100vh - {mainHeaderHeight}px - {0.7 * 2}em - 8px) + 1px); width: 100%;"
			onclick={() => (subPanelContentFocused = true)}
			onkeydown={() => (subPanelContentFocused = true)}
			bind:this={subPanelContentElement}
			use:subPanelDrop={data}
		>
			{#await panelData}
				<LoadingContainer />
			{:then loadedData}
				{#if loadedData}
					<Comp data={loadedData.data} />
				{:else}
					<Comp />
				{/if}
			{:catch error}
				<h2>Something has gone wrong:</h2>
				<p style="color: red;">{error}</p>
			{/await}
		</div>
	{/if}
	{#if children}
		<div class="viz-sub_panel-content" style="white-space: nowrap;" data-pane-key={keyId}>
			{@render children()}
		</div>
	{/if}
</Pane>

<style lang="scss">
	.viz-sub_panel-header.locked {
		opacity: 0.7;
		pointer-events: auto; /* still accept clicks for context menu */
	}

	.viz-lock-indicator {
		position: absolute;
		right: 0.5em;
		top: 0.25em;
		background: transparent;
		border: none;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.25em;
		cursor: default;
	}
	#viz-debug-button {
		position: absolute;
		right: 0;
	}

	.viz-sub_panel-header {
		min-height: 1em;
		background-color: var(--imag-100);
		font-size: 13px;
		display: flex;
		align-items: center;
		position: relative;
	}

	.viz-sub_panel-content {
		text-overflow: clip;
		position: relative;
		display: flex;
		flex-direction: column;
		height: 100%;
		max-height: 100%;
	}

	.viz-tab-button {
		display: flex;
		align-items: center;
		position: relative;
		padding: 0.3em 0.7em;
		cursor: default;
		height: 100%;
		max-width: 11em;
		overflow: hidden;
		gap: 0.3em;
		font-size: 0.9em;

		&:hover {
			background-color: hsl(219, 26%, 15%);
		}
	}

	.viz-sub_panel-name {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	:global(
			.splitpanes__pane > *:last-child,
			.viz-sub_panel-content > :last-child:not(.splitpanes--horizontal, .splitpanes--vertical, .viz-view-container)
		) {
		padding: 0.5em;
	}

	:global(
			.splitpanes__pane
				> :is(.splitpanes, .splitpanes__pane, .viz-sub_panel-content, .splitpanes--horizontal, .splitpanes--vertical)
		) {
		padding: 0em;
	}

	.active-tab {
		box-shadow: 0 -1.5px 0 0 var(--imag-40) inset;
	}

	:global(.drop-hover-above) {
		outline: 1.5px solid var(--imag-outline-colour);
	}
</style>
