<script module lang="ts">
	import Splitpanes from "$lib/third-party/svelte-splitpanes/Splitpanes.svelte";

	export type InternalSubPanelContainer = Omit<
		VizSubPanel,
		"childs" | "children" | "$$events" | "$$slots" | "header" | "views"
	>;
	export type InternalPanelContainer = Omit<
		ComponentProps<typeof Splitpanes>,
		"children" | "$$events" | "$$slots"
	>;
	export type Content = Omit<VizSubPanel, "childs" | "id"> & {
		id?: string;
		views: VizView[];
	};
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
	import {
		onMount,
		setContext,
		untrack,
		type ComponentProps,
		type Snippet
	} from "svelte";
	import { Pane } from "$lib/third-party/svelte-splitpanes";
	import MaterialIcon from "../MaterialIcon.svelte";
	import { dev } from "$app/environment";
	import type { TabData } from "$lib/views/tabs.svelte";
	import TabOps from "$lib/views/tabs.svelte";
	import VizView from "$lib/views/views.svelte";
	import {
		measureComponentRenderTimes,
		resetAndReloadLayout
	} from "$lib/dev/components.svelte";
	import { views } from "$lib/layouts/views";
	import LoadingContainer from "../LoadingContainer.svelte";
	import { isElementScrollable } from "$lib/utils/dom";
	import { findSubPanel, generateKeyId } from "$lib/utils/layout";
	import { goto } from "$app/navigation";
	import ContextMenu, {
		type MenuItem
	} from "$lib/context-menu/ContextMenu.svelte";
	import {
		layoutState,
		layoutTree
	} from "$lib/third-party/svelte-splitpanes/state.svelte";
	import {
		cleanupEmptyPanels,
		duplicateView,
		normalizePanelSizes,
		removeEmptyContent,
		splitPanelHorizontally,
		splitPanelVertically
	} from "$lib/layouts/panel-operations";
	import {
		buildLayoutContextMenu,
		buildPanelContextMenu,
		buildTabContextMenu,
		type TabHandlers
	} from "./subpanel-context";
	import { debugMode } from "$lib/states/index.svelte";
	import SubPanelHeader from "./SubPanelHeader.svelte";
	import SubPanelContent from "./SubPanelContent.svelte";

	if (dev) {
		window.resetAndReloadLayout = resetAndReloadLayout;
	}

	if (debugMode) {
		measureComponentRenderTimes();
	}

	interface Props {
		id: string;
		header?: boolean;
		views: VizView[];
		children?: Snippet;
	}

	const defaultClass = "viz-panel";
	let className: string = $state(defaultClass);

	const allProps: Props & ComponentProps<typeof Pane> = $props();

	let id = allProps.id;
	let header = $state(allProps.header ?? false);

	const children = allProps.children;
	const keyId = allProps.paneKeyId ?? generateKeyId();
	const minSize = $state(allProps.minSize ?? 10);

	// pane size overrides when locked to prevent resizing
	let paneMinSize: number | undefined = $state(minSize);
	let paneMaxSize: number | undefined = $state(allProps.maxSize);

	// Helper: match svelte-kit style dynamic routes like "/collections/[uid]" to concrete paths
	function pathMatches(
		pattern: string | undefined,
		actual: string | undefined
	): boolean {
		if (!pattern || !actual) {
			return false;
		}

		if (pattern === actual) {
			return true;
		}

		// Escape regex specials, then turn dynamic segments \[param\] into [^/]+
		const escaped = pattern
			.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
			.replace(/\\\[[^\]]+\\\]/g, "[^/]+");
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
		const {
			id: panelViewId,
			name: panelViewName,
			path: panelViewPath
		} = storedView;

		// Try to find by path (supports dynamic segments), then by name, then by id
		const matchedView = views.find((view) => {
			if (panelViewPath && view.path && pathMatches(view.path, panelViewPath))
				return true;
			return view.name === panelViewName || view.id === panelViewId;
		});

		if (!matchedView?.component) {
			console.warn(
				`Could not find component for view: ${panelViewName} (id: ${panelViewId}, path: ${panelViewPath})`
			);
			continue;
		}

		// Reconstruct VizView instance from serialized data
		const v = VizView.fromJSON(storedView as any, matchedView.component);
		v.parent = keyId;

		initialViews[i] = v;
	}

	$effect(() => {
		if (allProps.class) {
			className = allProps.class;
		}

		if (initialViews.length > 0) {
			header = true;
		}

		if (header === true && initialViews.length === 0) {
			throw new Error(
				"Viz: Header is showing, but no tabs are provided for: " + keyId
			);
		}
	});

	const storedActiveView = $derived(
		panelViews.find((view) => view.isActive === true)
	);
	let activeView = $derived(storedActiveView ?? panelViews[0]);
	let panelData = $derived(
		activeView.viewData ?? activeView?.getComponentData()
	);

	function checkIfPanelLocked(): boolean {
		// Resolve the subpanel data from the global layout and respect its locked flag
		const result = findSubPanel("paneKeyId", keyId);
		if (!result) return false;
		const sp = result.subPanel as any;
		if (!sp) return false;
		return !!sp.locked;
	}

	let isPanelLocked = $derived(checkIfPanelLocked());

	// When a panel is locked, override min/max sizes to prevent resizing.
	$effect(() => {
		const result = findSubPanel("paneKeyId", keyId);
		const sp = result?.subPanel as any;
		if (sp && sp.locked) {
			// Try to use stored size, fallback to configured minSize
			const lockedSize =
				typeof sp.size === "number" && sp.size !== null ? sp.size : minSize;
			paneMinSize = lockedSize;
			paneMaxSize = lockedSize;
		} else {
			paneMinSize = minSize;
			paneMaxSize = allProps.maxSize;
		}
	});

	function togglePanelLock() {
		const result = findSubPanel("paneKeyId", keyId);
		if (!result) return;
		const { parentIndex, childIndex, isChild, subPanel } = result as any;
		if (!subPanel) return;
		// Toggle locked state on the underlying data model
		subPanel.locked = !subPanel.locked;
		// If this was a top-level panel, ensure the layoutState.tree entry is updated
		if (!isChild) {
			layoutState.tree[parentIndex].locked = subPanel.locked;
		} else {
			// If this is a child content entry, bubble up to the parent panel where appropriate
			(layoutState.tree[parentIndex].childs.content[childIndex] as any).locked =
				subPanel.locked;
		}
	}

	let subPanelContentElement: HTMLDivElement | undefined = $state();
	let subPanelContentFocused = $state(false);

	let showContextMenu = $state(false);
	let contextMenuItems = $state<MenuItem[]>([]);
	let contextMenuAnchor = $state<{ x: number; y: number } | null>(null);
	let contextMenuTargetView: VizView | null = $state(null);

	// Layout-level context menu (for locking the entire splitpanes)
	let showLayoutContextMenu = $state(false);
	let layoutContextMenuItems = $state<MenuItem[]>([]);
	let layoutContextMenuAnchor = $state<{ x: number; y: number } | null>(null);

	if (debugMode === true) {
		$inspect("active view", keyId, activeView);
		$inspect("panel views", keyId, panelViews);
	}

	let tabDropper = $state(new TabOps(initialViews));

	if (initialViews.length) {
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

	function headerDraggable(node: HTMLElement) {
		if (checkIfPanelLocked()) {
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
		if (checkIfPanelLocked() || view.locked) return;
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
				currentPanel.views = currentPanel.childs.content.flatMap(
					(c) => c.views
				);

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
		if (checkIfPanelLocked()) return;
		// Do not close locked tabs. Keep the exceptView and any locked tabs.
		panelViews = panelViews.filter((v) => v.id === exceptView.id || v.locked);

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
		if (checkIfPanelLocked() || view.locked) return;
		const index = panelViews.findIndex((v) => v.id === view.id);
		if (index === -1 || index === panelViews.length - 1) return;

		// Keep everything up to the specified index, but never remove locked tabs to the right.
		const viewsToRight = panelViews.slice(index + 1);
		const nonLockedRight = viewsToRight.filter((v) => !v.locked);
		const viewsToKeep = panelViews
			.slice(0, index + 1)
			.concat(viewsToRight.filter((v) => v.locked));
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
		if (checkIfPanelLocked()) return;
		// Close all tabs except those that are locked
		panelViews = panelViews.filter((v) => v.locked);

		const subPanel = findSubPanel("paneKeyId", keyId)?.subPanel;
		if (subPanel) {
			subPanel.views = panelViews;
		}
	}

	// Handlers object passed to the context menu builder so the menu logic
	// stays in a separate module and this file remains tidy.
	const menuHandlers: TabHandlers = {
		closeTab,
		closeOtherTabs,
		closeTabsToRight,
		splitRight,
		splitDown,
		moveToPanel,
		closeAllTabs,
		toggleTabLock: (v: VizView) => {
			v.locked = !v.locked;
			const sp = findSubPanel("paneKeyId", keyId)?.subPanel;
			if (sp) {
				const idx = sp.views.findIndex((x) => x.id === v.id);
				if (idx !== -1) sp.views[idx] = v;
			}
		}
	};

	/**
	 * Splits the current panel and moves a view to a new panel on the right
	 */
	function splitRight(view: VizView) {
		if (checkIfPanelLocked() || view.locked) return;
		const result = findSubPanel("paneKeyId", keyId);
		if (!result) return;

		const { parentIndex } = result;

		// Create a new view instance and split the panel
		const newView = duplicateView(view);
		const newPanel = splitPanelVertically(
			layoutState.tree,
			parentIndex,
			newView
		);

		if (newPanel) {
			newView.setActive(true);
		}
	}

	/**
	 * Splits the current panel and moves a view to a new content group below within the same parent
	 */
	function splitDown(view: VizView) {
		if (checkIfPanelLocked() || view.locked) return;
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
	function moveToPanel(view: VizView, direction: string) {
		if (checkIfPanelLocked() || view.locked) return;

		// Validate direction and cast to the narrower union for internal logic
		if (!["left", "right", "up", "down"].includes(direction)) {
			console.warn(`moveToPanel called with invalid direction: ${direction}`);
			return;
		}
		const dir = direction as "left" | "right" | "up" | "down";

		// Resolve the location of the view by its parent id to be precise
		const viewParentId = view.parent ?? keyId;
		const result = findSubPanel("paneKeyId", viewParentId);
		if (!result) return;

		if (debugMode) {
			const layoutSummary = layoutState.tree.map((p) => ({
				paneKeyId: p.paneKeyId,
				contentCount: p.childs?.content?.length ?? 0
			}));
			console.debug("[Viz] moveToPanel start", {
				viewId: view.id,
				viewParent: view.parent,
				result,
				layoutSummary,
				dir
			});
		}

		let { parentIndex, isChild, childIndex } = result;
		const currentPanel = layoutState.tree[parentIndex];

		// If findSubPanel returned the top-level panel (isChild === false), determine which content group contains this view
		if (!isChild) {
			const foundIdx =
				currentPanel.childs?.content?.findIndex((c) =>
					c.views.some((v) => v.id === view.id)
				) ?? -1;
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

		if (debugMode) {
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
			: (currentPanel.childs?.content?.findIndex((c) =>
					c.views.some((v) => v.id === view.id)
				) ?? -1);

		if (dir === "left") {
			if (resolvedChildIndex > 0) {
				targetPanelIndex = parentIndex;
				targetContentIndex = resolvedChildIndex - 1;
			} else if (parentIndex > 0) {
				// move into the previous panel's rightmost content group
				targetPanelIndex = parentIndex - 1;
				const prevPanel = layoutState.tree[targetPanelIndex];
				targetContentIndex = Math.max(
					0,
					(prevPanel.childs?.content?.length ?? 1) - 1
				);
			}
		} else if (dir === "right") {
			if (
				resolvedChildIndex !== -1 &&
				resolvedChildIndex < (currentPanel.childs?.content?.length ?? 0) - 1
			) {
				targetPanelIndex = parentIndex;
				targetContentIndex = resolvedChildIndex + 1;
			} else if (parentIndex < layoutState.tree.length - 1) {
				// move into the next panel's leftmost content group
				targetPanelIndex = parentIndex + 1;
				targetContentIndex = 0;
			}
		} else if (dir === "up" && parentIndex > 0) {
			targetPanelIndex = parentIndex - 1;
			targetContentIndex = 0;
		} else if (dir === "down" && parentIndex < layoutState.tree.length - 1) {
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

	function handleViewActive(view: VizView) {
		if (dev && activeView.id === view.id && view.path) {
			goto(view.path);
			return;
		}

		makeViewActive(view instanceof VizView ? view : new VizView(view));
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

<ContextMenu
	bind:showMenu={showContextMenu}
	items={contextMenuItems}
	anchor={contextMenuAnchor}
/>
<ContextMenu
	bind:showMenu={showLayoutContextMenu}
	items={layoutContextMenuItems}
	anchor={layoutContextMenuAnchor}
/>

<Pane
	class={className + (checkIfPanelLocked() ? " locked" : "")}
	{...allProps}
	{id}
	paneKeyId={keyId}
	minSize={paneMinSize}
	maxSize={paneMaxSize}
>
	<!--
TODO:
Make the header draggable too. Use the same drag functions. If we're dragging
a header into a different panel, place that panel in place and update the state
for Splitpanes
	-->
	{#if header}
		<SubPanelHeader
			{keyId}
			bind:panelViews
			bind:activeView
			bind:isPanelLocked
			{tabDropper}
			{dev}
			onViewActive={handleViewActive}
			onTogglePanelLock={togglePanelLock}
			{menuHandlers}
		/>
	{/if}
	{#if activeView?.component}
		{@const Comp = activeView.component}
		<SubPanelContent
			{keyId}
			{tabDropper}
			bind:panelData
			bind:panelViews
			bind:activeView
			bind:subPanelContentFocused
			componentToRender={Comp}
			onFocus={() => (subPanelContentFocused = true)}
		/>
	{/if}
	{#if children}
		<div
			class="viz-sub_panel-content"
			style="white-space: nowrap;"
			data-pane-key={keyId}
		>
			{@render children()}
		</div>
	{/if}
</Pane>

<style lang="scss">
	:global(.splitpanes__pane.locked) + :global(.splitpanes__resizer),
	:global(.splitpanes__resizer) + :global(.splitpanes__pane.locked) {
		pointer-events: none;
		opacity: 0.45;
		cursor: default;
	}

	:global(
			.splitpanes__pane > *:last-child,
			.viz-sub_panel-content
				> :last-child:not(
					.splitpanes--horizontal,
					.splitpanes--vertical,
					.viz-view-container
				)
		) {
		padding: 0.5em;
	}

	:global(
			.splitpanes__pane
				> :is(
					.splitpanes,
					.splitpanes__pane,
					.viz-sub_panel-content,
					.splitpanes--horizontal,
					.splitpanes--vertical
				)
		) {
		padding: 0em;
	}

	.viz-sub_panel-content {
		text-overflow: clip;
		position: relative;
		display: flex;
		flex-direction: column;
		height: 100%;
		max-height: 100%;
	}

	:global(.drop-hover-above) {
		outline: 1.5px solid var(--imag-outline-colour);
	}
</style>
