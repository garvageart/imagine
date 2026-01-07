<script lang="ts">
	import type { TabGroup } from "$lib/layouts/model.svelte";
	import MaterialIcon from "../MaterialIcon.svelte";
	import LoadingContainer from "../LoadingContainer.svelte";
	import { setContext, untrack } from "svelte";
	import { tabOps } from "$lib/layouts/tab-ops.svelte";
	import { workspaceState } from "$lib/states/workspace.svelte";
	import { dev } from "$app/environment";
	import { resetAndReloadLayout } from "$lib/dev/components.svelte";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import type { MenuItem } from "$lib/context-menu/types";
	import {
		buildTabContextMenu,
		buildPanelContextMenu,
		buildLayoutContextMenu,
		type TabHandlers
	} from "./workspace-context";
	import type VizView from "$lib/views/views.svelte";
	import { goto } from "$app/navigation";
	import { DragData } from "$lib/drag-drop/data";
	import tippy, { type Instance } from "tippy.js";

	interface Props {
		group: TabGroup;
	}

	let { group }: Props = $props();

	setContext(
		"content",
		untrack(() => group)
	);

	let activeView = $derived(group.activeView);
	let Comp = $derived(activeView?.component);
	let isFocused = $derived(
		workspaceState.workspace?.activeGroupId === group.id
	);

	function handleFocus() {
		workspaceState.workspace?.setActiveGroup(group.id);
	}

	// Scrollbar and Dragging
	let headerEl: HTMLElement | undefined = $state();
	let scrollLeft = $state(0);
	let clientWidth = $state(0);
	let scrollWidth = $state(0);
	let isHoveringHeader = $state(false);
	let isDraggingScrollbar = $state(false);
	let dragScrollInterval: ReturnType<typeof setInterval> | null = null;
	const SCROLL_SPEED = 10;
	const SCROLL_THRESHOLD = 50;

	$effect(() => {
		if (!headerEl) {
			return;
		}

		const updateMetrics = () => {
			if (headerEl) {
				scrollLeft = headerEl.scrollLeft;
				clientWidth = headerEl.clientWidth;
				scrollWidth = headerEl.scrollWidth;
			}
		};

		updateMetrics();

		const resizeObserver = new ResizeObserver(updateMetrics);
		resizeObserver.observe(headerEl);

		headerEl.addEventListener("scroll", updateMetrics);
		return () => {
			resizeObserver.disconnect();
			headerEl?.removeEventListener("scroll", updateMetrics);
		};
	});

	function handleScrollbarDragStart(event: MouseEvent) {
		event.preventDefault();
		event.stopPropagation();
		isDraggingScrollbar = true;

		const startX = event.clientX;
		const startScrollLeft = scrollLeft;
		const thumbWidthRatio = clientWidth / scrollWidth;
		const thumbWidth = Math.max(20, clientWidth * thumbWidthRatio);
		const trackScrollableWidth = clientWidth - thumbWidth;
		const contentScrollableWidth = scrollWidth - clientWidth;
		const pxRatio = contentScrollableWidth / (trackScrollableWidth || 1);

		function onMouseMove(e: MouseEvent) {
			if (!headerEl) {
				return;
			}

			const deltaX = e.clientX - startX;
			headerEl.scrollLeft = startScrollLeft + deltaX * pxRatio;
		}

		function onMouseUp() {
			isDraggingScrollbar = false;
			window.removeEventListener("mousemove", onMouseMove);
			window.removeEventListener("mouseup", onMouseUp);
		}

		window.addEventListener("mousemove", onMouseMove);
		window.addEventListener("mouseup", onMouseUp);
	}

	function handleWheelScroll(event: WheelEvent) {
		if (headerEl) {
			if (event.deltaY !== 0 && Math.abs(event.deltaX) === 0) {
				event.preventDefault();
				headerEl.scrollLeft += event.deltaY;
			}
		}
	}

	function startDragScroll(direction: "left" | "right") {
		if (dragScrollInterval) return;
		dragScrollInterval = setInterval(() => {
			if (headerEl) {
				headerEl.scrollLeft +=
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
		if (!headerEl) {
			return;
		}

		const rect = headerEl.getBoundingClientRect();
		if (event.clientX < rect.left + SCROLL_THRESHOLD) {
			startDragScroll("left");
		} else if (event.clientX > rect.right - SCROLL_THRESHOLD) {
			startDragScroll("right");
		} else {
			stopDragScroll();
		}
	}

	// Keyboard Navigation
	function handleKeyDown(event: KeyboardEvent) {
		const views = group.views;
		const activeIndex = views.findIndex((v) => v.id === activeView?.id);
		if (activeIndex === -1) return;

		let nextIndex = -1;
		switch (event.key) {
			case "ArrowRight":
				nextIndex = (activeIndex + 1) % views.length;
				break;
			case "ArrowLeft":
				nextIndex = (activeIndex - 1 + views.length) % views.length;
				break;
			case "Home":
				nextIndex = 0;
				break;
			case "End":
				nextIndex = views.length - 1;
				break;
		}

		if (nextIndex !== -1 && views[nextIndex]) {
			event.preventDefault();
			group.setActive(views[nextIndex].id);
			const button = headerEl?.querySelector(
				`[role="tab"][aria-selected="true"]`
			) as HTMLElement;
			button?.focus();
		}
	}

	// Context Menus
	let tabCtxMenu = $state<{ show: boolean; items: MenuItem[]; anchor: any }>({
		show: false,
		items: [],
		anchor: null
	});
	let headerCtxMenu = $state<{ show: boolean; items: MenuItem[]; anchor: any }>(
		{ show: false, items: [], anchor: null }
	);

	const menuHandlers: TabHandlers = {
		closeTab: (v) => {
			group.removeTab(v.id);
			workspaceState.workspace?.cleanupNode(group);
		},
		closeOtherTabs: (v) => {
			group.views = group.views.filter(
				(view) => view.id === v.id || view.locked
			);
			workspaceState.workspace?.cleanupNode(group);
		},
		closeTabsToRight: (v) => {
			const index = group.views.findIndex((view) => view.id === v.id);
			group.views = group.views.filter((view, i) => i <= index || view.locked);
			workspaceState.workspace?.cleanupNode(group);
		},
		closeAllTabs: () => {
			group.views = group.views.filter((v) => v.locked);
			workspaceState.workspace?.cleanupNode(group);
		},
		closePanel: () => {
			group.views = [];
			workspaceState.workspace?.cleanupNode(group);
		},
		toggleTabLock: (v) => {
			v.locked = !v.locked;
		},
		splitRight: (v) => {
			workspaceState.workspace?.splitGroup(group.id, v, "right");
		},
		splitDown: (v) => {
			workspaceState.workspace?.splitGroup(group.id, v, "bottom");
		}
	};

	function triggerTabContextMenu(event: MouseEvent, view: VizView) {
		event.preventDefault();
		event.stopPropagation();

		tabCtxMenu.anchor = { x: event.clientX, y: event.clientY };
		tabCtxMenu.items = buildTabContextMenu(view, group, menuHandlers);
		tabCtxMenu.show = true;
	}

	function triggerHeaderContextMenu(event: MouseEvent) {
		event.preventDefault();

		headerCtxMenu.anchor = { x: event.clientX, y: event.clientY };
		headerCtxMenu.items = [
			...buildLayoutContextMenu(),
			...buildPanelContextMenu(group, menuHandlers)
		];
		headerCtxMenu.show = true;
	}

	// Tab Drag and Drop
	let dragTooltip: Instance | null = $state(null);

	function handleTabDragOver(e: DragEvent, view: VizView) {
		if (!e.dataTransfer) {
			return;
		}

		// Check if the view has a handler for any of the dragged types
		for (const type of e.dataTransfer.types) {
			const tabActions = view.getTabDropHandler(type);
			if (tabActions) {
				e.preventDefault();
				e.stopPropagation();
				e.dataTransfer.dropEffect = "copy";

				const target = e.currentTarget as HTMLElement;
				target.classList.add("drop-target-active");

				if (!dragTooltip) {
					dragTooltip = tippy(target, {
						content: tabActions.label,
						trigger: "manual",
						theme: "viz-theme",
						followCursor: "initial",
						animation: "shift-away",
						offset: [0, 0],
						delay: [200, 0],
						arrow: false
					});
					dragTooltip.show();
				}
				return;
			}
		}
	}

	function handleTabDragLeave(event: DragEvent) {
		const target = event.currentTarget as HTMLElement;
		target.classList.remove("drop-target-active");

		if (dragTooltip) {
			dragTooltip.destroy();
			dragTooltip = null;
		}
	}

	async function handleTabDrop(e: DragEvent, view: VizView) {
		if (!e.dataTransfer) {
			return;
		}

		const target = e.currentTarget as HTMLElement;
		target.classList.remove("drop-target-active");

		handleTabDragLeave(e);

		for (const type of e.dataTransfer.types) {
			const handler = view.getTabDropHandler(type);
			if (handler) {
				e.preventDefault();
				e.stopPropagation();

				const data = DragData.getData(e.dataTransfer, type);
				if (data) {
					await handler.dropHandler(data.payload, view);
				}
				return;
			}
		}
	}
</script>

<div
	class="tab-group-panel"
	use:tabOps.dropTarget={group.id}
	onclickcapture={handleFocus}
	role="none"
>
	{#if isFocused}
		<div class="viz-panel-active-overlay"></div>
	{/if}
	<div
		class="tab-group-header"
		role="toolbar"
		tabindex="-1"
		onmouseenter={() => (isHoveringHeader = true)}
		onmouseleave={() => (isHoveringHeader = false)}
		use:tabOps.addToGroup={group.id}
		oncontextmenu={triggerHeaderContextMenu}
	>
		<div
			bind:this={headerEl}
			class="tab-group-tabs-container"
			role="tablist"
			tabindex="0"
			onwheel={handleWheelScroll}
			onkeydown={handleKeyDown}
			ondragover={handleDragOver}
			ondragleave={stopDragScroll}
			ondrop={stopDragScroll}
		>
			{#each group.views as view}
				<button
					class="tab-button"
					class:active={group.activeViewId === view.id}
					role="tab"
					aria-selected={group.activeViewId === view.id}
					tabindex={group.activeViewId === view.id ? 0 : -1}
					onclick={() => {
						// TODO: make this configurable in user settings
						// first make it active and end it
						if (group.activeViewId !== view.id) {
							group.setActive(view.id);
							return;
						}

						// if the user clicks again and there's a path, go to it
						if (view.path && view.openPathFromTab) {
							goto(view.path);
						}
					}}
					oncontextmenu={(e) => triggerTabContextMenu(e, view)}
					use:tabOps.draggable={{ viewId: view.id, sourceGroupId: group.id }}
					ondragover={(e) => handleTabDragOver(e, view)}
					ondragleave={handleTabDragLeave}
					ondrop={(e) => handleTabDrop(e, view)}
				>
					<MaterialIcon
						style={`transform: translateY(${view.opticalCenterFix}px);`}
						weight={300}
						iconName="menu"
					/>
					<span class="tab-name">{view.name}</span>
					{#if view.locked}
						<MaterialIcon class="tab-lock" iconName="lock" />
					{/if}
				</button>
			{/each}
		</div>

		{#if scrollWidth > clientWidth}
			<div
				class="viz-custom-scrollbar {isHoveringHeader || isDraggingScrollbar
					? 'visible'
					: ''}"
				onmousedown={handleScrollbarDragStart}
				role="slider"
				tabindex="0"
				aria-valuemin="0"
				aria-valuemax="100"
				aria-valuenow={Math.round(
					(scrollLeft / (scrollWidth - clientWidth || 1)) * 100
				)}
			>
				<div
					class="viz-custom-scrollbar-thumb"
					style:width="{Math.max(
						20,
						(clientWidth / scrollWidth) * clientWidth
					)}px"
					style:transform="translateX({(scrollLeft /
						(scrollWidth - clientWidth || 1)) *
						(clientWidth -
							Math.max(20, (clientWidth / scrollWidth) * clientWidth))}px)"
				></div>
			</div>
		{/if}

		<div class="header-actions">
			{#if dev}
				<button
					class="header-action-button"
					aria-label="Reset and Reload"
					title="Reset and Reload"
					onclick={() => resetAndReloadLayout?.()}
				>
					<MaterialIcon iconName="refresh" />
				</button>
			{/if}
		</div>
	</div>

	<div class="tab-group-content">
		{#if activeView}
			{#if activeView.viewData}
				{#if Comp}
					<Comp data={activeView.viewData?.data} view={activeView} />
				{/if}
			{:else}
				{#await activeView.derivedViewData}
					<div class="loading-overlay">
						<LoadingContainer />
					</div>
				{:then loadedData}
					{#if Comp}
						{#if loadedData}
							<Comp data={loadedData?.data} view={activeView} />
						{:else}
							<Comp view={activeView} />
						{/if}
					{/if}
				{:catch error}
					<div class="error-container">
						<h3>Error loading data</h3>
						<p>{error.message}</p>
					</div>
				{/await}
			{/if}
		{:else}
			<div class="empty-group">
				<p>No active view</p>
			</div>
		{/if}
	</div>
</div>

<ContextMenu
	bind:showMenu={tabCtxMenu.show}
	items={tabCtxMenu.items}
	anchor={tabCtxMenu.anchor}
	offsetY={4}
/>
<ContextMenu
	bind:showMenu={headerCtxMenu.show}
	items={headerCtxMenu.items}
	anchor={headerCtxMenu.anchor}
	offsetY={4}
/>

<style lang="scss">
	.tab-group-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
		overflow: hidden;
		position: relative;
	}

	.viz-panel-active-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 900;
		box-shadow:
			0 1.5px 0 var(--imag-primary) inset,
			1.5px 0 0 var(--imag-primary) inset,
			-1.5px 0 0 var(--imag-primary) inset,
			0 -1.5px 0 var(--imag-primary) inset;
	}

	.tab-group-header {
		background-color: var(--imag-100);
		font-size: 0.8rem;
		display: flex;
		align-items: center;
		position: relative;
		overflow: hidden;
		width: 100%;
		min-width: 0;
		flex-shrink: 0;
		height: 1.8em;
		transition: background-color 0.2s;

		&:global(.drop-active) {
			background-color: var(--imag-90);
		}
	}

	.tab-group-tabs-container {
		display: flex;
		align-items: center;
		flex: 1;
		min-width: 0;
		overflow-x: auto;
		overflow-y: hidden;
		height: 100%;
		white-space: nowrap;
		scrollbar-width: none;
		-ms-overflow-style: none;
		&::-webkit-scrollbar {
			display: none;
		}
	}

	.tab-button {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		position: relative;
		padding: 0.2em 0.5em;
		cursor: default;
		height: 100%;
		max-width: 11em;
		overflow: hidden;
		gap: 0.3em;
		font-size: 0.85em;
		font-weight: 450;
		border: none;
		background: transparent;
		color: inherit;
		transition: background-color 0.15s ease;

		&:hover {
			background-color: var(--imag-90);
		}

		&.active {
			box-shadow: 0 -2px 0 0 var(--imag-primary) inset;
		}

		&:global(.drop-target-active) {
			background-color: color-mix(
				in srgb,
				var(--imag-primary) 30%,
				transparent
			) !important;
			outline: 1.5px solid var(--imag-primary);
			outline-offset: -1.5px;
			z-index: 10;
		}
	}

	.tab-name {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.tab-group-content {
		flex: 1;
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		background-color: var(--imag-bg-color);
	}

	.viz-custom-scrollbar {
		position: absolute;
		bottom: 0;
		left: 0;
		height: 3px;
		width: 100%;
		z-index: 10;
		opacity: 0;
		pointer-events: none;
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

	.header-actions {
		display: flex;
		align-items: center;
		height: 100%;
		flex-shrink: 0;
		background-color: var(--imag-100);
		z-index: 2;
	}

	.header-action-button {
		background: transparent;
		border: none;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0 0.5em;
		cursor: pointer;
		height: 100%;
	}

	.loading-overlay,
	.empty-group,
	.error-container {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		padding: 2rem;
		text-align: center;
		color: var(--imag-60);
	}

	.error-container {
		color: var(--imag-error);
	}

	:global(.tab-lock) {
		font-size: 0.9em;
		opacity: 0.7;
		margin-left: 0.25em;
	}
</style>
