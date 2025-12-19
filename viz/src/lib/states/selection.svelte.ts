import { SvelteMap, SvelteSet } from "svelte/reactivity";

export class SelectionScope<T extends { uid: string; } = any> {
    selected = $state(new SvelteSet<T>());
    active = $state<T | undefined>(undefined);
    id: string;

    constructor(id: string = "default") {
        this.id = id;
    }

    add(item: T) {
        this.selected.add(item);
    }

    remove(item: T) {
        this.selected.delete(item);
    }

    has(item: T) {
        return this.selected.has(item);
    }

    clear() {
        this.selected.clear();
        this.active = undefined;
    }

    toggle(item: T) {
        if (this.selected.has(item)) {
            this.selected.delete(item);
        } else {
            this.selected.add(item);
        }
    }

    /**
     * Selects a single item, clearing previous selection.
     * Sets it as the active (primary) selection.
     */
    select(item: T) {
        this.selected.clear();
        this.selected.add(item);
        this.active = item;
    }

    /**
     * Selects multiple items, clearing previous selection.
     * Does NOT set a single active item (unless items has length 1, but usually active implies user focus)
     */
    selectMultiple(items: Iterable<T>) {
        this.selected.clear();
        for (const item of items) {
            this.selected.add(item);
        }
    }

    /**
     * Adds multiple items to current selection
     */
    addMultiple(items: Iterable<T>) {
        for (const item of items) {
            this.selected.add(item);
        }
    }
}

export class SelectionManager {
    scopes = new SvelteMap<string, SelectionScope>();

    // A default global scope for simple use cases
    global = new SelectionScope("global");

    constructor() { }

    getScope<T extends { uid: string; } = any>(scopeId: string): SelectionScope<T> {
        if (!this.scopes.has(scopeId)) {
            this.scopes.set(scopeId, new SelectionScope<T>(scopeId));
        }
        return this.scopes.get(scopeId) as SelectionScope<T>;
    }

    removeScope(scopeId: string) {
        this.scopes.delete(scopeId);
    }

    /**
     * aggregated helper: get all selected items across all scopes
     * (Useful for global filtering/actions)
     */
    getAllSelectedItems<T extends { uid: string; } = any>(): T[] {
        const all: T[] = [];
        // Include default global
        all.push(...this.global.selected as unknown as T[]);

        for (const scope of this.scopes.values()) {
            all.push(...scope.selected as unknown as T[]);
        }
        return all;
    }
}

export const selectionManager = new SelectionManager();
