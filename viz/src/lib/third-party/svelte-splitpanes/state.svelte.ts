import { SvelteMap } from "svelte/reactivity";
import type { IPaneSerialized, ITree } from ".";
import { writable } from "svelte/store";
import type { Content, VizSubPanel } from "$lib/components/panels/SubPanel.svelte";
import type VizView from "$lib/views/views.svelte";

// this might cause bugs idk
export const allSplitpanes = writable(new SvelteMap<string, IPaneSerialized[]>());
export const layoutState: { tree: VizSubPanel[]; } = $state({
    tree: []
});
export const allTabs = writable(new SvelteMap<string, VizView[]>());
export const layoutTree = $state({}) as ITree;

export const getAllSubPanels = () => {
    const subPanels = layoutState.tree.flat();
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
    let subPanel: VizSubPanel | Content | undefined;
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