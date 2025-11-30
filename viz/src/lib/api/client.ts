/**
 * Configured oazapfts API client
 * Auto-generated functions with proper base URL and SvelteKit fetch integration
 */
import { defaults, servers as generatedServers } from "./client.gen";
import { dev } from "$app/environment";

// Compute a static API base URL constant at module load time.
// This is exported as `API_BASE_URL` and also applied to `defaults.baseUrl`.
// Priority:
// 1) Build-time injected `__RUNTIME_CONFIG__.servers['api']` -> compose `http://{host}:{port}`
// 2) In dev: use `http://localhost:7770`
// 3) Production / proxy: use any relative server string from generated `servers` (e.g. '/api'), default '/api'
function parseRuntimeConfig(): any | undefined {
    try {
        const maybe = (globalThis as any).__RUNTIME_CONFIG__;
        if (!maybe) return undefined;
        return typeof maybe === 'string' ? JSON.parse(maybe) : maybe;
    } catch (e) {
        return undefined;
    }
}

const runtimeCfg = parseRuntimeConfig();
const rtServers = runtimeCfg?.servers;

let computedBaseUrl: string | undefined;

if (rtServers && rtServers['api']) {
    const s = rtServers['api'];
    const host = s.host;
    const port = s.port ? String(s.port) : undefined;
    if (host && port) {
        computedBaseUrl = `http://${host}:${port}`;
    }
}

if (!computedBaseUrl) {
    if (dev) {
        computedBaseUrl = `http://localhost:7770`;
    } else {
        let relative = generatedServers.productionApi || '/api';
        try {
            if (generatedServers) {
                for (const [, v] of Object.entries(generatedServers as any)) {
                    if (typeof v === 'string' && v.startsWith('/')) {
                        relative = v;
                        break;
                    }
                }
            }
        } catch (e) {
            // ignore
        }
        computedBaseUrl = relative;
    }
}

const API_BASE_URL = computedBaseUrl as string;

// Warn once if we're still falling back to localhost (helps catch missing build-time config)
let doneFallback = false;

function warnIfLocalhostFallback() {
    if (doneFallback) {
        return;
    }

    try {
        if (typeof window !== 'undefined' && API_BASE_URL.includes('localhost')) {
            console.warn('Frontend is using a localhost fallback for API URL. Build-time config not injected or runtime config not set.');
            doneFallback = true;
        }
    } catch (e) {
        // ignore
    }
}

defaults.baseUrl = API_BASE_URL;

// Include credentials (cookies) with all requests for authentication
defaults.credentials = "include";

// Re-export everything from the generated client
export * from "./client.gen";
export * from "./init";
export * from "./functions.custom";
export { API_BASE_URL, warnIfLocalhostFallback };
