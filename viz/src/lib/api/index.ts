// viz/src/lib/api/index.ts
import * as generated from "./client.gen";
import type * as Oazapfts from "@oazapfts/runtime"; // Import type for RequestOpts
import * as QS from "@oazapfts/runtime/query";
import { defaults, servers } from "./client.gen";
import type { ImageUploadFileData } from "$lib/upload/manager.svelte";

// Initialize defaults for the underlying oazapfts runtime
defaults.baseUrl = servers.productionApi;
defaults.credentials = "include";

let currentFetch: typeof globalThis.fetch = globalThis.fetch; // Default to window.fetch initially

/**
 * Initializes the API client with SvelteKit's enhanced fetch function.
 * This must be called once in the root layout load function or similar entry point.
 *
 * @param fetch - SvelteKit's enhanced fetch from the load context.
 */
export function initApi(fetch: typeof globalThis.fetch) {
    currentFetch = fetch;
    // Also explicitly set the fetch on the generated defaults, though the proxy below
    // will ensure currentFetch is passed to every call regardless.
    generated.defaults.fetch = currentFetch;
}

// Create a proxy for the generated API functions
type GeneratedApi = typeof generated; // Get the type of the generated API module

// Proxy the 'generated' module directly to leverage its types and properties
const apiProxy: GeneratedApi = new Proxy(generated, {
    get(target: GeneratedApi, prop: keyof GeneratedApi) {
        const originalMethod = target[prop];

        // If it's a non-function property (like a type, defaults, servers), return it directly
        // This handles types, constants, or other non-callable exports from client.gen.ts
        if (typeof originalMethod !== 'function') {
            return originalMethod;
        }

        // Return a new function that wraps the original generated API method
        return function (this: any, ...methodArgs: any[]): ReturnType<typeof originalMethod> {
            const finalArgs = [...methodArgs]; // Create a new array for manipulation

            let opts: Oazapfts.RequestOpts | undefined = undefined;
            let optsIndex = -1;

            // Check if the last argument is an object (potential opts)
            // This heuristic works for oazapfts where opts is always the last argument if present.
            if (finalArgs.length > 0 && typeof finalArgs[finalArgs.length - 1] === 'object' && finalArgs[finalArgs.length - 1] !== null) {
                opts = finalArgs[finalArgs.length - 1] as Oazapfts.RequestOpts;
                optsIndex = finalArgs.length - 1;
            }

            // Inject the currentFetch into the options
            const injectedOpts: Oazapfts.RequestOpts = {
                credentials: "include",
                ...opts,
                fetch: currentFetch
            };

            if (optsIndex !== -1) {
                // Replace the existing opts with the new one that has fetch injected
                finalArgs[optsIndex] = injectedOpts;
            } else {
                // If no opts object was found, append the newOpts object as the last argument
                finalArgs.push(injectedOpts);
            }

            // Call the original generated API method with the modified arguments.
            // Using 'as (...args: any[]) => any' helps TypeScript with the dynamic nature of apply.
            // The 'this' context is preserved.
            return (originalMethod as (...args: any[]) => any).apply(this, finalArgs) as ReturnType<typeof originalMethod>;
        };
    },
});

// Export the proxied API client as 'api'
export const api = apiProxy;

// Re-export other non-function exports like defaults, servers, and all types separately
export { defaults, servers };
export * from "./client.gen"; // This re-exports all types from the generated client.

// --- Exports from the old client.ts and custom functions ---
export const API_BASE_URL = defaults.baseUrl; // Export the configured base URL

let doneFallback = false;

