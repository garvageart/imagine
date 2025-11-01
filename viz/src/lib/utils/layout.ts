import type { Content, VizSubPanel } from "$lib/components/panels/SubPanel.svelte";
import VizSubPanelData from "$lib/layouts/subpanel.svelte";
import { layoutState } from "$lib/third-party/svelte-splitpanes/state.svelte";
import type { TabData } from "$lib/views/tabs.svelte";
import type VizView from "$lib/views/views.svelte";
import { generateRandomString } from "./misc";

export function generateKeyId(length = 10): string {
    return "sp-" + generateRandomString(length);
}

/**
 * Type guard for VizSubPanelData instances returned by findSubPanel().
 */
export function isVizSubPanelData(obj: any): obj is VizSubPanelData {
    return obj instanceof VizSubPanelData;
}
/**
 * Returns a flattened array of all subpanels in the layout. This is useful if you want to iterate over all subpanels
 * in the layout for any reason.
 *
 * @returns {VizSubPanel[]} - An array of all subpanels in the layout.
 */

export const getAllSubPanels = (): VizSubPanel[] => {
    const subPanels = layoutState.tree.flat();
    return subPanels;
};

/**
 * Finds a subpanel in the layoutState tree by its key.
 * @param key The key to search for
 * @param value The value to search for
 * @returns The subpanel found, or null if not found, or an object with the following properties:
 * - `parentIndex`: The index in the `layoutState.tree` array of the parent of the subpanel.
 * - `childIndex`: The index in the `parent.childs.subPanel` array of the subpanel.
 * - `isChild`: Whether the subpanel is a child of another subpanel.
 * - `subPanel`: The subpanel found.
 */
export function findSubPanel(key: keyof VizSubPanelData, value: VizSubPanelData[keyof VizSubPanelData]) {
    let parentIndex = layoutState.tree.findIndex((panel) => panel[key as keyof VizSubPanelData] === value);
    let subPanel: VizSubPanelData | Content | undefined;
    let childIndex = -1;
    let isChild = false;

    if (parentIndex !== -1) {
        childIndex = parentIndex;
        subPanel = layoutState.tree[parentIndex];
        return { parentIndex, childIndex, isChild, subPanel };
    }

    if (!subPanel) {
        for (let i = 0; i < layoutState.tree.length; i++) {
            const panel = layoutState.tree[i];
            if (!panel.childs?.content) {
                continue;
            }

            for (let j = 0; j < panel.childs.content.length; j++) {
                const sub = panel.childs.content[j];
                if (sub[key as keyof Content] === value) {
                    subPanel = sub;
                    isChild = true;
                    parentIndex = i;
                    childIndex = j;

                    break;
                }
            }
        }
    }

    if (!subPanel) {
        return null;
    }

    return {
        parentIndex,
        childIndex,
        isChild,
        subPanel
    };
}

export function addViewToContent(view: VizView, parentIndex: number, contentIndex: number) {
    layoutState.tree[parentIndex].childs.content[contentIndex].views.push(view);
}

export function removeViewFromContent(view: VizView, parentIndex: number, contentIndex: number) {
    layoutState.tree[parentIndex].childs.content[contentIndex].views = layoutState.tree[parentIndex].childs.content[contentIndex].views.filter((v) => v !== view);
}

export function isTabData(obj: any): obj is TabData {
    const objKeys = Object.keys(obj);
    const hasValidAttrs = objKeys.includes("view") && objKeys.includes("index");
    return obj !== null && typeof obj === 'object' && hasValidAttrs;

}