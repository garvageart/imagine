<script module lang="ts">
	import { type Component } from "svelte";
	import Splitpanes from "$lib/third-party/svelte-splitpanes/Splitpanes.svelte";
	// TODO: Reorganise and clean up component
	// e.g. move types to seperate file, clean up props etc etc

	export interface VizTab {
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
				parentSubPanel: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "tabs">;
				parentPanel: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
				subPanel: Omit<VizSubPanel, "childs">[];
			};
		};
</script>

<script lang="ts">
	import type { ComponentProps, Snippet } from "svelte";
	import { Pane } from "$lib/third-party/svelte-splitpanes";
	import { generateKeyId, swapArrayElements } from "$lib/utils";
	import MaterialIcon from "../MaterialIcon.svelte";
	import { allTabs, layoutState } from "$lib/third-party/svelte-splitpanes/state";
	import { views } from "$lib/layouts/test";

	interface Props {
		id: string;
		header?: boolean;
		tabs?: VizTab[];
		children?: Snippet;
	}

	interface TabData {
		index: number;
		data: VizTab;
	}

	const defaultClass = "viz-panel";
	let className: string = $state(defaultClass);

	const allProps: Props & ComponentProps<typeof Pane> = $props();

	let panelTabs = $state(allProps.tabs ?? []);
	let id = allProps.id;
	let header = allProps.header ?? true;

	const children = allProps.children;
	const keyId = allProps.paneKeyId ?? generateKeyId();
	const minSize = allProps.minSize ?? 10;

	// inject parent id into tabs
	for (const tab of panelTabs) {
		tab.parent = keyId;
		tab.component = views.find((view) => view.id === tab.id)?.component!;
	}

	if (allProps.class) {
		className = allProps.class;
	}

	if (header === true && panelTabs?.length === 0) {
		throw Error("Viz: Header is showing, but no tabs are provided");
	}

	const storedActiveTab = $allTabs.get(keyId)?.find((tab) => tab.isActive === true);
	let activeTab = $state.raw(storedActiveTab ?? panelTabs[0]);

	if (panelTabs.length > 0) {
		$allTabs.set(keyId, panelTabs);
		panelTabs[0].isActive = true;
	} else {
		$allTabs.delete(keyId);
	}

	if (window.debug === true) {
		$inspect("panel tabs " + keyId, panelTabs);
		$inspect("all tabs " + keyId, $allTabs);
	}
	function draggable(node: HTMLElement, data: TabData) {
		let state = JSON.stringify(data);

		node.draggable = true;

		node.addEventListener("dragstart", (e) => {
			e.dataTransfer?.setData("text/json", state);
		});

		return {
			update(data: TabData) {
				state = JSON.stringify(data);
			},
			destroy() {
				node.removeEventListener("dragstart", (e) => {
					e.dataTransfer?.setData("text/json", state);
				});
			}
		};
	}

	function onDropOver(event: DragEvent) {
		event.preventDefault();
		if (event.dataTransfer) {
			event.dataTransfer.dropEffect = "move";
		}
	}

	function ondrop(node: HTMLElement, event: DragEvent) {
		event.preventDefault();

		if (event.dataTransfer) {
			const data = event.dataTransfer.getData("text/json");
			const state = JSON.parse(data) as TabData;
			const tabKeyId = node.getAttribute("data-tab-id");
			const nodeParentId = node.parentElement?.getAttribute("data-viz-sp-id");

			if (state.data.id === parseInt(tabKeyId!)) {
				return;
			}

			if (panelTabs.map((tab) => tab.id).includes(state.data.id) === false) {
				// Remove it from its original pane first
				const splicedTabs = $allTabs.get(state.data.parent!)?.splice(state.index, 1);

				if (!splicedTabs || splicedTabs.length === 0) {
					return;
				}

				state.data.component = splicedTabs[0].component;
				state.data.isActive = true;

				panelTabs.push(...splicedTabs);
				$allTabs.set(keyId, [...panelTabs]);

				if ($allTabs.get(state.data.parent!)?.length === 0) {
					$allTabs.delete(state.data.parent!);

					const currentLayout = $layoutState;
					let indx = currentLayout.findIndex((panel) => panel.paneKeyId === state.data.parent);
					let spliced: VizSubPanel[] = [];
					let isChild = false;

					if (indx === -1) {
						isChild = true;
						indx = currentLayout[0].childs?.subPanel.findIndex((panel) => panel.paneKeyId === state.data.parent) ?? -1;
					}

					if (indx !== -1 && isChild) {
						spliced = currentLayout[0].childs?.subPanel.splice(indx, 1) ?? [];
					} else {
						spliced = currentLayout.splice(indx, 1);
					}

					const splicedTabs = spliced.map((panel) => panel.tabs ?? []).flat();
					const subPanelToUpdate = currentLayout.find((panel) => panel.paneKeyId === keyId);

					subPanelToUpdate?.tabs?.push(...splicedTabs);
					$layoutState = [...currentLayout];
				}

				state.data.parent = nodeParentId!;
				state.index = panelTabs.length - 1;
				activeTab = state.data;

				return;
			}

			const originalTab = panelTabs.find((tab) => tab.id === state.data.id);
			if (!originalTab) {
				return;
			}

			// JSON.stringify got rid of the component,
			// so just get it from the original tab
			state.data.component = originalTab.component;

			// If element is dropped on the header, just move the element
			// to the end of the array since there are no other elements
			// in the header
			const tabIndex = panelTabs.findIndex((tab) => tab.id === state.data.id);

			if (tabIndex === panelTabs.length - 1) {
				activeTab = state.data;
				return;
			}

			if (node.classList.contains("viz-sub_panel-header") && tabIndex === state.index) {
				panelTabs.push(state.data);
				// index shifts up one when added towards the end
				// so just track back
				//
				// btw: this is so wank but it works
				if (state.index === 0) {
					panelTabs.splice(state.index, 1);
				} else {
					panelTabs.splice(state.index - 1, 1);
				}
			} else if (tabIndex === state.index) {
				// Swap it if it's dropped on a tab
				swapArrayElements(
					panelTabs,
					state.index,
					panelTabs.findIndex((tab) => tab.id === parseInt(node.getAttribute("data-tab-id")!))
				);
			}

			activeTab = state.data;
		}
	}

	function tabDrop(node: HTMLElement) {
		node.addEventListener("drop", (e) => {
			node.classList.remove("drop-hover-above");
			ondrop(node, e);
		});

		node.addEventListener("dragenter", (e) => {
			e.preventDefault();
			if (node === e.target) {
				return;
			}

			node.classList.add("drop-hover-above");
		});

		node.addEventListener("dragleave", (e) => {
			const target = e.target as HTMLElement;
			if (node === target) {
				return;
			}

			node.classList.remove("drop-hover-above");
		});

		node.addEventListener("dragend", (e) => {
			node.classList.remove("drop-hover-above");
		});

		return {
			destroy() {
				node.removeEventListener("drop", (e) => {
					ondrop(node, e);
				});

				node.removeEventListener("dragenter", (e) => {
					e.preventDefault();
					if (node === e.target) {
						return;
					}
				});

				node.removeEventListener("dragend", (e) => {
					node.classList.remove("drop-hover-above");
				});
			}
		};
	}
