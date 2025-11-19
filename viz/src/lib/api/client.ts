/**
 * Configured oazapfts API client
 * Auto-generated functions with proper base URL and SvelteKit fetch integration
 */
import { defaults, servers as generatedServers } from "./client.gen";
import { dev } from "$app/environment";
import type { DownloadRequest, ErrorResponse } from "./client.gen";
import * as QS from "@oazapfts/runtime/query";

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

/**
 * Download images as a ZIP blob using a token
 * This is a custom implementation because oazapfts doesn't properly handle binary responses
 */
export async function downloadImagesBlob(
    token: string,
    downloadRequest: DownloadRequest,
    password?: string
): Promise<
    | { status: 200; data: Blob; }
    | { status: 400; data: ErrorResponse; }
    | { status: 401; data: ErrorResponse; }
    | { status: 403; data: ErrorResponse; }
    | { status: 500; data: ErrorResponse; }
> {
    const baseUrl = defaults.baseUrl || "";
    const queryParams = QS.query(QS.explode({ token, password }));
    const url = `${baseUrl}/download${queryParams}`;

    try {
        const response = await fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                ...defaults.headers
            },
            credentials: "include",
            body: JSON.stringify(downloadRequest)
        });

        if (response.ok) {
            const blob = await response.blob();
            return { status: 200, data: blob };
        }

        // Try to parse error response as JSON
        const contentType = response.headers.get("content-type") || "";
        if (contentType.includes("application/json")) {
            const errorData = await response.json();
            return {
                status: response.status as 400 | 401 | 403 | 500,
                data: errorData
            };
        }

        // Fallback error
        return {
            status: response.status as 400 | 401 | 403 | 500,
            data: { error: `Request failed with status ${response.status}` }
        };
    } catch (error) {
        return {
            status: 500,
            data: { error: error instanceof Error ? error.message : "Network error" }
        };
    }
}

// Re-export everything from the generated client
export * from "./client.gen";
export { initApi } from "./init";
export { API_BASE_URL };
export { warnIfLocalhostFallback };