export function warnIfLocalhostFallback() {
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

export interface UploadImageOptions {
    data: ImageUploadFileData;
    onUploadProgress?: (event: ProgressEvent<XMLHttpRequestEventTarget>) => void;
    request?: XMLHttpRequest;
}

/**
 * Upload an image with progress tracking using XMLHttpRequest.
 * This is a custom implementation because openapi-fetch doesn't support progress events.
 *
 * Note: Maps the API response `id` to `uid` for consistency with the rest of the app.
 */
export async function uploadImageWithProgress(
    options: UploadImageOptions
): Promise<{ data: generated.ImageUploadResponse; status: number; }> {
    const { onUploadProgress, data } = options;

    const xhr = new XMLHttpRequest();

    // Bind XHR instance back to caller's request reference for cancellation support
    if (options.request !== undefined) {
        options.request = xhr;
    }

    return new Promise((resolve, reject) => {
        xhr.addEventListener('error', (error) => reject(error));
        xhr.addEventListener('load', () => {
            if (xhr.readyState === 4 && xhr.status >= 200 && xhr.status < 300) {
                const response = xhr.response as generated.ImageUploadResponse;
                resolve({ data: response, status: xhr.status });
            } else {
                reject({ data: xhr.response as generated.ErrorResponse, status: xhr.status });
            }
        });

        if (onUploadProgress) {
            xhr.upload.addEventListener('progress', (event) => onUploadProgress(event));
        }

        const formData = new FormData();
        for (const [key, value] of Object.entries(data)) {
            // Map filename to file_name to match API expectation
            if (key === "filename") {
                formData.append("file_name", value);
            } else {
                formData.append(key, value);
            }
        }

        const base = API_BASE_URL; // Use the exported API_BASE_URL
        xhr.open('POST', `${base}/images`);
        xhr.withCredentials = true;
        xhr.responseType = 'json';
        xhr.send(formData);
    });
}

export function getFullImagePath(path: string): string {
    // If path is already a full URL (starts with http:// or https://), return as-is
    if (path.startsWith('http://') || path.startsWith('https://')) {
        return path;
    }
    const base = API_BASE_URL; // Use the exported API_BASE_URL
    return `${base}${path}`;
}

// --- Realtime helpers (custom) ---

export async function getJobsSnapshot(): Promise<{ data: any; status: number; }> {
    const base = API_BASE_URL; // Use the exported API_BASE_URL
    // Use `currentFetch` for custom fetch calls
    const res = await currentFetch(`${base}/jobs/snapshot`, {
        credentials: "include"
    });
    const data = await res.json().catch(() => ({}));
    return { data, status: res.status };
}

export async function updateJobTypeConcurrency(
    jobType: string,
    body: { concurrency: number; }
): Promise<{ data: any; status: number; }> {
    const base = API_BASE_URL; // Use the exported API_BASE_URL
    const url = `${base}/jobs/types/${encodeURIComponent(jobType)}}/concurrency`;
    try {
        // Use `currentFetch` for custom fetch calls
        const res = await currentFetch(url, {
            method: "PUT",
            credentials: "include",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(body)
        });

        const data = await res.json().catch(() => ({}));
        return { data, status: res.status };
    } catch (err) {
        return { data: { error: err instanceof Error ? err.message : String(err) }, status: 500 };
    }
}

/**
 * Download images as a ZIP blob using a token
 * This is a custom implementation because oazapfts doesn't properly handle binary responses
 */
export async function downloadImagesZipBlob(
    token: string,
    downloadRequest: generated.DownloadRequest,
    password?: string,
    opts?: Oazapfts.RequestOpts
): Promise<
    | { status: 200; data: Blob; }
    | { status: 400; data: generated.ErrorResponse; }
    | { status: 401; data: generated.ErrorResponse; }
    | { status: 403; data: generated.ErrorResponse; }
    | { status: 500; data: generated.ErrorResponse; }
> {
    const baseUrl = defaults.baseUrl || "";
    const queryParams = QS.query(QS.explode({ token, password }));
    const url = `${baseUrl}/download${queryParams}`;
    const fetchToUse = opts?.fetch || currentFetch; // Use currentFetch if not overridden by opts

    try {
        const defaultHeaders = defaults.headers;
        const customHeaders = opts?.headers || {};
        const headers: Record<string, any> = {};
        for (const [key, value] of Object.entries(defaultHeaders)) {
            headers[key] = value;
        }

        for (const [key, value] of Object.entries(customHeaders)) {
            headers[key] = value;
        }

        const response = await fetchToUse(url, {
            cache: opts?.cache || defaults.cache,
            credentials: opts?.credentials || defaults.credentials,
            keepalive: opts?.keepalive || defaults.keepalive,
            integrity: opts?.integrity || defaults.integrity,
            method: opts?.method || defaults.method,
            redirect: opts?.redirect || defaults.redirect,
            referrer: opts?.referrer || defaults.referrer,
            referrerPolicy: opts?.referrerPolicy || defaults.referrerPolicy,
            mode: opts?.mode || defaults.mode,
            signal: opts?.signal || defaults.signal,
            priority: opts?.priority || defaults.priority,
            headers: {
                "Content-Type": "application/json",
                ...headers
            },
            body: JSON.stringify(downloadRequest),
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

/**
 * Custom API function to fetch an image file as a Blob.
 * This is a custom implementation because oazapfts.fetchJson does not correctly handle binary responses.
 */
export async function getImageFileBlob(
    uid: string,
    params: {
        format?: "webp" | "png" | "jpg" | "jpeg" | "avif" | "heif";
        width?: number;
        height?: number;
        quality?: number;
        download?: "1";
        token?: string;
        password?: string;
    } = {}, opts?: Oazapfts.RequestOpts
): Promise<
    | { status: 200; data: Blob; }
    | { status: 304; }
    | { status: 400; data: generated.ErrorResponse; }
    | { status: 401; data: generated.ErrorResponse; }
    | { status: 403; data: generated.ErrorResponse; }
    | { status: 500; data: generated.ErrorResponse; }
> {
    const baseUrl = API_BASE_URL;
    const queryParams = QS.query(QS.explode(params));
    const url = `${baseUrl}/images/${encodeURIComponent(uid)}/file${queryParams}`;
    const fetchToUse = opts?.fetch || currentFetch; // Use currentFetch if not overridden by opts

    try {
        const defaultHeaders = defaults.headers;
        const customHeaders = opts?.headers || {};
        const headers: Record<string, any> = {};
        for (const [key, value] of Object.entries(defaultHeaders)) {
            headers[key] = value;
        }

        for (const [key, value] of Object.entries(customHeaders)) {
            headers[key] = value;
        }

        const response = await fetchToUse(url, {
            cache: opts?.cache || defaults.cache,
            credentials: opts?.credentials || defaults.credentials,
            keepalive: opts?.keepalive || defaults.keepalive,
            integrity: opts?.integrity || defaults.integrity,
            method: opts?.method || defaults.method,
            redirect: opts?.redirect || defaults.redirect,
            referrer: opts?.referrer || defaults.referrer,
            referrerPolicy: opts?.referrerPolicy || defaults.referrerPolicy,
            mode: opts?.mode || defaults.mode,
            signal: opts?.signal || defaults.signal,
            priority: opts?.priority || defaults.priority,
            headers
        });

        // Ideally we never get here
        if (response.status === 304) {
            return { status: 304 }; // Return 304 for Not Modified
        }

        if (response.ok) {
            const blob = await response.blob();
            return { status: response.status as 200, data: blob };
        }

        // Try to parse error response as JSON
        const contentType = response.headers.get("content-type") || "";
        if (contentType.includes("application/json")) {
            const errorData = await response.json();
            return {
                status: response.status as 400 | 401 | 403 | 500,
                data: errorData,
            };
        }

        // Fallback error
        return {
            status: response.status as 400 | 401 | 403 | 500,
            data: { error: `Request failed with status ${response.status}` },
        };
    } catch (error) {
        return {
            status: 500,
            data: { error: error instanceof Error ? error.message : "Network error" },
        };
    }
}