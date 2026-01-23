import { SvelteSet } from "svelte/reactivity";

export enum SelectionScopeNames {
    DEFAULT = "default",
    PHOTOS_DEFAULT = "photos-default",
    PHOTOS_MAIN = "photos-main",
    COLLECTIONS_MAIN = "collections-main",
    COLLECTION_PREFIX = "collection-",
    SEARCH_IMAGES = "search-images",
    SEARCH_COLLECTIONS = "search-collections"
}

export class SelectionScope<T extends { uid: string; } = any> {
    selected = $state(new SvelteSet<T>());
    active = $state<T | undefined>(undefined);
    source = $state<T[]>([]); // All items available in this scope
    id: string;

    constructor(id: string = SelectionScopeNames.DEFAULT) {
        this.id = id;
    }

    setSource(items: T[]) {
        this.source = items;
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

    /**
     * Updates an item in a given source array and also updates the selection if the item is selected.
     * @param updatedItem The new item data.
     * @param sourceArray The original, mutable source array of items (e.g. a $state array).
     */
    updateItem(updatedItem: T, sourceArray: T[]) {
        const idx = sourceArray.findIndex(i => i.uid === updatedItem.uid);
        if (idx !== -1) {
            // Update source array (keeping reference if possible is good, but here we replace)
            sourceArray[idx] = updatedItem;
        }

        // Find and update in selected set by UID
        let itemInSet: T | undefined;
        for (const item of this.selected) {
            if (item.uid === updatedItem.uid) {
                itemInSet = item;
                break;
            }
        }

        if (itemInSet) {
            this.remove(itemInSet);
            this.add(updatedItem);
        }

        if (this.active?.uid === updatedItem.uid) {
            this.active = updatedItem;
        }
    }

    selectNext() {
        if (!this.active || this.source.length === 0) return false;
        const idx = this.source.findIndex(i => i.uid === this.active!.uid);
        if (idx === -1 || idx === this.source.length - 1) return false;
        
        this.select(this.source[idx + 1]);
        return true;
    }

    selectPrevious() {
        if (!this.active || this.source.length === 0) return false;
        const idx = this.source.findIndex(i => i.uid === this.active!.uid);
        if (idx === -1 || idx === 0) return false;

        this.select(this.source[idx - 1]);
        return true;
    }
}

export class SelectionManager {
    scopes = new Map<string, SelectionScope>();
    activeScopeId = $state<string | null>(null);

    // A default global scope for simple use cases
    global = new SelectionScope("global");

    constructor() { }

    get activeScope() {
        if (!this.activeScopeId) {
            return this.global;
        }
        return this.scopes.get(this.activeScopeId) ?? this.global;
    }

    /**
     * The primary item focused in the active scope. 
     * This is what the Inspector panel should primarily display.
     */
    get focusedItem() {
        return this.activeScope.active;
    }

    /**
     * All items selected in the active scope.
     */
    get selectedItems() {
        return Array.from(this.activeScope.selected);
    }

    setActive(scopeId: string) {
        this.activeScopeId = scopeId;
    }

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
