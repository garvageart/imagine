import type { VizSubPanel } from "../components/panels/SubPanel.svelte";
import { layoutState } from "../third-party/svelte-splitpanes/state.svelte";
import { sleep, swapArrayElements } from "../utils/misc";
import VizView from "./views.svelte";
import { dev } from "$app/environment";
import { views } from "$lib/layouts/views";
import { cleanupSubPanels, findChildIndex, findPanelIndex, getSubPanelParent } from "./utils";
import type { DropPosition } from "./types";
import VizSubPanelData, { Content } from "$lib/layouts/subpanel.svelte";

export interface TabData {
    index: number;
    view: VizView;
}

class TabOps {
    private panelViews: VizView[] = $state([]);
    public activeView: VizView | null = $state(null);

    constructor(panelViews: VizView[]) {
        this.panelViews = panelViews;
    }

    /**
     * Finds the index of a panel in the given layout by its pane key ID.
     *
     * @param {VizSubPanel[]} layout The layout to search in.
     * @param {string|undefined} paneKeyId The pane key ID to search for.
     * @returns {number} The index of the panel in the layout, or -1 if not found.
     */
    private _findPanelIndex = findPanelIndex;

    /**
     * Finds the index of a child subpanel in the given child structure by its pane key ID.
     *
     * @param {{internalSubPanelContainer: Omit<VizSubPanel, "childs" | "children" | "$$events" | "$$slots" | "header" | "views">; internalPanelContainer: Omit<ComponentProps<typeof Splitpanes>, "children" | "$$events" | "$$slots">; subPanel: Omit<VizSubPanel, "childs">[]} | undefined} childs The child structure to search in.
     * @param {string|undefined} paneKeyId The pane key ID to search for.
     * @returns {number} The index of the child subpanel in the child structure, or -1 if not found.
     */
    private _findChildIndex = findChildIndex;


    /**
     * Retrieves the pane key ID of the parent panel for a given subpanel identified by its pane key ID.
     *
     * @param {VizSubPanel[]} layout - The layout containing panels and their subpanels.
     * @param {string | undefined} paneKeyId - The pane key ID of the subpanel to find the parent for.
     * @returns {string | null} The pane key ID of the parent panel, or null if not found.
     */
    private _getSubPanelParent = getSubPanelParent;

    private _cleanupSubPanels = cleanupSubPanels;

