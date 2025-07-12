import type { ComponentProps } from "svelte";
import type { VizSubPanel } from "../components/panels/SubPanel.svelte";
import { views } from "../layouts/test";
import type { Splitpanes } from "../third-party/svelte-splitpanes";
import { findSubPanel, getAllSubPanels, layoutState } from "../third-party/svelte-splitpanes/state.svelte";
import { swapArrayElements } from "../utils";
import VizView from "./views.svelte";

export interface TabData {
    index: number;
    view: VizView;
}

class TabOps {
    private panelViews: VizView[];
    private keyId: string;
    public activeView: VizView | null = $state(null);

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
        if (parentPanel.childs?.subPanels?.length) {
            const firstChild = parentPanel.childs.subPanels[0];
            Object.assign(parentPanel, {
                id: firstChild.id,
                maxSize: firstChild.maxSize,
                minSize: firstChild.minSize,
                paneKeyId: firstChild.paneKeyId,
                views: firstChild.views,
                childs: {
                    ...parentPanel.childs,
                    subPanel: parentPanel.childs.subPanels.slice(1)
                }
            });
        } else {
            currentLayout.splice(parentIndex, 1);
        }

        if (parentPanel.childs?.subPanels?.length === 0) {
            parentPanel.size = parentPanel.childs.internalSubPanelContainer.size;
            parentPanel.minSize = parentPanel.childs.internalSubPanelContainer.minSize;
            parentPanel.maxSize = parentPanel.childs.internalSubPanelContainer.maxSize;

            delete parentPanel.childs;
        }

