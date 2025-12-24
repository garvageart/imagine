import { openDB, type IDBPDatabase } from "idb";
import { type Image, type Collection } from "$lib/api";
import { DateTime } from "luxon";
import { initDB } from "$lib/db/client";

export type Asset = Image | Collection;

export interface DateRange {
    after?: string;
    before?: string;
}

export interface NumberRange {
    min?: number;
    max?: number;
}

export interface ImageFilters {
    rating: number | null;
    date: DateRange;
    iso: NumberRange;
    fStop: NumberRange;
    shutterSpeed: NumberRange;
    focalLength: NumberRange;
    tags: string[];
    camera: string[];
    lens: string[];
}

export interface ImageFacets {
    cameras: Map<string, number>;
    lenses: Map<string, number>;
    tags: Map<string, number>;
    iso: { min: number; max: number; };
    fStop: { min: number; max: number; };
    shutterSpeed: { min: number; max: number; };
    focalLength: { min: number; max: number; };
}

export interface CollectionFilters { }

export interface CollectionFacets { }

interface SavedScopeState {
    criteria: any;
    uiState: { expanded: Record<string, boolean>; };
}

interface SavedFilterState {
    keepFilters?: boolean;
    scopes?: Record<string, SavedScopeState>;
}

const DEFAULT_IMAGE_FILTERS: ImageFilters = {
    rating: null,
    date: {},
    iso: {},
    fStop: {},
    shutterSpeed: {},
    focalLength: {},
    tags: [],
    camera: [],
    lens: []
};

const DEFAULT_IMAGE_UI_STATE = {
    expanded: {
        rating: true,
        camera: false,
        lens: false,
        tags: true,
        tech: false,
        date: false
    }
};

const DEFAULT_COLLECTION_FILTERS: CollectionFilters = {};
const DEFAULT_COLLECTION_UI_STATE = { expanded: {} };

const DB_KEY = "filter-manager-state";
const SETTINGS_STORE = "settings";

export class FilterScope<F, A extends Asset> {
    type: 'images' | 'collections';
    criteria: F = $state() as F;
    facets: any = $state();
    uiState: { expanded: Record<string, boolean>; } = $state() as any;

    constructor(type: 'images' | 'collections', defaultFilters: F, defaultUiState: { expanded: Record<string, boolean>; }) {
        this.type = type;
        this.criteria = { ...defaultFilters };
        this.uiState = { ...defaultUiState };

        if (type === 'images') {
            this.facets = {
                cameras: new Map(),
                lenses: new Map(),
                tags: new Map(),
                iso: { min: 0, max: 0 },
                fStop: { min: 0, max: 0 },
                shutterSpeed: { min: 0, max: 0 },
                focalLength: { min: 0, max: 0 }
            };
        } else {
            this.facets = {};
        }
    }

    resetCriteria() {
        if (this.type === 'images') {
            Object.assign(this.criteria as ImageFilters, DEFAULT_IMAGE_FILTERS);
        } else if (this.type === 'collections') {
            Object.assign(this.criteria as CollectionFilters, DEFAULT_COLLECTION_FILTERS);
        }
    }

    isImageScope(): this is FilterScope<ImageFilters, Image> {
        return this.type === 'images';
    }

    isCollectionScope(): this is FilterScope<CollectionFilters, Collection> {
        return this.type === 'collections';
    }

    private parseFStop(val: string | undefined): number | undefined {
        if (!val) {
            return undefined;
        }
        const clean = val.toLowerCase().replace("f/", "").replace("f", "").trim();
        const num = parseFloat(clean);
        return isNaN(num) ? undefined : num;
    }

    private parseISO(val: string | undefined): number | undefined {
        if (!val) {
            return undefined;
        }
        const num = parseFloat(val);
        return isNaN(num) ? undefined : num;
    }

    private parseFocalLength(val: string | undefined): number | undefined {
        if (!val) {
            return undefined;
        }
        const clean = val.toLowerCase().replace("mm", "").trim();
        const num = parseFloat(clean);
        return isNaN(num) ? undefined : num;
    }

    private parseShutterSpeed(val: string | undefined): number | undefined {
        if (!val) {
            return undefined;
        }
        if (val.includes("/")) {
            const [num, den] = val.split("/");
            const n = parseFloat(num);
            const d = parseFloat(den);
            if (!isNaN(n) && !isNaN(d) && d !== 0) {
                return n / d;
            }
        }
        const num = parseFloat(val);
        return isNaN(num) ? undefined : num;
    }

