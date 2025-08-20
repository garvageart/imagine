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
	import { findSubPanel, generateKeyId } from "$lib/utils/layout";
	import { goto } from "$app/navigation";

	if (dev) {
		window.resetAndReloadLayout = resetAndReloadLayout;
	}

	if (window.debug) {
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

	let panelViews = $state(allProps.views ?? []);

	// construct the views from the stored data
	for (let i = 0; i < panelViews.length; i++) {
		const panelViewId = panelViews[i].id;
		const v = new VizView({
			...panelViews[i],
			parent: keyId,
			component: views.find((view) => view.id === panelViewId)?.component!
		});

		v.isActive = panelViews[i].isActive;
		panelViews[i] = v;
	}

	if (allProps.class) {
		className = allProps.class;
	}

	if (panelViews.length > 0) {
		header = true;
	}

	if (header === true && panelViews.length === 0) {
		throw new Error("Viz: Header is showing, but no tabs are provided for: " + keyId);
	}

	const storedActiveView = $derived(panelViews.find((view) => view.isActive === true));
	let activeView = $derived(storedActiveView ?? panelViews[0]);
	let panelData = $derived(activeView?.getComponentData());

	let subPanelContentElement: HTMLDivElement | undefined = $state();
	let subPanelContentFocused = $state(false);

	if (window.debug === true) {
		$inspect("active view", keyId, activeView);
		if (panelViews.length) {
			// lmaooooo???? fucked up language. avoiding the state_referenced_locally error
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

	if (panelViews.length) {
		tabDropper = new TabOps(panelViews);
		setContext<Content>("content", {
			paneKeyId: keyId,
			views: panelViews
		});
		$effect(() => {
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
		});
	}

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
		return tabDropper.draggable(node, data);
	}

	function onDropOver(event: DragEvent) {
		return tabDropper.onDropOver(event);
	}

	function tabDrop(node: HTMLElement) {
		return tabDropper.tabDrop(node);
	}

	function headerDraggable(node: HTMLElement) {}

	function subPanelDrop(node: HTMLElement, data: TabData) {
		return tabDropper.subPanelDropInside(node, data);
	}

	function makeViewActive(view: VizView) {
		if (view.id === activeView.id) {
			return;
		}

		view.isActive = true;
		activeView.isActive = false;
		activeView = view;

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

<Pane class={className} {minSize} {...allProps} {id} paneKeyId={keyId}>
	<!--
TODO:
Make the header draggable too. Use the same drag functions. If we're dragging
a header into a different panel, place that panel in place and update the state
for Splitpanes
	-->
	{#if panelViews.length > 0}
		<div
			class="viz-sub_panel-header"
			role="tablist"
			tabindex="0"
			use:headerDraggable
			use:tabDrop
			ondragover={(event) => onDropOver(event)}
		>
			{#each panelViews as view, i}
				{#if view.name.trim() != ""}
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

							makeViewActive(view instanceof VizView ? view : new VizView(view));
						}}
						use:tabDragable={data}
						use:tabDrop
						ondragover={(event) => onDropOver(event)}
					>
						<span class="viz-sub_panel-name">{view.name}</span>
						<!--
						Every tab name needs to manually align itself with the icon
						Translate is used instead of margin or position is used to avoid
						shifting the layout  
						-->
						<MaterialIcon
							style={`transform: translateY(${view.opticalCenterFix ?? 0.5}px); margin-left: 0.5em;`}
							iconName="menu"
						/>
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
			{:then panelData}
				{#if panelData}
					<Comp data={panelData.data} />
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

		&:hover {
			background-color: hsl(219, 26%, 15%);
		}
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
