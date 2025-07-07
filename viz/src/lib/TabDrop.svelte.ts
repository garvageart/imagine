import type { ComponentProps } from "svelte";
import type { VizSubPanel, VizView } from "./components/panels/SubPanel.svelte";
import { views } from "./layouts/test";
import type { Splitpanes } from "./third-party/svelte-splitpanes";
import { getAllSubPanels, layoutState } from "./third-party/svelte-splitpanes/state.svelte";
import { swapArrayElements } from "./utils";

export interface TabData {
    index: number;
    view: VizView;
}

class TabDropper {
    private panelViews: VizView[];
    private allViews: VizView[] = getAllSubPanels().flatMap((subpanel) => subpanel.views ?? []);
    private keyId: string;
    activeView: VizView | null = $state(null);

    constructor(keyId: string, panelViews: VizView[]) {
        this.keyId = keyId;
        this.panelViews = panelViews;
    }

    /**
     * Promotes the first child subpanel to the parent panel, or removes the parent if no children.
     * Mutates currentLayout in place.
     * @param currentLayout The current layout as a VizSubPanel[]
     * @param parentIndex The index in the currentLayout of the parent panel to promote a child from
     */
    promoteChildToParent(currentLayout: VizSubPanel[], parentIndex: number) {
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
            console.log(`Promoting child ${$state.snapshot(parentPanel.paneKeyId)}`, $state.snapshot(parentPanel));
        }
    }

    findPanelIndex(layout: VizSubPanel[], paneKeyId: string | undefined) {
        return layout.findIndex((panel) => panel.paneKeyId === paneKeyId);
    }

    findChildIndex(
        childs:
            | {
                internalSubPanelContainer: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">;
                internalPanelContainer: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
                subPanel: Omit<VizSubPanel, "childs">[];
            }
            | undefined,
        paneKeyId: string | undefined
    ) {
        return childs?.subPanel?.findIndex((sub) => sub.paneKeyId === paneKeyId) ?? -1;
    }

    getSubPanelParent(layout: VizSubPanel[], paneKeyId: string | undefined) {
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

    async ondrop(node: HTMLElement, event: DragEvent) {
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

        if (!this.panelViews.some((view) => view.id === state.view.id)) {
            const tab = this.allViews.find((view) => view.id === state.view.id);

            if (!tab) {
                return;
            }

            const layout = layoutState.tree;
            const parentIdx = this.findPanelIndex(layout, state.view.parent);
            const childs = layout[parentIdx]?.childs;

            const childIdx = this.findChildIndex(childs, nodeParentId);
            const childPanel = childs?.subPanel?.[childIdx];

            const srcParent = this.getSubPanelParent(layout, state.view.parent);
            const dstParent = this.getSubPanelParent(layout, nodeParentId);

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

                const srcIdx = this.findChildIndex(childs, state.view.parent);
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
            else if (parentIdx !== -1 && state.view.parent !== nodeParentId && this.findPanelIndex(layout, nodeParentId) !== -1) {
                if (window.debug === true) {
                    console.log("Move tab from one parent subpanel to a different parent subpanel (or its child)");
                }

                const srcIdx = this.findPanelIndex(layout, state.view.parent);
                const dstIdx = this.findPanelIndex(layout, nodeParentId);

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
                            this.promoteChildToParent(layout, srcIdx);
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

                        layout[0].childs.internalSubPanelContainer.size = 100;
                    }
                }
            }

            // 3. Move tab from parent to its own child subpanel (promote child to parent if no more tabs)
            else if (parentIdx !== -1 && childPanel) {
                if (window.debug === true) {
                    console.log("Move tab from parent to its own child subpanel");
                }

                const parentPanel = layout[parentIdx];
                const tabIndex = parentPanel.views.findIndex((view) => view.id === state.view.id);

                if (tabIndex !== -1) {
                    const movedTab = parentPanel.views.splice(tabIndex, 1)[0];
                    childPanel.views.push(movedTab);
                    movedTab.parent = nodeParentId;

                    // If parent has no more tabs, promote its first child.
                    if (parentPanel.views.length === 0) {
                        if (window.debug === true) {
                            console.log("Parent has no more tabs, promoting child");
                        }
                        this.promoteChildToParent(layout, parentIdx);
                    }
                }
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

                const srcIdx = this.findPanelIndex(layout, state.view.parent);
                const dstIdx = layout.findIndex((panel) => panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === nodeParentId));

                if (srcIdx !== -1 && dstIdx !== -1) {
                    const srcViews = layout[srcIdx].views;
                    const viewIdx = srcViews.findIndex((view) => view.id === state.view.id);
                    let movedView;

                    if (viewIdx !== -1) {
                        movedView = srcViews.splice(viewIdx, 1)[0];
                        movedView.parent = nodeParentId;
                    }

                    const destChildIdx = this.findChildIndex(layout[dstIdx].childs, nodeParentId);

                    if (movedView && destChildIdx !== -1 && layout[dstIdx].childs?.subPanel) {
                        layout[dstIdx].childs.subPanel[destChildIdx].views.push(movedView);
                    }

                    if (!srcViews.length) {
                        // Only promote if there are child subpanels, otherwise just remove the panel
                        const srcPanel = layout[srcIdx];
                        if (srcPanel.childs?.subPanel?.length) {
                            this.promoteChildToParent(layout, srcIdx);
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

                        layout[0].childs.internalSubPanelContainer.size = 100;
                    }
                }
            }

            // 5. Move tab from child subpanel to parent subpanel (and remove empty child subpanel)
            else if (this.keyId === nodeParentId && state.view.parent !== this.keyId) {
                if (window.debug === true) {
                    console.log("Move tab from child subpanel to parent subpanel (and remove empty child subpanel)");
                }

                let srcParentIdx = layout.findIndex((panel) =>
                    panel.childs?.subPanel?.some((sub: any) => sub.paneKeyId === state.view.parent)
                );
                let srcChildIdx = -1;

                if (srcParentIdx !== -1) {
                    srcChildIdx = this.findChildIndex(layout[srcParentIdx].childs, state.view.parent);
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
                        const dstChildIdx = this.findChildIndex(layout[dstParentIdx].childs, nodeParentId);
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
            this.activeView = tab;

            return tab;
        }

        // No tabs to reconfigure if it's the only one in the subpanel
        if (this.panelViews.length === 1) {
            return;
        }

        const originalView = views.find((view) => view.id === state.view.id);
        if (!originalView) {
            return;
        }

        const viewIndex = this.panelViews.findIndex((view) => view.id === state.view.id);

        if (viewIndex === this.panelViews.length - 1) {
            this.activeView = originalView;
            return;
        }

        // if we're dropping on the header, add it to the end of the header and
        // remove it from it's old position
        if (node.classList.contains("viz-sub_panel-header") && viewIndex === state.index) {
            this.panelViews.push(state.view);
            if (state.index === 0) {
                this.panelViews.splice(state.index, 1);
            } else {
                this.panelViews.splice(state.index - 1, 1);
            }
        } else if (viewIndex === state.index) {
            swapArrayElements(
                this.panelViews,
                state.index,
                this.panelViews.findIndex((view) => view.id === parseInt(node.getAttribute("data-tab-id")!))
            );

            return;
        }

        this.activeView = originalView;
    }

    onDropOver(event: DragEvent) {
        event.preventDefault();
        if (event.dataTransfer) {
            event.dataTransfer.dropEffect = "move";
        }
    }

    draggable(node: HTMLElement, data: TabData) {
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

    tabDrop(node: HTMLElement) {
        node.addEventListener("drop", (e) => {
            this.ondrop(node, e);
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
            destroy: () => {
                node.removeEventListener("drop", (e) => {
                    this.ondrop(node, e);
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
}

export default TabDropper;