    updateFacets(items: A[]) {
        if (this.isImageScope()) {
            const cameras = new Map<string, number>();
            const lenses = new Map<string, number>();
            const tags = new Map<string, number>();

            let minIso = Infinity, maxIso = -Infinity;
            let minF = Infinity, maxF = -Infinity;
            let minSS = Infinity, maxSS = -Infinity;
            let minFL = Infinity, maxFL = -Infinity;

            for (const item of items) {
                if ('image_metadata' in item) {
                    const img = item as Image;
                    const exif = img.exif || {};

                    if (exif.model) {
                        const model = exif.model.trim();
                        cameras.set(model, (cameras.get(model) || 0) + 1);
                    }

                    if (exif.lens_model) {
                        const lens = exif.lens_model.trim();
                        lenses.set(lens, (lenses.get(lens) || 0) + 1);
                    }

                    if (img.image_metadata?.keywords) {
                        for (const tag of img.image_metadata.keywords) {
                            tags.set(tag, (tags.get(tag) || 0) + 1);
                        }
                    }

                    const iso = this.parseISO(exif.iso);
                    if (iso !== undefined) {
                        if (iso < minIso) { minIso = iso; }
                        if (iso > maxIso) { maxIso = iso; }
                    }

                    const f = this.parseFStop(exif.aperture ?? exif.f_number);
                    if (f !== undefined) {
                        if (f < minF) { minF = f; }
                        if (f > maxF) { maxF = f; }
                    }

                    const ss = this.parseShutterSpeed(exif.exposure_time);
                    if (ss !== undefined) {
                        if (ss < minSS) { minSS = ss; }
                        if (ss > maxSS) { maxSS = ss; }
                    }

                    const fl = this.parseFocalLength(exif.focal_length);
                    if (fl !== undefined) {
                        if (fl < minFL) { minFL = fl; }
                        if (fl > maxFL) { maxFL = fl; }
                    }
                }
            }

            this.facets = {
                cameras,
                lenses,
                tags,
                iso: { min: minIso === Infinity ? 0 : minIso, max: maxIso === -Infinity ? 12800 : maxIso },
                fStop: { min: minF === Infinity ? 0 : minF, max: maxF === -Infinity ? 32 : maxF },
                shutterSpeed: { min: minSS === Infinity ? 0 : minSS, max: maxSS === -Infinity ? 30 : maxSS },
                focalLength: { min: minFL === Infinity ? 0 : minFL, max: maxFL === -Infinity ? 1000 : maxFL }
            };
        } else if (this.isCollectionScope()) {
            this.facets = {};
        }
    }

    apply(items: A[]): A[] {
        if (this.isImageScope()) {
            const criteria = this.criteria;
            const hasActiveFilters =
                criteria.rating !== null ||
                criteria.date.after || criteria.date.before ||
                criteria.iso.min !== undefined || criteria.iso.max !== undefined ||
                criteria.fStop.min !== undefined || criteria.fStop.max !== undefined ||
                criteria.shutterSpeed.min !== undefined || criteria.shutterSpeed.max !== undefined ||
                criteria.focalLength.min !== undefined || criteria.focalLength.max !== undefined ||
                criteria.tags.length > 0 ||
                criteria.camera.length > 0 ||
                criteria.lens.length > 0;

            if (!hasActiveFilters) {
                return items;
            }

            return items.filter(item => {
                if (!('image_metadata' in item)) {
                    return false;
                }
                
                const img = item as Image;
                const exif = img.exif || {};

                if (criteria.rating !== null) {
                    const imgRating = img.image_metadata?.rating ?? 0;
                    if (imgRating < criteria.rating) {
                        return false;
                    }
                }

                if (criteria.date.after || criteria.date.before) {
                    const takenAt = img.taken_at ? DateTime.fromISO(img.taken_at) : null;
                    if (!takenAt || !takenAt.isValid) {
                        return false;
                    }

                    if (criteria.date.after) {
                        const after = DateTime.fromISO(criteria.date.after);
                        if (takenAt < after) {
                            return false;
                        }
                    }
                    if (criteria.date.before) {
                        const before = DateTime.fromISO(criteria.date.before);
                        if (takenAt > before) {
                            return false;
                        }
                    }
                }

                if (criteria.iso.min !== undefined || criteria.iso.max !== undefined) {
                    const val = this.parseISO(exif.iso);
                    if (val === undefined) {
                        return false;
                    }
                    if (criteria.iso.min !== undefined && val < criteria.iso.min) {
                        return false;
                    }
                    if (criteria.iso.max !== undefined && val > criteria.iso.max) {
                        return false;
                    }
                }

                if (criteria.fStop.min !== undefined || criteria.fStop.max !== undefined) {
                    const val = this.parseFStop(exif.aperture ?? exif.f_number);
                    if (val === undefined) {
                        return false;
                    }
                    if (criteria.fStop.min !== undefined && val < criteria.fStop.min) {
                        return false;
                    }
                    if (criteria.fStop.max !== undefined && val > criteria.fStop.max) {
                        return false;
                    }
                }

                if (criteria.shutterSpeed.min !== undefined || criteria.shutterSpeed.max !== undefined) {
                    const val = this.parseShutterSpeed(exif.exposure_time);
                    if (val === undefined) {
                        return false;
                    }
                    if (criteria.shutterSpeed.min !== undefined && val < criteria.shutterSpeed.min) {
                        return false;
                    }
                    if (criteria.shutterSpeed.max !== undefined && val > criteria.shutterSpeed.max) {
                        return false;
                    }
                }

                if (criteria.focalLength.min !== undefined || criteria.focalLength.max !== undefined) {
                    const val = this.parseFocalLength(exif.focal_length);
                    if (val === undefined) {
                        return false;
                    }
                    if (criteria.focalLength.min !== undefined && val < criteria.focalLength.min) {
                        return false;
                    }
                    if (criteria.focalLength.max !== undefined && val > criteria.focalLength.max) {
                        return false;
                    }
                }

                if (criteria.tags.length > 0) {
                    const imgTags = img.image_metadata?.keywords || [];
                    const hasAll = criteria.tags.every(t => imgTags.includes(t));
                    if (!hasAll) {
                        return false;
                    }
                }

                if (criteria.camera.length > 0) {
                    const model = (exif.model ?? "").trim();
                    if (!criteria.camera.includes(model)) {
                        return false;
                    }
                }

                if (criteria.lens.length > 0) {
                    const lens = (exif.lens_model ?? "").trim();
                    if (!criteria.lens.includes(lens)) {
                        return false;
                    }
                }
                return true;
            });
        } else if (this.isCollectionScope()) {
            return items; // No collection filters yet
        }
        return items;
    }
}