</script>

<Pane class={className} {minSize} {...allProps} {id} paneKeyId={keyId}>
	{#if header && panelTabs?.length > 0}
		<!--
	TODO:
	Make the header draggable too. Use the same drag functions. If we're dragging
	a header into a different panel, place that panel in place and update the state
	for Splitpanes
		-->
		<div class="viz-sub_panel-header" role="tablist" tabindex="0" use:tabDrop ondragover={(event) => onDropOver(event)}>
			{#each panelTabs as tab, i}
				{#if tab.name.trim() != ""}
					{@const tabNameId = tab.name.toLowerCase().replaceAll(" ", "-")}
					{@const data = { index: i, data: tab }}
					<button
						id={tabNameId + "-tab"}
						class="viz-tab-button {activeTab.id === tab.id ? 'active-tab' : ''}"
						data-tab-id={tab.id}
						role="tab"
						onclick={() => {
							activeTab.isActive = false;
							tab.isActive = true;
							activeTab = tab;
						}}
						use:draggable={data}
						use:tabDrop
						ondragover={(event) => onDropOver(event)}
					>
						<span class="viz-sub_panel-name">{tab.name}</span>
						<!--
						Every tab name needs to manually align itself with the icon
						Translate is used instead of margin or position is used to avoid
						shifting the layout  
						-->
						<MaterialIcon showHoverBG={true} style={`transform: translateY(${tab.opticalCenterFix ?? 0.5}px);`} iconName="menu" />
					</button>
				{/if}
			{/each}

			<button
				id="viz-debug-button"
				class="viz-tab-button"
				aria-label="Reset and Reload"
				title="Reset and Reload"
				onclick={() => {
					localStorage.removeItem("viz:layout");
					location.reload();
				}}
			>
				<span class="viz-sub_panel-name">Reset Layout</span>
				<MaterialIcon iconName="refresh" />
			</button>
		</div>
	{/if}
	{#if activeTab?.component}
		{@const Comp = activeTab.component}
		<div class="viz-sub_panel-content">
			<Comp />
		</div>
	{/if}
	{#if children}
		<div class="viz-sub_panel-content" data-pane-key={keyId}>
			{@render children?.()}
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
