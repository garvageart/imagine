import type { ImageObjectData } from "$lib/entities/image";
import type CollectionData from "$lib/entities/collection";
import type { AssetSort } from "$lib/types/asset";
import type { UploadImage } from "$lib/upload/asset.svelte";
import { cookieMethods } from "$lib/utils/cookie";
import { writable } from "svelte/store";
import { type User, getCurrentUser, logout } from "$lib/api/client";
import { VizLocalStorage } from "$lib/utils/misc";
import { goto } from "$app/navigation";

export let login = $state({
    state: cookieMethods.get("imag-state")
});

export let sidebar = $state({
    open: false
});

export let showHeader = writable(true);

export let user = $state<{ data: User | null; loading: boolean; fetched: boolean; error?: string | null; isAdmin: boolean; }>({
    data: null,
    loading: false,
    fetched: false,
    error: null,
    isAdmin: false
});

export let search = $state({
    loading: false,
    executed: false,
    data: {
        collections: {
            data: [] as CollectionData[]
        },
        images: {
            data: [] as ImageObjectData[]
        }
    },
    value: "",
    enableHomePageSearch: false,
    element: undefined as unknown as HTMLInputElement | undefined
});

export let modal = $state({
    show: false
});

export let lightbox = $state({
    show: false
});

class SortState {
    private storage = new VizLocalStorage<AssetSort>("sort");

    value: AssetSort = $state(
        this.storage.get() ?? {
            display: "cover",
            group: {
                by: "year",
                order: "asc",
            },
            by: "name",
            order: "asc",
        }
    );

    constructor() { }

    save() {
        this.storage.set(this.value);
    }
}

export const sortState = new SortState();
export let sort = sortState.value;

export let upload = $state({
    files: [] as UploadImage[],
    concurrency: 2,
    stats: {
        errors: 0,
        duplicates: 0,
        success: 0,
        total: 0
    }
});

export let continuePath = $state<string | null>(null);