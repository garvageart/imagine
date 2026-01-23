import type { AssetSort, AssetGridView } from "$lib/types/asset";
import type { UploadImage } from "$lib/upload/asset.svelte";
import { type User, type SystemStatusResponse, type ImageAsset, type Collection, updateUserSetting, type UserSetting } from "$lib/api";
import { VizLocalStorage, VizCookieStorage } from "$lib/utils/misc";
import { MediaQuery } from "svelte/reactivity";
import type { MenuItem } from "$lib/context-menu/types";
import { SettingNames } from "$lib/components/settings/names";
import { page } from "$app/state";

// Types
interface UserState {
    data: User | null;
    loading: boolean;
    fetched: boolean;
    error?: string | null;
    isAdmin: boolean;
    connectionError?: boolean;
    settings?: UserSetting[] | null;
}

export let system = $state<{
    data: SystemStatusResponse | null;
    loading: boolean;
    fetched: boolean;
    error: string | null;
}>({
    data: null,
    loading: false,
    fetched: false,
    error: null
});

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

export let showHeader = $state(true);

export let user = $state<UserState>({
    data: null,
    loading: false,
    fetched: false,
    error: null,
    isAdmin: false,
    connectionError: false,
    settings: null
});

export let search = $state({
    loading: false,
    executed: false,
    data: {
        collections: {
            data: [] as Collection[]
        },
        images: {
            data: [] as ImageAsset[]
        }
    },
    value: "",
    enableHomePageSearch: false,
    element: undefined as HTMLInputElement | undefined
});

export let modal = $state({
    show: false
});

export let lightbox = $state({
    show: false
});

export function isLayoutPage() {
    return page.url.pathname === "/";
}

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
    private defaults: AssetSort = {
        display: "cover",
        group: {
            by: "year",
            order: "asc",
        },
        by: "name",
        order: "asc",
    } as const;

    value: AssetSort = $state(
        this.storage.get() ?? this.defaults
    );

    save() {
        this.storage.set(this.value);
    }
}

export const sortState = new SortState();
export let sort = sortState.value;

export class TableColumnSettings {
    storage = new VizLocalStorage<string[]>('tableColumnSettings');
    // Default columns (Preview is hardcoded in AssetGrid, so we just track the dynamic ones)
    value: string[] = $state(this.storage.get() ?? ['name', 'created_at']);

    toggle(column: string) {
        if (this.value.includes(column)) {
            this.value = this.value.filter(c => c !== column);
        } else {
            this.value = [...this.value, column];
        }
        this.storage.set(this.value);
    }

    set(columns: string[]) {
        this.value = columns;
        this.storage.set(this.value);
    }
}

export const tableColumnSettings = new TableColumnSettings();

class ViewSettingsState {
    storage = new VizLocalStorage<AssetGridView>('viewSettings');
    current: AssetGridView = $state(this.storage.get() ?? 'grid');
    displayOptions: MenuItem[] = [
        { id: "view-grid", label: "Grid" },
        { id: "view-list", label: "List" },
        { id: "view-thumbnails", label: "Thumbnails" }
    ];

    setView(view: AssetGridView) {
        this.current = view;
        this.storage.set(view);
    }
}

export const viewSettings = new ViewSettingsState();

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
    ls = new VizLocalStorage<ThemeOption>(SettingNames.Theme);
    prefLs = new VizLocalStorage<ThemeOption>('preferred_theme');

    value: ThemeOption = $state(this.getInitialTheme());
    preferredTheme: ThemeOption = $state(this.getInitialPreference());

    private media = typeof window !== 'undefined' ? new MediaQuery('prefers-color-scheme: dark') : undefined;
    systemPref = $derived((this.media?.current ?? false) ? 'dark' : 'light');

    resolved = $derived(this.value === 'system' ? this.systemPref : this.value);
    resolvedPreference = $derived(this.preferredTheme === 'system' ? this.systemPref : this.preferredTheme);

    constructor() {
    }

    private getInitialTheme(): ThemeOption {
        // Preference: 1. LocalStorage, 2. Default to system
        const storedTheme = this.ls.get();
        if (storedTheme) {
            return storedTheme;
        }
        return 'system';
    }

    private getInitialPreference(): ThemeOption {
        const stored = this.prefLs.get();
        if (stored) {
            return stored;
        }

        return 'system';
    }

    async setPreferredTheme(theme: ThemeOption) {
        this.preferredTheme = theme;
        this.prefLs.set(theme);
        try {
            await updateUserSetting(SettingNames.Theme, { value: theme });
        } catch (e) {
            console.error("Failed to sync theme preference", e);
        }
    }

    toggle() {
        // If we are NOT at our default preference, return to it.
        if (this.resolved !== this.resolvedPreference) {
            this.value = this.preferredTheme;
            return;
        }

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