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
	import { generateKeyId, resetAndReloadLayout, swapArrayElements } from "$lib/utils";
	import MaterialIcon from "../MaterialIcon.svelte";
	import { allTabs, layoutState } from "$lib/third-party/svelte-splitpanes/state";
	import { views } from "$lib/layouts/test";
	import { dev } from "$app/environment";

	if (dev) {
		window.resetAndReloadLayout = resetAndReloadLayout;
	}

	interface Props {
		id: string;
		header?: boolean;
		tabs: VizTab[];
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

	if (header === true && panelTabs.length === 0) {
		throw Error("Viz: Header is showing, but no tabs are provided for: " + keyId);
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

	/**
	 * Promotes the first child subpanel to the parent panel, or removes the parent if no children.
	 * Mutates currentLayout in place.
	 */
	function promoteChildToParent(currentLayout: VizSubPanel[], parentIndex: number) {
		const parentPanel = currentLayout[parentIndex];
		if (parentPanel.childs?.subPanel?.length) {
			const firstChild = parentPanel.childs.subPanel[0];
			Object.assign(parentPanel, {
				id: firstChild.id,
				maxSize: firstChild.maxSize,
				minSize: firstChild.minSize,
				paneKeyId: firstChild.paneKeyId,
				tabs: firstChild.tabs,
				childs: {
					...parentPanel.childs,
					subPanel: parentPanel.childs.subPanel.slice(1)
				}
			});
		} else {
			currentLayout.splice(parentIndex, 1);
		}
	}

	function findPanelIndex(layout: VizSubPanel[], paneKeyId: string | undefined) {
		return layout.findIndex((panel) => panel.paneKeyId === paneKeyId);
	}

	function findChildIndex(
		childs:
			| {
					parentSubPanel: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "tabs">;
					parentPanel: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
					subPanel: Omit<VizSubPanel, "childs">[];
			  }
			| undefined,
		paneKeyId: string | undefined
	) {
		return childs?.subPanel?.findIndex((sub: any) => sub.paneKeyId === paneKeyId) ?? -1;
	}

	function getSubPanelParent(layout: VizSubPanel[], paneKeyId: string | undefined) {
		if (!paneKeyId) return null;
		for (const panel of layout) {
			if (!panel.childs?.subPanel) continue;
			for (const sub of panel.childs.subPanel) {
				if (sub.paneKeyId === paneKeyId) return panel.paneKeyId;
			}
		}
		return null;
	}

	function ondrop(node: HTMLElement, event: DragEvent) {
		event.preventDefault();

		if (!event.dataTransfer) {
			return;
		}

		const data = event.dataTransfer.getData("text/json");
		const state = JSON.parse(data) as TabData;
		const tabKeyId = node.getAttribute("data-tab-id")!;
		const nodeParentId = node.parentElement?.getAttribute("data-viz-sp-id");

		if (!nodeParentId) {
			throw new Error("Viz: Node parent ID is missing");
		}

		if (state.data.id === parseInt(tabKeyId)) {
			return;
		}

		if (!panelTabs.some((tab) => tab.id === state.data.id)) {
			const splicedTabs = $allTabs.get(state.data.parent!)?.splice(state.index, 1);

			if (!splicedTabs?.length) {
				return;
			}

			state.data.component = splicedTabs[0].component;
			state.data.isActive = true;

			panelTabs.push(...splicedTabs);
			$allTabs.set(keyId, [...panelTabs]);

			const layout = $layoutState;
			const parentIdx = findPanelIndex(layout, state.data.parent);
			const childs = layout[parentIdx]?.childs;

			const childIdx = findChildIndex(childs, nodeParentId);
			const childPanel = childs?.subPanel?.[childIdx];

			const srcParent = getSubPanelParent(layout, state.data.parent);
			const dstParent = getSubPanelParent(layout, nodeParentId);

			// --- All move logic below ---
			// 1. Move tab between child subpanels of the same parent
			if (
				srcParent &&
				dstParent &&
				srcParent === dstParent &&
				state.data.parent !== nodeParentId &&
				childs &&
				Array.isArray(childs.subPanel)
			) {
				if (window.debug === true) {
					console.log("Move tab between child subpanels of the same parent");
				}

				const srcIdx = findChildIndex(childs, state.data.parent);
				const dstIdx = childIdx;

				if (srcIdx !== -1 && dstIdx !== -1) {
					const tabIdx = childs.subPanel[srcIdx].tabs.findIndex((tab: any) => tab.id === state.data.id);

					if (tabIdx !== -1) {
						const movedTab = childs.subPanel[srcIdx].tabs.splice(tabIdx, 1)[0];
						childs.subPanel[dstIdx].tabs.push(movedTab);
						movedTab.parent = nodeParentId;

						// Remove the source child subpanel if it is now empty
						if (!childs.subPanel[srcIdx].tabs.length) {
							childs.subPanel.splice(srcIdx, 1);
						}
					}
				}
			}

			// 2. Move tab from one parent subpanel to a different parent subpanel (or its child)
			else if (parentIdx !== -1 && state.data.parent !== nodeParentId && findPanelIndex(layout, nodeParentId) !== -1) {
				if (window.debug === true) {
					console.log("Move tab from one parent subpanel to a different parent subpanel (or its child)");
				}

				const srcIdx = findPanelIndex(layout, state.data.parent);
				const dstIdx = findPanelIndex(layout, nodeParentId);

				if (srcIdx !== -1 && dstIdx !== -1) {
					const srcTabs = layout[srcIdx].tabs;
					const tabIdx = srcTabs.findIndex((tab) => tab.id === state.data.id);
					let movedTab;

					if (tabIdx !== -1) {
						movedTab = srcTabs.splice(tabIdx, 1)[0];
						movedTab.parent = nodeParentId;
					}

					if (movedTab) {
						if (!layout[dstIdx].tabs) layout[dstIdx].tabs = [];
						layout[dstIdx].tabs.push(movedTab);
					}

					if (!srcTabs.length) {
						// Only promote if there are child subpanels, otherwise just remove the panel
						const srcPanel = layout[srcIdx];
						if (srcPanel.childs?.subPanel?.length) {
							promoteChildToParent(layout, srcIdx);
						} else {
							layout.splice(srcIdx, 1);
						}
					}
				}
			}

			// 3. Move tab from parent to its own child subpanel (promote child to parent)
			else if (parentIdx !== -1 && childPanel) {
				if (window.debug === true) {
					console.log("Move tab from parent to its own child subpanel (promote child to parent)");
				}

				const parentPanel = layout[parentIdx];
				const newParent = {
					...parentPanel,
					id: childPanel.id,
					maxSize: childPanel.maxSize,
					minSize: childPanel.minSize,
					paneKeyId: childPanel.paneKeyId,
					tabs: childPanel.tabs
				};

				newParent.tabs!.push(state.data);

				if (childs && childs.subPanel) {
					childs.subPanel = childs.subPanel.filter((panel: any) => panel.paneKeyId !== nodeParentId);
				}

				newParent.childs = childs;
				layout.splice(parentIdx, 1, newParent);
			}

			// 4. Move tab from parent to a child subpanel of a different parent
			else if (
				parentIdx !== -1 &&
				state.data.parent !== nodeParentId &&
				layout.some((panel) => panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === nodeParentId))
			) {
				if (window.debug === true) {
					console.log("Move tab from parent to a child subpanel of a different parent");
				}

				const srcIdx = findPanelIndex(layout, state.data.parent);
				const dstIdx = layout.findIndex((panel) => panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === nodeParentId));

				if (srcIdx !== -1 && dstIdx !== -1) {
					const srcTabs = layout[srcIdx].tabs;
					const tabIdx = srcTabs.findIndex((tab) => tab.id === state.data.id);
					let movedTab;

					if (tabIdx !== -1) {
						movedTab = srcTabs.splice(tabIdx, 1)[0];
						movedTab.parent = nodeParentId;
					}

					const destChildIdx = findChildIndex(layout[dstIdx].childs, nodeParentId);

					if (movedTab && destChildIdx !== -1 && layout[dstIdx].childs?.subPanel) {
						layout[dstIdx].childs.subPanel[destChildIdx].tabs.push(movedTab);
					}

					if (!srcTabs.length) {
						// Only promote if there are child subpanels, otherwise just remove the panel
						const srcPanel = layout[srcIdx];
						if (srcPanel.childs?.subPanel?.length) {
							promoteChildToParent(layout, srcIdx);
						} else {
							layout.splice(srcIdx, 1);
						}
					}
				}
			}

			// 5. Move tab from child subpanel to parent subpanel (and remove empty child subpanel)
			else if (keyId === nodeParentId && state.data.parent !== keyId) {
				if (window.debug === true) {
					console.log("Move tab from child subpanel to parent subpanel (and remove empty child subpanel)");
				}

				let srcParentIdx = layout.findIndex((panel) =>
					panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === state.data.parent)
				);
				let srcChildIdx = -1;

				if (srcParentIdx !== -1) {
					srcChildIdx = findChildIndex(layout[srcParentIdx].childs, state.data.parent);
				}

				let dstParentIdx = layout.findIndex(
					(panel) =>
						panel.paneKeyId === nodeParentId || panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === nodeParentId)
				);

				// FIX: Only check that both indices are valid
				if (srcParentIdx !== -1 && dstParentIdx !== -1) {
					const srcChild = layout[srcParentIdx].childs?.subPanel[srcChildIdx];
					if (!srcChild) {
						throw new Error("Viz: No source child subpanel found");
					}

					const tabIdx = srcChild.tabs.findIndex((tab) => tab.id === state.data.id);
					if (tabIdx === -1) {
						throw new Error("Viz: Tab not found in source child subpanel");
					}

					const movedTab = srcChild.tabs.splice(tabIdx, 1)[0];

					// Remove the source child subpanel if it is now empty
					if (srcChild.tabs.length === 0) {
						layout[srcParentIdx].childs?.subPanel.splice(srcChildIdx, 1);
					}

					if (layout[dstParentIdx].paneKeyId === nodeParentId) {
						if (!layout[dstParentIdx].tabs) layout[dstParentIdx].tabs = [];
						layout[dstParentIdx].tabs.push(movedTab);
						movedTab.parent = nodeParentId;
					} else {
						const dstChildIdx = findChildIndex(layout[dstParentIdx].childs, nodeParentId);
						if (dstChildIdx !== -1) {
							layout[dstParentIdx].childs?.subPanel[dstChildIdx].tabs.push(movedTab);
							movedTab.parent = nodeParentId;
						}
					}
				}
			}

			$layoutState = [...layout];

			state.data.parent = nodeParentId;
			state.index = panelTabs.length - 1;
			activeTab = state.data;
			return;
		}

		if (panelTabs.length === 1) {
			return;
		}

		const originalTab = panelTabs.find((tab) => tab.id === state.data.id);
		if (!originalTab) {
			return;
		}

		state.data.component = originalTab.component;
		const tabIndex = panelTabs.findIndex((tab) => tab.id === state.data.id);

		if (tabIndex === panelTabs.length - 1) {
			activeTab = state.data;
			return;
		}

		if (node.classList.contains("viz-sub_panel-header") && tabIndex === state.index) {
			panelTabs.push(state.data);
			if (state.index === 0) {
				panelTabs.splice(state.index, 1);
			} else {
				panelTabs.splice(state.index - 1, 1);
			}
		} else if (tabIndex === state.index) {
			swapArrayElements(
				panelTabs,
				state.index,
				panelTabs.findIndex((tab) => tab.id === parseInt(node.getAttribute("data-tab-id")!))
			);
		}

		activeTab = state.data;
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
	{#if header && panelTabs.length > 0}
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
						title={tab.name}
						aria-label={tab.name}
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
