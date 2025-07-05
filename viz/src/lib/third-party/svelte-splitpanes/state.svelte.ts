import { SvelteMap } from "svelte/reactivity";
import type { IPaneSerialized } from ".";
import { writable, type Writable } from "svelte/store";
import type { VizSubPanel, VizTab } from "$lib/components/panels/SubPanel.svelte";

// this might cause bugs idk
export const allSplitpanes = writable(new SvelteMap<string, IPaneSerialized[]>());
export const layoutState: { tree: VizSubPanel[]; } = $state({
    tree: []
});
export const allTabs = writable(new SvelteMap<string, VizTab[]>());

export const getAllSubPanels = () => {
    let subPanels = layoutState.tree.flat();

    // I hate this so much
    if (subPanels.flatMap((panel) => panel.childs).length > 0) {
        subPanels = subPanels?.concat(subPanels.flatMap((panel) => panel.childs?.subPanel ?? []));

        if (subPanels.flatMap((panel) => panel.childs?.parentSubPanel).length > 0) {
            subPanels = subPanels.concat(
                subPanels
                    .flatMap((panel) => panel.childs?.parentSubPanel ?? [])
                    .filter((pane): pane is VizSubPanel => !!pane && typeof pane === "object" && "id" in pane)
            );
        }
    }

    return subPanels;

};