        if (window.debug === true) {
            console.log(`Promoting child ${$state.snapshot(parentPanel.paneKeyId)}`, $state.snapshot(parentPanel));
        }
    }

    /**
     * Finds the index of a panel in the given layout by its pane key ID.
     *
     * @param {VizSubPanel[]} layout The layout to search in.
     * @param {string|undefined} paneKeyId The pane key ID to search for.
     * @returns {number} The index of the panel in the layout, or -1 if not found.
     */
    findPanelIndex(layout: VizSubPanel[], paneKeyId: string | undefined): number {
        return layout.findIndex((panel) => panel.paneKeyId === paneKeyId);
    }

    /**
     * Finds the index of a child subpanel in the given child structure by its pane key ID.
     *
     * @param {{internalSubPanelContainer: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">; internalPanelContainer: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">; subPanel: Omit<VizSubPanel, "childs">[]} | undefined} childs The child structure to search in.
     * @param {string|undefined} paneKeyId The pane key ID to search for.
     * @returns {number} The index of the child subpanel in the child structure, or -1 if not found.
     */
    findChildIndex(
        childs:
            | {
                internalSubPanelContainer: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">;
                internalPanelContainer: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">;
                subPanels: Omit<VizSubPanel, "childs">[];
            }
            | undefined,
        paneKeyId: string | undefined
    ): number {
        return childs?.subPanels?.findIndex((sub) => sub.paneKeyId === paneKeyId) ?? -1;
    }


    /**
     * Retrieves the pane key ID of the parent panel for a given subpanel identified by its pane key ID.
     *
     * @param {VizSubPanel[]} layout - The layout containing panels and their subpanels.
     * @param {string | undefined} paneKeyId - The pane key ID of the subpanel to find the parent for.
     * @returns {string | null} The pane key ID of the parent panel, or null if not found.
     */
    getSubPanelParent(layout: VizSubPanel[], paneKeyId: string | undefined): string | null {
        if (!paneKeyId) {
            return null;
        }

        for (const panel of layout) {
            if (!panel.childs?.subPanels) {
                continue;
            }

            for (const sub of panel.childs.subPanels) {
                if (sub.paneKeyId === paneKeyId) {
                    return panel.paneKeyId ?? null;;
                }
            }
        }

        return null;
    }

    /**
     * Moves a tab to a different parent panel.
     *
     * @param {VizSubPanel[]} layout - The layout of the subpanels.
     * @param {TabData} state - The state of the tab.
     * @param {string} nodeParentId - The pane key ID of the new parent panel.
     */
    private moveTabDifferentParent(layout: VizSubPanel[], state: TabData, nodeParentId: string) {
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
                if (srcPanel.childs?.subPanels?.length) {
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

    /**
     * Adds a tab to an existing parent subpanel.
     *
     * @param {VizSubPanel[]} layout - The layout of the subpanels.
     * @param {TabData} state - The state of the tab.
     * @param {number} parentIdx - The index of the parent subpanel.
     * @param {string} nodeParentId - The ID of the parent node.
     * @param {number} childIdx - The index of the child subpanel.
     */
    private addTabToExistingParent(layout: VizSubPanel[], state: TabData, parentIdx: number, nodeParentId: string, childIdx: number) {
        const childs = layout[parentIdx].childs;
        const srcIdx = this.findChildIndex(childs, state.view.parent);
        const dstIdx = childIdx;

        if (childs && srcIdx !== -1 && dstIdx !== -1) {
            const tabIdx = childs.subPanels[srcIdx].views.findIndex((tab) => tab.id === state.view.id);

            if (tabIdx !== -1) {
                const movedTab = childs.subPanels[srcIdx].views.splice(tabIdx, 1)[0];
                childs.subPanels[dstIdx].views.push(movedTab);
                movedTab.parent = nodeParentId;

                // Remove the source child subpanel if it is now empty
                if (!childs.subPanels[srcIdx].views.length) {
                    childs.subPanels.splice(srcIdx, 1);
                    layout[parentIdx].childs = childs;
                }
            }
        }
    }

    /**
     * Adds a tab to a new parent subpanel.
     *
     * @param {VizSubPanel[]} layout - The layout of the subpanels.
     * @param {TabData} state - The state of the tab.
     * @param {number} parentIdx - The index of the parent subpanel.
     * @param {string} nodeParentId - The ID of the parent node.
     * @param {number} childIdx - The index of the child subpanel.
     */
    private addTabToNewParent(layout: VizSubPanel[], state: TabData, parentIdx: number, nodeParentId: string, childIdx: number) {
        const parentPanel = layout[parentIdx];
        const tabIndex = parentPanel.views.findIndex((view) => view.id === state.view.id);
        const childs = parentPanel?.childs;
        const childPanel = childs?.subPanels?.[childIdx];

        if (childPanel && tabIndex !== -1) {
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

    /**
     * Moves a tab within the same parent subpanel.
     * @param {VizSubPanel[]} layout - The layout of the subpanels.
     * @param {TabData} state - The state of the tab.
     * @param {string} nodeParentId - The ID of the parent node.
     */
    private moveTabToSameParent(layout: VizSubPanel[], state: TabData, nodeParentId: string) {
        const srcIdx = this.findPanelIndex(layout, state.view.parent);
        const dstIdx = layout.findIndex((panel) => panel.childs?.subPanels?.some((sub: any) => sub.paneKeyId === nodeParentId));

        if (srcIdx !== -1 && dstIdx !== -1) {
            const srcViews = layout[srcIdx].views;
            const viewIdx = srcViews.findIndex((view) => view.id === state.view.id);
            let movedView;

            if (viewIdx !== -1) {
                movedView = srcViews.splice(viewIdx, 1)[0];
                movedView.parent = nodeParentId;
            }

            const destChildIdx = this.findChildIndex(layout[dstIdx].childs, nodeParentId);

            if (movedView && destChildIdx !== -1 && layout[dstIdx].childs?.subPanels) {
                layout[dstIdx].childs.subPanels[destChildIdx].views.push(movedView);
            }

            if (!srcViews.length) {
                // Only promote if there are child subpanels, otherwise just remove the panel
                const srcPanel = layout[srcIdx];
                if (srcPanel.childs?.subPanels?.length) {
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

    /**
     * Moves a tab to a new child subpanel.
     * @param {VizSubPanel[]} layout - The layout of the subpanels.
     * @param {TabData} state - The state of the tab.
     * @param {string} nodeParentId - The ID of the parent node.
     */
    private moveTabToNewChild(layout: VizSubPanel[], state: TabData, nodeParentId: string) {
        let srcParentIdx = layout.findIndex((panel) =>
            panel.childs?.subPanels?.some((sub) => sub.paneKeyId === state.view.parent)
        );
        let srcChildIdx = -1;

        if (srcParentIdx !== -1) {
            srcChildIdx = this.findChildIndex(layout[srcParentIdx].childs, state.view.parent);
        }

        let dstParentIdx = layout.findIndex(
            (panel) => panel.paneKeyId === nodeParentId || panel.childs?.subPanels?.some((sub) => sub.paneKeyId === nodeParentId)
        );

        // FIX: Only check that both indices are valid
        if (srcParentIdx !== -1 && dstParentIdx !== -1) {
            const srcChild = layout[srcParentIdx].childs?.subPanels[srcChildIdx];
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
                layout[srcParentIdx].childs?.subPanels.splice(srcChildIdx, 1);
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
                    layout[dstParentIdx].childs?.subPanels[dstChildIdx].views.push(movedView);
                }
            }

            if (!layout[srcChildIdx].childs?.subPanels.length) {
                layout[srcParentIdx].size = layout[srcParentIdx].childs?.internalSubPanelContainer.size;
                layout[srcParentIdx].minSize = layout[srcParentIdx].childs?.internalSubPanelContainer.minSize;
                layout[srcParentIdx].maxSize = layout[srcParentIdx].childs?.internalSubPanelContainer.maxSize;

                delete layout[srcParentIdx].childs;
            }
        }
    }

    /**
     * Handles the drop event of a draggable element.
     * @param {DragEvent} event The drop event.
     */
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
            const tab = getAllSubPanels().flatMap((subpanel) => subpanel.views ?? []).find((view) => view.id === state.view.id);

            if (!tab) {
                return;
            }

            const layout = layoutState.tree;
            const parentIdx = this.findPanelIndex(layout, state.view.parent);
            const childs = layout[parentIdx]?.childs;

            const childIdx = this.findChildIndex(childs, nodeParentId);
            const childPanel = childs?.subPanels?.[childIdx];

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
                Array.isArray(childs.subPanels)
            ) {
                if (window.debug === true) {
                    console.log("Move tab between child subpanels of the same parent");
                }

                this.addTabToExistingParent(layout, state, parentIdx, nodeParentId, childIdx);
            }

            // 2. Move tab from one parent subpanel to a different parent subpanel (or its child)
            else if (parentIdx !== -1 && state.view.parent !== nodeParentId && this.findPanelIndex(layout, nodeParentId) !== -1) {
                if (window.debug === true) {
                    console.log("Move tab from one parent subpanel to a different parent subpanel (or its child)");
                }

                this.moveTabDifferentParent(layout, state, nodeParentId);
            }

            // 3. Move tab from parent to its own child subpanel (promote child to parent if no more tabs)
            else if (parentIdx !== -1 && childPanel) {
                if (window.debug === true) {
                    console.log("Move tab from parent to its own child subpanel");
                }

                this.addTabToNewParent(layout, state, parentIdx, nodeParentId, childIdx);
            }

            // 4. Move tab from parent to a child subpanel of a different parent
            else if (
                parentIdx !== -1 &&
                state.view.parent !== nodeParentId &&
                layout.some((panel) => panel.childs?.subPanels?.some((sub) => sub.paneKeyId === nodeParentId))
            ) {
                if (window.debug === true) {
                    console.log("Move tab from parent to a child subpanel of a different parent");
                }

                this.moveTabToSameParent(layout, state, nodeParentId);
            }

            // 5. Move tab from child subpanel to parent subpanel (and remove empty child subpanel)
            else if (this.keyId === nodeParentId && state.view.parent !== this.keyId) {
                if (window.debug === true) {
                    console.log("Move tab from child subpanel to parent subpanel (and remove empty child subpanel)");
                }

                this.moveTabToNewChild(layout, state, nodeParentId);
            } else {
                console.error(tab);
                throw new Error("Viz: Invalid tab movement");
            }

            tab.parent = nodeParentId;
            tab.isActive = true;
            this.activeView = tab;

            return;
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

    /**
     * Handles the dragover event of a draggable element.
     * @param {DragEvent} event The dragover event.
     */
    onDropOver(event: DragEvent) {
        event.preventDefault();
        if (event.dataTransfer) {
            event.dataTransfer.dropEffect = "move";
        }
    }

    /**
     * Makes an element draggable. */
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

    /**
     * Attaches drag-and-drop event listeners to an HTML element to handle
     * visual feedback and drop actions for draggable elements.
     * 
     * The function adds event listeners for "drop", "dragenter", "dragleave",
     * and "dragend" events. It applies a CSS class for visual feedback
     * when a draggable element is dragged over the target element and
     * invokes the `ondrop` handler when a drop occurs. The `destroy` 
     * method is provided to clean up the event listeners.
     * @param {HTMLElement} node The HTML element to which drag-and-drop
     *                           event listeners are attached.
     * @returns {Object} An object containing a `destroy` method to
     *                   remove the event listeners when no longer needed.
     */
    tabDrop(node: HTMLElement): object {
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

    // TODO: When dragging over the subpanel, determine the coordinates of where
    // in the subpanel we're hovering a create the dropzone within those bounds, usually half
    // note: probably debounce it a lil to avoid sudden layout shifts
    subPanelDropInside(node: HTMLElement) {
        node.addEventListener("dragenter", async (e) => {
            // console.log(Array.from(node.children)?.length > 1, node !== e.target);
            if (Array.from(node.children)?.length > 1) {
                return;
            }


            const elChildren = Array.from(node.children) as HTMLElement[];
            elChildren.forEach((el) => el.style.pointerEvents = "none");

            node.setAttribute("style", "height: calc(100% - 1.8em);");
            node.classList.add("viz-sub_panel-dropzone_overlay");
        });

        node.addEventListener("dragleave", (e) => {
            if (node !== e.target) {
                return;
            }

            node.removeAttribute("style");
            node.classList.remove("viz-sub_panel-dropzone_overlay");
            const elChildren = Array.from(node.children) as HTMLElement[];
            elChildren.forEach((el) => el.style.pointerEvents = "auto");
        });

        return {
            destroy: () => {
                node.removeEventListener("dragenter", async (e) => {
                    // console.log(Array.from(node.children)?.length > 1, node !== e.target);
                    if (Array.from(node.children)?.length > 1) {
                        return;
                    }


                    const elChildren = Array.from(node.children) as HTMLElement[];
                    elChildren.forEach((el) => el.style.pointerEvents = "none");

                    node.setAttribute("style", "height: calc(100% - 1.8em);");
                    node.classList.add("viz-sub_panel-dropzone_overlay");
                });

                node.removeEventListener("dragleave", (e) => {
                    if (node !== e.target) {
                        return;
                    }

                    node.removeAttribute("style");
                    node.classList.remove("viz-sub_panel-dropzone_overlay");
                    const elChildren = Array.from(node.children) as HTMLElement[];
                    elChildren.forEach((el) => el.style.pointerEvents = "auto");
                });
            }
        };
    }
}

export default TabOps;