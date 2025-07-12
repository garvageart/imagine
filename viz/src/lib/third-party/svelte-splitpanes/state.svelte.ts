import { SvelteMap } from "svelte/reactivity";
import type { IPaneSerialized, ITree } from ".";
import { writable } from "svelte/store";
import type { VizSubPanel } from "$lib/components/panels/SubPanel.svelte";
import type VizView from "$lib/views/views.svelte";

// this might cause bugs idk
export const allSplitpanes = writable(new SvelteMap<string, IPaneSerialized[]>());
export const layoutState: { tree: VizSubPanel[]; } = $state({
    tree: []
});
export const allTabs = writable(new SvelteMap<string, VizView[]>());
export const layoutTree = $state({}) as ITree;

export const getAllSubPanels = () => {
    let subPanels = layoutState.tree.flat();

    // I hate this so much
    if (subPanels.flatMap((panel) => panel.childs).length > 0) {
        subPanels = subPanels?.concat(subPanels.flatMap((panel) => panel.childs?.subPanels ?? []));

        if (subPanels.flatMap((panel) => panel.childs?.internalSubPanelContainer).length > 0) {
            subPanels = subPanels.concat(
                subPanels
                    .flatMap((panel) => panel.childs?.internalSubPanelContainer ?? [])
                    .filter((pane): pane is VizSubPanel => !!pane && typeof pane === "object" && "id" in pane)
            );
        }
    }

    return subPanels;

};

// TODO: Move to a seperate file
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
export function findSubPanel(key: keyof VizSubPanel, value: VizSubPanel[keyof VizSubPanel]) {
    let parentIndex = layoutState.tree.findIndex((panel) => panel[key as keyof VizSubPanel] === value);
    let subPanel: VizSubPanel | undefined;
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
            if (!panel.childs?.subPanels) {
                continue;
            }

            for (let j = 0; j < panel.childs.subPanels.length; j++) {
                const sub = panel.childs.subPanels[j];
                if (sub[key as keyof Omit<VizSubPanel, "childs">] === value) {
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