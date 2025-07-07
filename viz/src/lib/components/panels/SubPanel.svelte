<script module lang="ts">
	import { type Component } from "svelte";
	import Splitpanes from "$lib/third-party/svelte-splitpanes/Splitpanes.svelte";
	
	// TODO: Reorganise and clean up component
	// e.g. move types to seperate file, clean up props etc etc
	export interface VizView {
		name: string;
		opticalCenterFix?: number;
		component: Component;
		id: number;
		parent?: string;
		isActive?: boolean;
	}

	export type VizSubPanel = Props &
		ComponentProps<typeof Pane> & {
			childs?: {
				internalSubPanelContainer: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">;
				internalPanelContainer: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
				subPanel: Omit<VizSubPanel, "childs">[];
			};
		};
</script>

<script lang="ts">
	import type { ComponentProps, Snippet } from "svelte";
	import { Pane } from "$lib/third-party/svelte-splitpanes";
	import { generateKeyId, resetAndReloadLayout } from "$lib/utils";
	import MaterialIcon from "../MaterialIcon.svelte";
	import { views } from "$lib/layouts/test";
	import { dev } from "$app/environment";
	import type { TabData } from "$lib/TabDrop.svelte";
	import TabDropper from "$lib/TabDrop.svelte";

	if (dev) {
		window.resetAndReloadLayout = resetAndReloadLayout;
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
	let header = allProps.header ?? true;

	const children = allProps.children;
	const keyId = allProps.paneKeyId ?? generateKeyId();
	const minSize = allProps.minSize ?? 10;

	let panelViews = $state(allProps.views ?? []);

	// inject parent id into tabs
	for (const v of panelViews) {
		v.parent = keyId;
		v.component = views.find((view) => view.id === v.id)?.component!;
	}

	if (allProps.class) {
		className = allProps.class;
	}

	if (header === true && panelViews.length === 0) {
		throw new Error("Viz: Header is showing, but no tabs are provided for: " + keyId);
	}

	const storedActiveView = panelViews?.find((view) => view.isActive === true);
	let activeView = $state(storedActiveView ?? panelViews[0]);

	if (window.debug === true) {
		if (panelViews.length) {
			$inspect("active view", keyId, activeView);
		}

		$inspect("panel views", keyId, panelViews);
	}

	const tabDropper = new TabDropper(keyId, panelViews);

	$effect(() => {
		if (tabDropper.activeView) {
			activeView.isActive = false;
			tabDropper.activeView.isActive = true;
			activeView = tabDropper.activeView;
		}
	});

	function draggable(node: HTMLElement, data: TabData) {
		return tabDropper.draggable(node, data);
	}

	function onDropOver(event: DragEvent) {
		return tabDropper.onDropOver(event);
	}

	function tabDrop(node: HTMLElement) {
		return tabDropper.tabDrop(node);
	}
</script>

<Pane class={className} {minSize} {...allProps} {id} paneKeyId={keyId}>
	<!--
TODO:
Make the header draggable too. Use the same drag functions. If we're dragging
a header into a different panel, place that panel in place and update the state
for Splitpanes
	-->
	{#if header && panelViews.length > 0}
		<div class="viz-sub_panel-header" role="tablist" tabindex="0" use:tabDrop ondragover={(event) => onDropOver(event)}>
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
						onclick={() => {
							if (view.id === activeView.id) {
								return;
							}

							activeView.isActive = false;
							view.isActive = true;
							activeView = view;
						}}
						use:draggable={data}
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
							showHoverBG={true}
							style={`transform: translateY(${view.opticalCenterFix ?? 0.5}px);`}
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
		<div class="viz-sub_panel-content">
			<Comp />
		</div>
	{/if}
	{#if children}
		<div class="viz-sub_panel-content" data-pane-key={keyId}>
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
		background-color: var(--imag-blue-100);
		font-size: 13px;
		display: flex;
		align-items: center;
		position: relative;
	}

	.viz-sub_panel-content {
		height: 100%;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: clip;
	}

	.viz-tab-button {
		display: flex;
		align-items: center;
		position: relative;
		padding: 0.3em 0.7em;
		cursor: default;
		height: 100%;

		&:hover {
			background-color: hsl(219 31% 20% / 1);
		}
	}

	:global(
			.splitpanes__pane > *:last-child,
			.viz-sub_panel-content > :first-child:not(.splitpanes--horizontal, .splitpanes--vertical)
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
		box-shadow: 0 -1.5px 0 0 var(--imag-blue-40) inset;
	}

	:global(.drop-hover-above) {
		outline: 1.5px solid var(--imag-outline-colour);
	}
</style>
