import { SvelteMap } from "svelte/reactivity";
import type { IPaneSerialized } from ".";
import { writable, type Writable } from "svelte/store";
import type { VizSubPanel, VizTab } from "$lib/components/panels/SubPanel.svelte";

// this might cause bugs idk
export const allSplitpanes = writable(new SvelteMap<string, IPaneSerialized[]>());
export const layoutState: Writable<VizSubPanel[]> = writable([]);
export const allTabs = writable(new SvelteMap<string, VizTab[]>([]));