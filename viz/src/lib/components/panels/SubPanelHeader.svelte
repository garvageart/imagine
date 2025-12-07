<script lang="ts">
	import MaterialIcon from "../MaterialIcon.svelte";
	import type { TabData } from "$lib/views/tabs.svelte.ts";
	import type TabOps from "$lib/views/tabs.svelte.ts";
	import VizView from "$lib/views/views.svelte";
	import { resetAndReloadLayout } from "$lib/dev/components.svelte";
	import ContextMenu, {
		type MenuItem
	} from "$lib/context-menu/ContextMenu.svelte";
	import {
		buildLayoutContextMenu,
		buildPanelContextMenu,
		buildTabContextMenu
	} from "./subpanel-context";
	import type { TabHandlers } from "./subpanel-context";

	interface Props {
		keyId: string;
		panelViews: VizView[];
		activeView: VizView;
		isPanelLocked: boolean;
		tabDropper: TabOps | undefined;
		dev: boolean;
		onViewActive: (view: VizView) => void;
		onTogglePanelLock: () => void;
		menuHandlers: TabHandlers;
	}

	let {
		keyId,
		panelViews,
		activeView,
		isPanelLocked,
		tabDropper,
		dev: isDevMode,
		onViewActive,
		onTogglePanelLock,
		menuHandlers
	}: Props = $props();

	let headerElement: HTMLElement | undefined = $state();
	let contextMenuItems = $state<MenuItem[]>([]);
	let contextMenuAnchor = $state<{ x: number; y: number } | null>(null);
	let showContextMenu = $state(false);
	let showLayoutContextMenu = $state(false);
	let layoutContextMenuItems = $state<MenuItem[]>([]);
	let layoutContextMenuAnchor = $state<{ x: number; y: number } | null>(null);

	// Custom Scrollbar State
	let scrollLeft = $state(0);
	let clientWidth = $state(0);
	let scrollWidth = $state(0);
	let isHoveringHeader = $state(false);
	let isDraggingScrollbar = $state(false);

	$effect(() => {
		if (!headerElement) return;

		const updateMetrics = () => {
			if (headerElement) {
				scrollLeft = headerElement.scrollLeft;
				clientWidth = headerElement.clientWidth;
				scrollWidth = headerElement.scrollWidth;
			}
		};

		// Initial update
		updateMetrics();

		// Observer for resize
		const resizeObserver = new ResizeObserver(updateMetrics);
		resizeObserver.observe(headerElement);

		// Listen for scroll events to update state
		headerElement.addEventListener("scroll", updateMetrics);

		return () => {
			resizeObserver.disconnect();
			headerElement?.removeEventListener("scroll", updateMetrics);
		};
	});

	function handleScrollbarDragStart(event: MouseEvent) {
		event.preventDefault();
		event.stopPropagation();
		isDraggingScrollbar = true;

		const startX = event.clientX;
		const startScrollLeft = scrollLeft;

		// Calculate ratio: how much scrollable content moves per pixel of scrollbar movement
		// thumb travel space = clientWidth - thumbWidth
		// content travel space = scrollWidth - clientWidth
		const thumbWidthRatio = clientWidth / scrollWidth;
		const thumbWidth = Math.max(20, clientWidth * thumbWidthRatio);
		const trackScrollableWidth = clientWidth - thumbWidth;
		const contentScrollableWidth = scrollWidth - clientWidth;

		const pxRatio = contentScrollableWidth / (trackScrollableWidth || 1);

		function onMouseMove(e: MouseEvent) {
			if (!headerElement) return;
			const deltaX = e.clientX - startX;
			headerElement.scrollLeft = startScrollLeft + deltaX * pxRatio;
		}

		function onMouseUp() {
			isDraggingScrollbar = false;
			window.removeEventListener("mousemove", onMouseMove);
			window.removeEventListener("mouseup", onMouseUp);
		}

		window.addEventListener("mousemove", onMouseMove);
		window.addEventListener("mouseup", onMouseUp);
	}

	// Action wrappers to handle undefined tabDropper safely
	function safeTabDrop(node: HTMLElement) {
		if (tabDropper) {
			return tabDropper.tabDrop(node);
		}
		return { destroy: () => {} };
	}

	function safeTabDraggable(node: HTMLElement, data: TabData) {
		if (isPanelLocked || (data && data.view && data.view.locked)) {
			return { destroy: () => {} };
		}
		if (tabDropper) {
			return tabDropper.draggable(node, data);
		}
		return { destroy: () => {} };
	}

	function handleWheelScroll(event: WheelEvent) {
		if (headerElement) {
			// Map vertical scroll (mouse wheel) to horizontal scroll
			// But allow native horizontal scroll (trackpad) to pass through
			if (event.deltaY !== 0 && Math.abs(event.deltaX) === 0) {
				event.preventDefault();
				headerElement.scrollLeft += event.deltaY;
			}
		}
	}

	let dragScrollInterval: ReturnType<typeof setInterval> | null = null;
	const SCROLL_SPEED = 10;
	const SCROLL_THRESHOLD = 50;

	function startDragScroll(direction: "left" | "right") {
		if (dragScrollInterval) return;
		dragScrollInterval = setInterval(() => {
			if (headerElement) {
				headerElement.scrollLeft +=
					direction === "right" ? SCROLL_SPEED : -SCROLL_SPEED;
			}
		}, 50);
	}

	function stopDragScroll() {
		if (dragScrollInterval) {
			clearInterval(dragScrollInterval);
			dragScrollInterval = null;
		}
	}

	function handleDragOver(event: DragEvent) {
		if (tabDropper) {
			tabDropper.onDropOver(event);
		}

		if (!headerElement) return;

		const rect = headerElement.getBoundingClientRect();
		const clientX = event.clientX;

		if (clientX < rect.left + SCROLL_THRESHOLD) {
			startDragScroll("left");
		} else if (clientX > rect.right - SCROLL_THRESHOLD) {
			startDragScroll("right");
		} else {
			stopDragScroll();
		}
	}

	function handleDragLeave() {
		stopDragScroll();
	}

	function handleDrop(event: DragEvent) {
		stopDragScroll();
	}

	function handleKeyDown(event: KeyboardEvent) {
		const tabs = Array.from(
			headerElement?.querySelectorAll('[role="tab"]') || []
		) as HTMLElement[];
		const activeTabIndex = panelViews.findIndex(
			(view) => view.id === activeView?.id
		);
		let nextTabIndex = -1;

		switch (event.key) {
			case "ArrowRight":
				event.preventDefault();
				nextTabIndex = (activeTabIndex + 1) % panelViews.length;
				break;
			case "ArrowLeft":
				event.preventDefault();
				nextTabIndex =
					(activeTabIndex - 1 + panelViews.length) % panelViews.length;
				break;
			case "Home":
				event.preventDefault();
				nextTabIndex = 0;
				break;
			case "End":
				event.preventDefault();
				nextTabIndex = panelViews.length - 1;
				break;
		}

		if (nextTabIndex !== -1 && panelViews[nextTabIndex]) {
			onViewActive(panelViews[nextTabIndex]);
			// Use a timeout to ensure the tab is rendered and focusable after reactivity updates
			setTimeout(() => {
				tabs[nextTabIndex]?.focus();
			}, 0);
		}
	}

	function triggerTabContextMenu(event: MouseEvent, view: VizView) {
		contextMenuAnchor = { x: event.clientX, y: event.clientY };
		contextMenuItems = buildTabContextMenu(
			view,
			panelViews,
			keyId,
			menuHandlers
		);
		showContextMenu = true;
	}

	function triggerHeaderContextMenu(event: MouseEvent) {
		layoutContextMenuAnchor = { x: event.clientX, y: event.clientY };
		layoutContextMenuItems = [
			...buildLayoutContextMenu(),
			...buildPanelContextMenu(keyId, panelViews)
		];
		showLayoutContextMenu = true;
	}
