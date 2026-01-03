import type { Component } from "svelte";
import { preloadData, invalidateAll } from "$app/navigation";
import { debugMode } from "$lib/states/index.svelte";
import type { MenuItem } from "$lib/context-menu/types";
import { sleep } from "$lib/utils/misc";

// usually this would be bad but the app is client only
// and doesn't share state with anyone i guess??
let idCount = 1;

/**
 * A global version counter for view invalidation.
 * Incrementing this will cause all active VizView instances to re-evaluate
 * their derivedViewData, effectively triggering a refresh of their content.
 */
export const invalidationState = $state({ version: 0 });

/**
 * Triggers a global invalidation of all VizView instances and SvelteKit load functions.
 * Use this instead of `invalidateAll()` when you want to ensure background panels
 * also refresh their data (e.g., after uploading images or modifying collections).
 */
export async function invalidateViz(opts?: { delay?: number }) {
    if (opts?.delay) {
        await sleep(opts.delay);
    }
    invalidationState.version += 1;
    return await invalidateAll();
}

export interface SerializedVizView {
    name: string;
    opticalCenterFix?: number;
    id: number;
    isActive: boolean;
    locked?: boolean;
    path?: string;
}

class VizView<C extends Component<any, any, any> = Component<any, any, any>> {
    name = $state<string>("");
    opticalCenterFix = $state<number | undefined>(undefined);
    component: C;
    id = $state<number>(0);
    isActive = $state<boolean>(false);
    locked = $state<boolean>(false);
    public viewData = $state<{
        type: "loaded";
        status: number;
        data: C extends Component<infer P, any, any> ? (P extends { data: infer D; } ? D : any) : any;
    } | undefined>(undefined);
    path = $state<string | undefined>(undefined);
    menuItems?: MenuItem[];

    constructor(opts: {
        name: string;
        component: C;
        opticalCenterFix?: number;
        path?: string;
        id?: number;
        isActive?: boolean;
        locked?: boolean;
        menuItems?: MenuItem[];
    }) {
        this.name = opts.name;
        this.component = opts.component;
        this.path = opts.path;
        this.opticalCenterFix = opts.opticalCenterFix ?? 0.5;
        
        if (opts.id !== undefined) {
            this.id = opts.id;
            // Update the global counter to ensure subsequent auto-generated IDs
            // do not collide with IDs from hydrated/serialized views.
            if (this.id >= idCount) {
                idCount = this.id + 1;
            }
        } else {
            this.id = idCount++;
        }
        
        this.isActive = opts.isActive ?? false;
        this.locked = opts.locked ?? false;
        this.menuItems = opts.menuItems;

        if (this.path) {
            // Preload data if path is available on construction
            this.getComponentData();
        }
    }

    setActive(active: boolean) {
        this.isActive = active;
    }

    /**
     * Resets the view state, clearing loaded data and resetting the name.
     * Useful when switching between different instances of the same component type (e.g., different collections).
     */
    reset(newName?: string) {
        this.viewData = undefined;
        if (newName) {
            this.name = newName;
        }
    }

    async getComponentData(): Promise<void | { type: "loaded"; status: number; data: any; }> {
        // Register dependency on invalidationVersion first so even views without path
        // will re-evaluate derivedViewData (and thus re-render) when invalidation happens.
        const version = invalidationState.version;

        if (!this.path) {
            return;
        }

        if (debugMode) {
            console.log(`Loading data ${this.path}`);
        }

        const sep = this.path.includes('?') ? '&' : '?';
        const urlWithCacheBust = `${this.path}${sep}invalidation=${version}`;

        const result = await preloadData(urlWithCacheBust);
        if (result.type === 'loaded' && result.status === 200) {
            this.viewData = result as any;
            return result;
        }
    }

    derivedViewData = $derived.by(() => {
        // Return a promise that resolves with the component data
        return this.getComponentData();
    });

    /**
     * Serializes the view to a plain object for localStorage
     * Excludes component and viewData which cannot be serialized
     */
    toJSON(): SerializedVizView {
        return {
            name: this.name,
            opticalCenterFix: this.opticalCenterFix,
            id: this.id,
            isActive: this.isActive,
            locked: this.locked,
            path: this.path
        };
    }

    /**
     * Creates a VizView instance from serialized data and a component lookup
     * @param serialized The serialized view data from localStorage
     * @param component The component to use for this view
     */
    static fromJSON<C extends Component<any, any, any> = Component<any, any, any>>(
        serialized: SerializedVizView,
        component: C
    ): VizView<C> {
        return new VizView({
            name: serialized.name,
            component: component,
            opticalCenterFix: serialized.opticalCenterFix,
            path: serialized.path,
            id: serialized.id,
            isActive: serialized.isActive,
            locked: serialized.locked ?? false
        });
    }
}

export default VizView;