import { SvelteSet } from "svelte/reactivity";
import { writable } from 'svelte/store';

// public store that components can subscribe to
export const appReady = writable(false);

// internal tracking of registered promises
const pending = new SvelteSet<Promise<any>>();
let resolved = false;
let fallbackTimer: ReturnType<typeof setTimeout> | null = null;
const GLOBAL_FALLBACK = 15000; // ms - max time before forcing ready

function checkPending() {
    if (resolved) {
        return;
    }

    if (pending.size === 0) {
        finalizeReady();
    }
}

// register a promise or array of promises. The app will become ready once
// all registered promises settle (resolve or reject). A safety timeout will
// force readiness after the given ms to avoid indefinite blocking.
export function registerReady(p: Promise<any> | Promise<any>[]) {
    if (resolved) {
        return;
    }

    const arr = Array.isArray(p) ? p : [p];
    arr.forEach((pr) => {
        const wrapper = pr.finally(() => {
            pending.delete(wrapper);
            checkPending();
        });
        pending.add(wrapper);
    });

    if (!fallbackTimer) {
        fallbackTimer = setTimeout(() => {
            if (!resolved) {
                finalizeReady();
            }

            fallbackTimer = null;
        }, GLOBAL_FALLBACK);
    }

    if (pending.size === 0) {
        checkPending();
    }
}

export function markReady() {
    if (resolved) {
        return;
    }

    finalizeReady();
}

function finalizeReady() {
    if (resolved) {
        return;
    }

    resolved = true;
    pending.clear();
    appReady.set(true);

    try {
        if (typeof document !== 'undefined') {
            const el = document.getElementById('app-splash');
            if (el) {
                el.classList.add('app-splash--fade');
                const cleanup = () => {
                    el.remove();
                    document.getElementById("splash-styles")?.remove();
                };
                el.addEventListener('transitionend', cleanup, { once: true });
                // fallback removal in case transitionend doesn't fire
                setTimeout(() => { if (el.parentNode) el.remove(); }, 700);
            }
        }
    } catch (e) {
        // ignore DOM removal errors
    }

    if (fallbackTimer) {
        clearTimeout(fallbackTimer);
        fallbackTimer = null;
    }
}

if (typeof document !== 'undefined' && 'fonts' in document) {
    try {
        registerReady(document.fonts.ready);
    } catch (e) {
    }
}

// Explicitly try to load Material Symbols / Material Icons families so the
// splash waits until those icon fonts are usable. This helps avoid the
// ligature flash/width-stretch where the literal icon name (e.g. "upload")
// briefly appears before the font maps it to the glyph.
if (typeof document !== 'undefined' && 'fonts' in document) {
    try {
        const families = [
            'Material Symbols Outlined',
            'Material Symbols Sharp',
            'Material Symbols Rounded',
            'Material Symbols Filled',
            'Material Icons'
        ];

        const loads = families.map((f) => document.fonts.load(`1em "${f}"`).catch(() => null));
        // Register as a combined readiness checkpoint (don't block indefinitely)
        registerReady(Promise.allSettled(loads));
    } catch (e) {
        // swallow - optional best-effort
    }
}

if (typeof window !== 'undefined') {
    const winLoad = new Promise<void>((res) => {
        if (document.readyState === 'complete') {
            res();
        } else {
            window.addEventListener('load', () => res(), { once: true });
        }
    });

    registerReady(winLoad);
}

export default appReady;
