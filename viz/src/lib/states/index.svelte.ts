import type { ImageObjectData } from "$lib/entities/image";
import type CollectionData from "$lib/entities/collection";
import type { AssetSort } from "$lib/types/asset";
import type { UploadImage } from "$lib/upload/asset.svelte";
import { writable } from "svelte/store";
import { type User } from "$lib/api/client";
import { VizLocalStorage, VizCookieStorage } from "$lib/utils/misc";

class LoginState {
    storage: VizCookieStorage<string>;
    value: string | null = $state(null);
    constructor() {
        // shitty hack
        this.storage = new VizCookieStorage<string>('state');
        this.storage.prefix = "imag";
        this.value = this.storage.get();
    }
}
const loginState = new LoginState();
export let login = $state({
    state: loginState.value
});

export let sidebar = $state({
    open: false
});

export let showHeader = writable(true);

export let user = $state<{ data: User | null; loading: boolean; fetched: boolean; error?: string | null; isAdmin: boolean; connectionError?: boolean; }>({
    data: null,
    loading: false,
    fetched: false,
    error: null,
    isAdmin: false,
    connectionError: false
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

// eventually this will move to a different page with a different way of enabling, this is just temporary
class DebugState {
    storage = new VizLocalStorage<boolean>('debugMode');
    value: boolean = $state(this.storage.get() ?? false);

    toggle() {
        this.value = !this.value;
    }
}
export const debugState = new DebugState();
export let debugMode = debugState.value;

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

export type ThemeOption = 'light' | 'dark' | 'system';

class ThemeState {
    ls = new VizLocalStorage<ThemeOption>('theme');

    value: ThemeOption = $state(this.getInitialTheme());
    systemPref: 'light' | 'dark' = $state('light');

    resolved = $derived(this.value === 'system' ? this.systemPref : this.value);

    constructor() {
        if (typeof window !== 'undefined') {
            const media = window.matchMedia('(prefers-color-scheme: dark)');
            this.systemPref = media.matches ? 'dark' : 'light';
            media.addEventListener('change', (e) => {
                this.systemPref = e.matches ? 'dark' : 'light';
            });
        }
    }

    private getInitialTheme(): ThemeOption {
        // Preference: 1. LocalStorage, 2. Default to system
        const storedTheme = this.ls.get();
        if (storedTheme) {
            return storedTheme;
        }
        return 'system';
    }

    toggle() {
        const currentResolved = this.resolved;
        const targetResolved = currentResolved === 'dark' ? 'light' : 'dark';

        // If 'system' preference matches the target visual state, prefer 'system'.
        // This ensures we default to system behavior when possible, while guaranteeing a visual switch.
        if (this.systemPref === targetResolved) {
            this.value = 'system';
        } else {
            this.value = targetResolved;
        }
    }
}

export const themeState = new ThemeState();
export function getTheme() {
    return themeState.resolved;
}

export function toggleTheme() {
    themeState.toggle();
}