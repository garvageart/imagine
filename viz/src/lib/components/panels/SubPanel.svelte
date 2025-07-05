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
				parentSubPanel: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">;
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
	import { getAllSubPanels, layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { views } from "$lib/layouts/test";
	import { dev } from "$app/environment";

	if (dev) {
		window.resetAndReloadLayout = resetAndReloadLayout;
	}

	interface Props {
		id: string;
		header?: boolean;
		views: VizView[];
		children?: Snippet;
	}

	interface TabData {
		index: number;
		view: VizView;
	}

	const defaultClass = "viz-panel";
	let className: string = $state(defaultClass);

	const allProps: Props & ComponentProps<typeof Pane> = $props();

	let id = allProps.id;
	let header = allProps.header ?? true;

	const children = allProps.children;
	const keyId = allProps.paneKeyId ?? generateKeyId();
	const minSize = allProps.minSize ?? 10;

	const allViews = getAllSubPanels().flatMap((subpanel) => subpanel.views ?? []);
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
				views: firstChild.views,
				childs: {
					...parentPanel.childs,
					subPanel: parentPanel.childs.subPanel.slice(1)
				}
			});
		} else {
			currentLayout.splice(parentIndex, 1);
		}

		if (window.debug === true) {
			console.log(`Promoting child ${parentPanel.paneKeyId}`, parentPanel);
		}
	}

	function findPanelIndex(layout: VizSubPanel[], paneKeyId: string | undefined) {
		return layout.findIndex((panel) => panel.paneKeyId === paneKeyId);
	}

	function findChildIndex(
		childs:
			| {
					parentSubPanel: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">;
					parentPanel: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
					subPanel: Omit<VizSubPanel, "childs">[];
			  }
			| undefined,
		paneKeyId: string | undefined
	) {
		return childs?.subPanel?.findIndex((sub: any) => sub.paneKeyId === paneKeyId) ?? -1;
	}

	function getSubPanelParent(layout: VizSubPanel[], paneKeyId: string | undefined) {
		if (!paneKeyId) {
			return null;
		}

		for (const panel of layout) {
			if (!panel.childs?.subPanel) {
				continue;
			}

			for (const sub of panel.childs.subPanel) {
				if (sub.paneKeyId === paneKeyId) {
					return panel.paneKeyId;
				}
			}
		}

		return null;
	}

	async function ondrop(node: HTMLElement, event: DragEvent) {
		event.preventDefault();

		if (!event.dataTransfer) {
			return;
		}

		const data = event.dataTransfer.getData("text/json");
		const state = JSON.parse(data) as TabData;
		const tabKeyId = node.getAttribute("data-tab-id")!;
		const nodeParentId = node.parentElement?.getAttribute("data-viz-sp-id");
		const nodeIsPanelHeader = node.classList.contains("viz-sub_panel-header");
		const nodeIsTab = node.classList.contains("viz-tab-button") && node.hasAttribute("data-tab-id");

		if (!nodeParentId && nodeIsPanelHeader) {
			throw new Error("Viz: Node parent ID is missing");
		}

		if (!nodeParentId) {
			return;
		}

		if (state.view.id === parseInt(tabKeyId)) {
			return;
		}

		if (window.debug) {
			console.log(`Attempting to move ${state.view.name} to ${nodeParentId}`);
		}

		if (!panelViews.some((view) => view.id === state.view.id)) {
			const tab = allViews.find((view) => view.id === state.view.id);

			if (!tab) {
				return;
			}

			const layout = layoutState.tree;
			const parentIdx = findPanelIndex(layout, state.view.parent);
			const childs = layout[parentIdx]?.childs;

			const childIdx = findChildIndex(childs, nodeParentId);
			const childPanel = childs?.subPanel?.[childIdx];

			const srcParent = getSubPanelParent(layout, state.view.parent);
			const dstParent = getSubPanelParent(layout, nodeParentId);

			// --- All move logic below ---
			// 1. Move tab between child subpanels of the same parent
			if (
				srcParent &&
				dstParent &&
				srcParent === dstParent &&
				state.view.parent !== nodeParentId &&
				childs &&
				Array.isArray(childs.subPanel)
			) {
				if (window.debug === true) {
					console.log("Move tab between child subpanels of the same parent");
				}

				const srcIdx = findChildIndex(childs, state.view.parent);
				const dstIdx = childIdx;

				if (srcIdx !== -1 && dstIdx !== -1) {
					const tabIdx = childs.subPanel[srcIdx].views.findIndex((tab) => tab.id === state.view.id);

					if (tabIdx !== -1) {
						const movedTab = childs.subPanel[srcIdx].views.splice(tabIdx, 1)[0];
						childs.subPanel[dstIdx].views.push(movedTab);
						movedTab.parent = nodeParentId;

						// Remove the source child subpanel if it is now empty
						if (!childs.subPanel[srcIdx].views.length) {
							childs.subPanel.splice(srcIdx, 1);
							layout[parentIdx].childs = childs;
						}
					}
				}
			}

			// 2. Move tab from one parent subpanel to a different parent subpanel (or its child)
			else if (parentIdx !== -1 && state.view.parent !== nodeParentId && findPanelIndex(layout, nodeParentId) !== -1) {
				if (window.debug === true) {
					console.log("Move tab from one parent subpanel to a different parent subpanel (or its child)");
				}

				const srcIdx = findPanelIndex(layout, state.view.parent);
				const dstIdx = findPanelIndex(layout, nodeParentId);

				if (srcIdx !== -1 && dstIdx !== -1) {
					const srcTabs = layout[srcIdx].views;
					const tabIdx = srcTabs.findIndex((tab) => tab.id === state.view.id);
					let movedTab;

					if (tabIdx !== -1) {
						movedTab = srcTabs.splice(tabIdx, 1)[0];
						movedTab.parent = nodeParentId;
					}

					if (movedTab) {
						if (!layout[dstIdx].views) {
							layout[dstIdx].views = [];
						}

						layout[dstIdx].views.push(movedTab);
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

					// explicitly set the size of the one and only subpanel to 100
					// splitpanes doesn't necessarily understand that to recalculate automatically oops
					if (layout.length === 1 && layout[0].childs) {
						if (window.debug === true) {
							console.log(`one panel ${layout[0].paneKeyId} left, setting maximum size to 100`);
						}

						layout[0].childs.parentSubPanel.size = 100;
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
					views: childPanel.views
				};

				newParent.views!.push(state.view);

				if (childs && childs.subPanel) {
					childs.subPanel = childs.subPanel.filter((panel: any) => panel.paneKeyId !== nodeParentId);
				}

				newParent.childs = childs;
				layout.splice(parentIdx, 1, newParent);
			}

			// 4. Move tab from parent to a child subpanel of a different parent
			else if (
				parentIdx !== -1 &&
				state.view.parent !== nodeParentId &&
				layout.some((panel) => panel.childs?.subPanel?.some((sub) => sub.paneKeyId === nodeParentId))
			) {
				if (window.debug === true) {
					console.log("Move tab from parent to a child subpanel of a different parent");
				}

				const srcIdx = findPanelIndex(layout, state.view.parent);
				const dstIdx = layout.findIndex((panel) => panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === nodeParentId));

				if (srcIdx !== -1 && dstIdx !== -1) {
					const srcViews = layout[srcIdx].views;
					const viewIdx = srcViews.findIndex((view) => view.id === state.view.id);
					let movedView;

					if (viewIdx !== -1) {
						movedView = srcViews.splice(viewIdx, 1)[0];
						movedView.parent = nodeParentId;
					}

					const destChildIdx = findChildIndex(layout[dstIdx].childs, nodeParentId);

					if (movedView && destChildIdx !== -1 && layout[dstIdx].childs?.subPanel) {
						layout[dstIdx].childs.subPanel[destChildIdx].views.push(movedView);
					}

					if (!srcViews.length) {
						// Only promote if there are child subpanels, otherwise just remove the panel
						const srcPanel = layout[srcIdx];
						if (srcPanel.childs?.subPanel?.length) {
							promoteChildToParent(layout, srcIdx);
						} else {
							layout.splice(srcIdx, 1);
						}
					}

					// explicitly set the size of the one and only subpanel to 100
					// splitpanes doesn't necessarily understand that to recalculate automatically oops
					if (layout.length === 1 && layout[0].childs) {
						if (window.debug === true) {
							console.log(`one panel ${layout[0].paneKeyId} left, setting maximum size to 100`);
						}

						layout[0].childs.parentSubPanel.size = 100;
					}
				}
			}

			// 5. Move tab from child subpanel to parent subpanel (and remove empty child subpanel)
			else if (keyId === nodeParentId && state.view.parent !== keyId) {
				if (window.debug === true) {
					console.log("Move tab from child subpanel to parent subpanel (and remove empty child subpanel)");
				}

				let srcParentIdx = layout.findIndex((panel) =>
					panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === state.view.parent)
				);
				let srcChildIdx = -1;

				if (srcParentIdx !== -1) {
					srcChildIdx = findChildIndex(layout[srcParentIdx].childs, state.view.parent);
				}

				let dstParentIdx = layout.findIndex(
					(panel) => panel.paneKeyId === nodeParentId || panel.childs?.subPanel?.some((sub) => sub.paneKeyId === nodeParentId)
				);

				// FIX: Only check that both indices are valid
				if (srcParentIdx !== -1 && dstParentIdx !== -1) {
					const srcChild = layout[srcParentIdx].childs?.subPanel[srcChildIdx];
					if (!srcChild) {
						throw new Error("Viz: No source child subpanel found");
					}

					const viewIdx = srcChild.views.findIndex((view) => view.id === state.view.id);
					if (viewIdx === -1) {
						throw new Error("Viz: Tab not found in source child subpanel");
					}

					const movedView = srcChild.views.splice(viewIdx, 1)[0];

					// Remove the source child subpanel if it is now empty
					if (srcChild.views.length === 0) {
						layout[srcParentIdx].childs?.subPanel.splice(srcChildIdx, 1);
					}

					if (layout[dstParentIdx].paneKeyId === nodeParentId) {
						if (!layout[dstParentIdx].views) {
							layout[dstParentIdx].views = [];
						}

						layout[dstParentIdx].views.push(movedView);
						movedView.parent = nodeParentId;
					} else {
						const dstChildIdx = findChildIndex(layout[dstParentIdx].childs, nodeParentId);
						if (dstChildIdx !== -1) {
							layout[dstParentIdx].childs?.subPanel[dstChildIdx].views.push(movedView);
						}
					}
				}
			} else {
				console.error(tab);
				throw new Error("Viz: Invalid tab movement");
			}

			tab.parent = nodeParentId;
			tab.isActive = true;
			tab.component = tab.component;
			activeView = tab;

			return;
		}

		// No tabs to reconfigure if it's the only one in the subpanel
		if (panelViews.length === 1) {
			return;
		}

		const originalView = views.find((view) => view.id === state.view.id);
		if (!originalView) {
			return;
		}

		const viewIndex = panelViews.findIndex((view) => view.id === state.view.id);

		if (viewIndex === panelViews.length - 1) {
			activeView = originalView;
			return;
		}

		// if we're dropping on the header, add it to the end of the header and
		// remove it from it's old position
		if (node.classList.contains("viz-sub_panel-header") && viewIndex === state.index) {
			panelViews.push(state.view);
			if (state.index === 0) {
				panelViews.splice(state.index, 1);
			} else {
				panelViews.splice(state.index - 1, 1);
			}
		} else if (viewIndex === state.index) {
			swapArrayElements(
				panelViews,
				state.index,
				panelViews.findIndex((view) => view.id === parseInt(node.getAttribute("data-tab-id")!))
			);

			return;
		}

		activeView = originalView;
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