    /**
     * Moves a tab to a new child subpanel.
     * @param {VizSubPanel[]} layout - The layout of the subpanels.
     * @param {TabData} state - The state of the tab.
     * @param {string} nodeParentId - The ID of the parent node.
     */
    private moveTabToNewChild(layout: VizSubPanel[], state: TabData, nodeParentId: string) {
        let srcParentIdx = layout.findIndex((panel) =>
            panel.childs.content?.some((sub) => sub.paneKeyId === state.view.parent)
        );
        let srcChildIdx = -1;

        if (srcParentIdx !== -1) {
            srcChildIdx = this._findChildIndex(layout[srcParentIdx].childs, state.view.parent);
        }

        let dstParentIdx = layout.findIndex(
            (panel) => panel.paneKeyId === nodeParentId || panel.childs.content?.some((sub) => sub.paneKeyId === nodeParentId)
        );

        // FIX: Only check that both indices are valid
        if (srcParentIdx !== -1 && dstParentIdx !== -1) {
            const srcChild = layout[srcParentIdx].childs.content[srcChildIdx];
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
                if (window.debug === true) {
                    console.log(`empty subpanel ${srcChild.paneKeyId}. removing it`);
                }

                layout[srcParentIdx].childs.content.splice(srcChildIdx, 1);
            }

            const dstChildIdx = this._findChildIndex(layout[dstParentIdx].childs, nodeParentId);
            if (dstChildIdx !== -1) {
                layout[dstParentIdx].childs.content[dstChildIdx].views.push(movedView);
            }

            const parentContent = layout[dstParentIdx].childs.content;
            if (Array.isArray(parentContent) && parentContent.length > 0) {
                const subSize = 100 / parentContent.length;
                parentContent.forEach(sub => {
                    sub.size = subSize;
                });
            }

            if (layout[srcParentIdx].childs.content.length === 0) {
                layout.splice(srcParentIdx, 1);

                if (window.debug === true) {
                    console.log(`one panel ${layout[0].paneKeyId} left, setting maximum size to 100`);
                }

                layout[0].childs.internalSubPanelContainer.size = 100;
            }
        }
    }

    /**
     * Handles the drop event of a draggable element.
     * @param {DragEvent} event The drop event.
     */
    private async ondrop(node: HTMLElement, event: DragEvent) {
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
        const panelContainsTab = this.panelViews.some((view) => view.id === state.view.id);

        if (!nodeParentId && nodeIsPanelHeader) {
            throw new Error("Viz: Node parent ID is missing");
        }

        if (!nodeParentId) {
            return;
        }

        if (state.view.id === parseInt(tabKeyId)) {
            return;
        }

        if (!panelContainsTab) {
            if (window.debug) {
                console.log(`Attempting to move ${state.view.name} to ${nodeParentId}`);
            }

            const layout = layoutState.tree;

            const srcParent = this._getSubPanelParent(layout, state.view.parent);
            const dstParent = this._getSubPanelParent(layout, nodeParentId);

            const parentIdx = this._findPanelIndex(layout, srcParent!);
            const childs = layout[parentIdx]?.childs;

            const tab = childs.content.find(panel => panel.views.find(view => view.id === state.view.id))?.views.find(view => view.id === state.view.id);

            if (!tab) {
                return;
            }

            // TODO: Clean up if statement checks (some unccessary checks in there)
            // --- All move logic below ---
            // 1. Move tab between child subpanels of the same parent
            if (
                srcParent &&
                dstParent &&
                srcParent === dstParent &&
                state.view.parent !== nodeParentId &&
                childs &&
                Array.isArray(childs.content)
            ) {
                if (window.debug === true) {
                    console.log("Move tab between child subpanels of the same parent");
                }

                this.moveTabToNewChild(layout, state, nodeParentId);
            }
            // 2. Move tab from parent to a child subpanel of a different parent
            else if (
                parentIdx !== -1 &&
                state.view.parent !== nodeParentId &&
                layout.some((panel) => panel.childs.content?.some((sub) => sub.paneKeyId === nodeParentId))
            ) {
                if (window.debug === true) {
                    console.log("Move tab from parent to a child subpanel of a different parent");
                }

                this.moveTabToNewChild(layout, state, nodeParentId);
            }
            // This shouldn't happen, error out if in dev, just log the error in prod and do nothing
            else {
                if (dev) {
                    throw new Error("Viz: Invalid tab movement");
                }

                console.error("Viz: Invalid tab movement", tab);
                return;
            }

            tab.parent = nodeParentId;
            tab.isActive = true;
            this.activeView = tab;

            return;
        }

        // No tabs to reconfigure if it's the only one in the subpanel or it's at the end of the subpanel
        if (this.panelViews.length === 1 || this.panelViews[this.panelViews.length - 1].id === state.view.id) {
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
            // FIXME: Add the tab in front of the tab we are dropping on instead of swapping them
            // e.g if tab at position 2 is dropped on tab position 5, it makes no sense to move 5 to 2, just put 2 in 5's position
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

        $effect(() => {
            state = JSON.stringify(data);
            return () => {
                node.draggable = false;
                node.removeEventListener("dragstart", (e) => {
                    e.dataTransfer?.setData("text/json", state);
                });
            };

        });
    }

    private handleTabDropInside(node: HTMLElement, event: DragEvent, activeViewData: TabData) {
        const data = event.dataTransfer!.getData("text/json");
        const state = JSON.parse(data) as TabData;

        const layout = layoutState.tree;
        const dstParentKeyId = this._getSubPanelParent(layout, activeViewData.view.parent!)!;
        const srcParentKeyId = this._getSubPanelParent(layout, state.view.parent!);
        const srcSubPanelIdx = layoutState.tree.find(p => p.paneKeyId === srcParentKeyId)?.childs.content.findIndex(sub => sub.paneKeyId === state.view.parent!)!;
        const srcSubPanel = layoutState.tree.find(p => p.paneKeyId === srcParentKeyId)?.childs.content[srcSubPanelIdx]!;

        const views = srcSubPanel.views!;
        const viewToMove = views.findIndex((view) => view.id === state.view.id);
        const dropZone = this.calculateDropZone(node, event);

        // Do some basic cleaning up and size normalization after moving the view
        // before the `svelte-splitpanes` handles the actual sizing after resize
        const cleanUpSubPanels = () => {
            views.splice(viewToMove!, 1);
            if (views.length === 0) {
                if (window.debug === true) {
                    console.log(`empty subpanel ${srcSubPanel.paneKeyId}. removing it`);
                }
                layoutState.tree.find(p => p.paneKeyId === srcParentKeyId)?.childs.content.splice(srcSubPanelIdx, 1);
            }

            // Normalize sizes for all subpanels in every panel's childs.content
            layout.forEach(panel => {
                if (panel.childs && Array.isArray(panel.childs.content) && panel.childs.content.length > 0) {
                    const subSize = 100 / panel.childs.content.length;
                    panel.childs.content.forEach(sub => {
                        sub.size = subSize;
                    });
                }
            });

            const srcPanelIdx = layout.findIndex(p => p.paneKeyId === srcParentKeyId);
            if (srcPanelIdx !== -1 && layout[srcPanelIdx].childs.content.length === 0) {
                layout.splice(srcPanelIdx, 1);
            }

            // Normalize sizes if more than one panel remains
            if (layout.length > 1) {
                const sizePerPanel = 100 / layout.length;
                layout.forEach(panel => {
                    panel.childs.internalSubPanelContainer.size = sizePerPanel;
                });
            } else if (layout.length === 1) {
                if (window.debug === true) {
                    console.log(`one panel ${layout[0].paneKeyId} left, setting maximum size to 100`);
                }
                layout[0].childs.internalSubPanelContainer.size = 100;
            }
        };

        switch (dropZone.dropPosition) {
            case "left":
                const newSubPanel = new VizSubPanelData({
                    content: [new Content({ views: [state.view] })]
                });
                layoutState.tree.splice(
                    this._findPanelIndex(layoutState.tree, dstParentKeyId),
                    0,
                    newSubPanel
                );

                break;
            case "right":
                const newSubPanelRight = new VizSubPanelData({
                    content: [new Content({ views: [state.view] })]
                });
                const index = this._findPanelIndex(layoutState.tree, dstParentKeyId);

                if (index === layoutState.tree.length - 1) {
                    layoutState.tree.push(newSubPanelRight);
                } else {
                    layoutState.tree.splice(
                        index,
                        0,
                        newSubPanelRight
                    );
                }

                break;
            case "top":
                layoutState.tree.find((panel) => panel.paneKeyId === dstParentKeyId)?.childs.content.splice(0, 0, new Content({ views: [state.view] }));
                break;
            case "bottom":
                layoutState.tree.find((panel) => panel.paneKeyId === dstParentKeyId)?.childs.content.push(new Content({ views: [state.view] }));
                break;
            default:
                console.error("Viz: Invalid drop position", dropZone.dropPosition);
                return;
        }

        cleanUpSubPanels();

        if (window.debug) {
            console.log("Creating panel from:", srcParentKeyId);
            console.log("Dropzone:", dropZone);
        }
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
     */
    tabDrop(node: HTMLElement) {
        $effect(() => {
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

            return () => {
                node.removeEventListener("drop", (e) => {
                    this.ondrop(node, e);
                });

                node.removeEventListener("dragenter", (e) => {
                    e.preventDefault();
                    if (node === e.target) {
                        return;
                    }

                    node.classList.add("drop-hover-above");
                });

                node.removeEventListener("dragleave", (e) => {
                    const target = e.target as HTMLElement;
                    if (node === target) {
                        return;
                    }

                    node.classList.remove("drop-hover-above");
                });

                node.removeEventListener("dragend", (e) => {
                    node.classList.remove("drop-hover-above");
                });
            };
        });
    }

    private calculateDropZone(node: HTMLElement, event: DragEvent | MouseEvent) {
        let dropPosition: DropPosition | null = null;
        const rect = node.getBoundingClientRect();

        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;

        const widthPercentage = (x / rect.width) * 100;
        const heightPercentage = (y / rect.height) * 100;

        const edgeThreshold = 40; // percent

        // Only allow left, right, top, or bottom
        if (widthPercentage < edgeThreshold) {
            dropPosition = "left";
        } else if (widthPercentage > 100 - edgeThreshold) {
            dropPosition = "right";
        } else if (heightPercentage < edgeThreshold) {
            dropPosition = "top";
        } else {
            dropPosition = "bottom";
        }

        return {
            x,
            y,
            dropPosition
        };
    }

    private handleDropInsideEnter(node: HTMLElement) {
        if (Array.from(node.children)?.length > 1) {
            return;
        }

        const elChildren = Array.from(node.children) as HTMLElement[];
        elChildren.forEach((el) => el.style.pointerEvents = "none");

        const overlayDiv = document.createElement("div");

        overlayDiv.style.pointerEvents = "none";
        overlayDiv.setAttribute("style", `height: ${node.clientHeight}px; width: ${node.clientWidth}px;`);
        overlayDiv.classList.add("viz-sub_panel-dropzone_overlay");
        node.insertBefore(overlayDiv, node.firstElementChild);
    }

    private drawDropzone(node: HTMLElement, e: MouseEvent) {
        const overlayDiv = node.querySelector(".viz-sub_panel-dropzone_overlay") as HTMLElement;
        if (!overlayDiv) {
            return;
        }

        const dropZone = this.calculateDropZone(node, e);
        overlayDiv.style.position = "absolute";

        switch (dropZone.dropPosition) {
            case "left":
                overlayDiv.style.left = `0px`;
                overlayDiv.style.top = `0px`;
                overlayDiv.style.width = `${node.clientWidth / 2}px`;
                overlayDiv.style.height = `${node.clientHeight}px`;
                break;
            case "right":
                overlayDiv.style.left = `${node.clientWidth / 2}px`;
                overlayDiv.style.top = `0px`;
                overlayDiv.style.width = `${node.clientWidth / 2}px`;
                overlayDiv.style.height = `${node.clientHeight}px`;
                break;
            case "top":
                overlayDiv.style.left = `0px`;
                overlayDiv.style.top = `0px`;
                overlayDiv.style.width = `${node.clientWidth}px`;
                overlayDiv.style.height = `${node.clientHeight / 2}px`;
                break;
            case "bottom":
                overlayDiv.style.left = `0px`;
                overlayDiv.style.top = `${node.clientHeight / 2}px`;
                overlayDiv.style.width = `${node.clientWidth}px`;
                overlayDiv.style.height = `${node.clientHeight / 2}px`;
                break;
        }
    }

    private removeDropzoneOverlay(node: HTMLElement, event: MouseEvent) {
        const overlayDiv = node.querySelector(".viz-sub_panel-dropzone_overlay") as HTMLElement;
        if (!overlayDiv) {
            return;
        }

        if (overlayDiv !== event.target) {
            return;
        }

        // TODO: adjust timings of the transition to stop flickering
        overlayDiv.style.transition = "width 0.2s ease-in-out, height 0.2s ease-in-out, left 0.2s ease-in-out, top 0.2s ease-in-out, display 0.2s ease-in-out";

        overlayDiv.style.width = "0px";
        overlayDiv.style.height = "0px";
        overlayDiv.style.left = "0px";
        overlayDiv.style.top = "0px";
        overlayDiv.style.display = "none";

        node.removeChild(overlayDiv);
        const elChildren = Array.from(node.children) as HTMLElement[];
        elChildren.forEach((el) => el.style.removeProperty("pointer-events"));
    }

    // TODO: When dragging over the subpanel, determine the coordinates of where
    // in the subpanel we're hovering a create the dropzone within those bounds, usually half
    // note: probably debounce it a lil to avoid sudden layout shifts
    subPanelDropInside(node: HTMLElement, data: TabData) {
        $effect(() => {
            node.addEventListener("dragenter", (e) => {
                this.handleDropInsideEnter(node);
            });

            node.addEventListener("drop", (e) => {
                e.preventDefault();

                this.handleTabDropInside(node, e, data);
                this.removeDropzoneOverlay(node, e);
            });

            node.addEventListener("dragover", (e) => {
                e.preventDefault();
                this.drawDropzone(node, e);
            });

            node.addEventListener("dragleave", async (e) => {
                e.preventDefault();

                await sleep(250).then(() => {
                    this.removeDropzoneOverlay(node, e);
                });
            });


            return () => {
                node.removeEventListener("dragenter", () => {
                    this.handleDropInsideEnter(node);
                });

                node.removeEventListener("drop", (e) => {
                    e.preventDefault();
                    this.removeDropzoneOverlay(node, e);
                });

                node.removeEventListener("dragover", (e) => {
                    e.preventDefault();
                    this.calculateDropZone(node, e);
                });
            };
        });

    }
}

export default TabOps;