</script>

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

<!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
<div
	class="viz-sub_panel-header {isPanelLocked ? 'locked' : ''}"
	oncontextmenu={(e) => triggerHeaderContextMenu(e)}
	role="toolbar"
	aria-label="Panel Header"
	tabindex="-1"
	onmouseenter={() => (isHoveringHeader = true)}
	onmouseleave={() => (isHoveringHeader = false)}
>
	<!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
	<div
		bind:this={headerElement}
		class="viz-sub_panel-tabs"
		role="tablist"
		tabindex="0"
		use:safeTabDrop
		onwheel={handleWheelScroll}
		onkeydown={handleKeyDown}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
		ondrop={handleDrop}
	>
		{#each panelViews as view, i}
			{#if view.name && view.name.trim() !== ""}
				{@const tabNameId = view.name.toLowerCase().replaceAll(" ", "-")}
				{@const data = { index: i, view: view }}
				<button
					id={tabNameId + "-tab"}
					class="viz-tab-button {activeView?.id === view.id
						? 'active-tab'
						: ''}"
					data-tab-id={view.id}
					role="tab"
					title={view.name}
					aria-label={view.name}
					onclick={() => onViewActive(view)}
					oncontextmenu={(e) => triggerTabContextMenu(e, view)}
					use:safeTabDraggable={data}
					aria-selected={activeView?.id === view.id ? "true" : "false"}
					tabindex={activeView?.id === view.id ? 0 : -1}
				>
					<MaterialIcon
						style={`transform: translateY(${view.opticalCenterFix}px);`}
						weight={200}
						iconName="menu"
					/>
					<span class="viz-sub_panel-name">{view.name}</span>
					{#if view.locked}
						<MaterialIcon class="viz-tab-lock" iconName="lock" />
					{/if}
				</button>
			{/if}
		{/each}
	</div>

	<!-- Custom Overlay Scrollbar -->
	{#if scrollWidth > clientWidth}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="viz-custom-scrollbar {isHoveringHeader || isDraggingScrollbar
				? 'visible'
				: ''}"
			onmousedown={handleScrollbarDragStart}
		>
			<div
				class="viz-custom-scrollbar-thumb"
				style:width="{Math.max(
					20,
					(clientWidth / scrollWidth) * clientWidth
				)}px"
				style:transform="translateX({(scrollLeft /
					(scrollWidth - clientWidth)) *
					(clientWidth -
						Math.max(20, (clientWidth / scrollWidth) * clientWidth))}px)"
			></div>
		</div>
	{/if}

	<div class="header-actions">
		{#if isDevMode}
			<button
				id="viz-debug-button"
				class="viz-tab-button icon-only"
				aria-label="Reset and Reload"
				title="Reset and Reload"
				onclick={() => resetAndReloadLayout()}
			>
				<MaterialIcon iconName="refresh" />
			</button>
		{/if}
		{#if isPanelLocked}
			<button
				class="viz-lock-indicator"
				aria-label="Toggle panel lock"
				title={isPanelLocked ? "Unlock Panel" : "Lock Panel"}
				onclick={() => onTogglePanelLock()}
			>
				<MaterialIcon iconName={isPanelLocked ? "lock" : "lock_open"} />
			</button>
		{/if}
	</div>
</div>

<style lang="scss">
	.viz-sub_panel-header {
		background-color: var(--imag-100);
		font-size: 0.8rem;
		display: flex;
		align-items: center;
		position: relative;
		overflow: hidden; /* Ensure no scrollbars on the main header */
		width: 100%; /* Take full width of its parent */
		min-width: 0; /* Critical: allows header to shrink inside flex parent instead of forcing overflow */
	}

	.viz-sub_panel-tabs {
		display: flex;
		align-items: center;
		flex: 1; /* Allow tabs to take up available space */
		min-width: 0; /* Critical for flex item scrolling: allows shrinking below content size */
		overflow-x: auto; /* Enable horizontal scrolling */
		overflow-y: hidden; /* Hide vertical overflow */
		height: 100%; /* Take full height of the header */
		white-space: nowrap; /* Prevent tabs from wrapping */

		/* hide native scrollbars */
		scrollbar-width: none;
		-ms-overflow-style: none;

		&::-webkit-scrollbar {
			display: none;
		}
	}

	/* Custom Scrollbar Styles */
	.viz-custom-scrollbar {
		position: absolute;
		bottom: 0;
		left: 0;
		height: 3px;
		width: 100%;
		z-index: 10;
		opacity: 0;
		pointer-events: none; /* Let clicks pass through when hidden */
		transition: opacity 0.2s;
	}

	.viz-custom-scrollbar.visible {
		opacity: 1;
		pointer-events: auto;
	}

	.viz-custom-scrollbar-thumb {
		background-color: color-mix(in srgb, var(--imag-40) 50%, transparent);
		height: 100%;
		cursor: pointer;
		position: absolute;
		top: 0;
		left: 0;
	}

	.viz-sub_panel-header.locked {
		opacity: 0.7;
		pointer-events: auto; /* Re-enable pointer events for locked state if needed */
	}

	.viz-tab-button {
		flex-shrink: 0; /* Prevent tabs from shrinking */
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
		border: none;
		background: transparent;
		color: inherit;

		&:hover {
			background-color: var(--imag-90);
		}
	}

	.viz-sub_panel-name {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.active-tab {
		box-shadow: 0 -2px 0 0 var(--imag-primary) inset;
	}

	.header-actions {
		display: flex;
		align-items: center;
		height: 100%;
		flex-shrink: 0; /* Prevent actions from shrinking */
		background-color: var(--imag-100);
		box-shadow: -5px 0 5px -5px rgba(0, 0, 0, 0.2); /* Shadow to indicate separation */
		z-index: 5; /* Ensure actions are above tabs if there's overlap */
	}

	.viz-lock-indicator,
	#viz-debug-button {
		background: transparent;
		border: none;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0 0.5em;
		cursor: pointer;
		height: 100%;
	}

	:global(.viz-tab-lock) {
		width: 1em;
		height: 1em;
		opacity: 0.9;
		margin-left: 0.25em;
		font-size: 0.9em;
	}
</style>
