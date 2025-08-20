import type { AssetSort } from "$lib/types/asset";
import type { Collection, IImageObjectData } from "$lib/types/images";
import { cookieMethods } from "$lib/utils/cookie";
import { writable } from "svelte/store";

export let login = $state({
    state: cookieMethods.get("imag-state")
});

export let sidebar = $state({
    open: false
});

export let showHeader = writable(true);

export let search = $state({
    loading: false,
    executed: false,
    data: {
        collections: [] as unknown as Collection[],
        images: [] as unknown as IImageObjectData[]
    },
    value: "",
    element: undefined as unknown as HTMLInputElement | undefined
});

export let modal = $state({
    show: false
});

export let lightbox = $state({
    show: false
});

/**
 * @todo Get sort options from saved settings by these are the defaults for now
 */
export let sort: AssetSort = $state({
    display: "cover",
    group: {
        by: "year",
        order: "asc",
    },
    by: "name",
    order: "asc",
});