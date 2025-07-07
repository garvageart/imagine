import { SvelteMap } from "svelte/reactivity";
import type { IPaneSerialized, ITree } from ".";
import { writable } from "svelte/store";
import type { VizSubPanel, VizView } from "$lib/components/panels/SubPanel.svelte";

// this might cause bugs idk
export const allSplitpanes = writable(new SvelteMap<string, IPaneSerialized[]>());
export const layoutState: { tree: VizSubPanel[]; } = $state({
    tree: []
});
export const allTabs = writable(new SvelteMap<string, VizView[]>());
export const layoutTree = $state({}) as ITree

export const getAllSubPanels = () => {
    let subPanels = layoutState.tree.flat();

    // I hate this so much
    if (subPanels.flatMap((panel) => panel.childs).length > 0) {
        subPanels = subPanels?.concat(subPanels.flatMap((panel) => panel.childs?.subPanel ?? []));

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