class FilterManager {
    scopes: Map<string, FilterScope<any, any>> = $state(new Map());
    activeScopeType: 'images' | 'collections' | null = $state('images');
    keepFilters: boolean = $state(false);

    private dbPromise: Promise<IDBPDatabase> | null = null;
    private isInitialized = false;

    constructor() {
        this.scopes.set('images', new FilterScope('images', DEFAULT_IMAGE_FILTERS, DEFAULT_IMAGE_UI_STATE));
        this.scopes.set('collections', new FilterScope('collections', DEFAULT_COLLECTION_FILTERS, DEFAULT_COLLECTION_UI_STATE));

        if (typeof window !== "undefined") {
            this.init();
        }
    }

    async init() {
        if (this.isInitialized) {
            return;
        }

        this.dbPromise = initDB();
        const db = await this.dbPromise;
        const savedState = (await db.get(SETTINGS_STORE, DB_KEY)) as SavedFilterState | undefined;

        if (savedState) {
            this.keepFilters = savedState.keepFilters ?? false;
            if (savedState.scopes) {
                for (const [key, scopeState] of Object.entries(savedState.scopes)) {
                    const scope = this.scopes.get(key);
                    if (scope) {
                        Object.assign(scope.criteria, scopeState.criteria);
                        Object.assign(scope.uiState, scopeState.uiState);
                    }
                }
            }
        }

        this.isInitialized = true;
    }

    get activeScope(): FilterScope<any, any> | undefined {
        if (this.activeScopeType) {
            return this.scopes.get(this.activeScopeType);
        }
        return undefined;
    }

    getScope(type: 'images' | 'collections'): FilterScope<any, any> | undefined {
        return this.scopes.get(type);
    }

    setActiveScopeType(type: 'images' | 'collections' | null) {
        this.activeScopeType = type;
    }

    toggleKeepFilters() {
        this.keepFilters = !this.keepFilters;
        this.save();
    }

    resetActiveScope(force = false) {
        if (this.keepFilters && !force) {
            return;
        }
        if (this.activeScope) {
            this.activeScope.resetCriteria();
            this.save();
        }
    }

    async save() {
        if (!this.dbPromise) {
            return;
        }

        const stateToSave = {
            keepFilters: this.keepFilters,
            scopes: {} as Record<string, any>
        };

        for (const [key, scope] of this.scopes.entries()) {
            stateToSave.scopes[key] = {
                criteria: $state.snapshot(scope.criteria),
                uiState: $state.snapshot(scope.uiState),
            };
        }

        try {
            const db = await this.dbPromise;
            await db.put(SETTINGS_STORE, stateToSave, DB_KEY);
        } catch (e) {
            console.error("Failed to save filter state", e);
        }
    }

    apply(items: Asset[]): Asset[] {
        if (!this.activeScope) {
            return items;
        }
        return this.activeScope.apply(items);
    }
}

export const filterManager = new FilterManager();
