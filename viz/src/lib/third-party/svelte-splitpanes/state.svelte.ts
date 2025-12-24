import { SvelteMap } from "svelte/reactivity";
import type { IPaneSerialized, ITree } from ".";
import { writable } from "svelte/store";
import type VizView from "$lib/views/views.svelte";
import type VizSubPanelData from "$lib/layouts/subpanel.svelte";

// this might cause bugs idk
export const allSplitpanes = writable(new SvelteMap<string, IPaneSerialized[]>());
export const layoutState: { tree: VizSubPanelData[]; } = $state({
    tree: []
});
export const allTabs = writable(new SvelteMap<string, VizView[]>());
export const layoutTree = $state({ locked: false, activeContentId: undefined }) as ITree;