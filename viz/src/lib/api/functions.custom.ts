/**
* Custom API functions that require special handling beyond what openapi-fetch provides.
* These are manually written for specific use cases like upload progress tracking.
*/
import type { ImageUploadFileData } from "$lib/upload/manager.svelte";
import { API_BASE_URL, defaults, type ImageUploadResponse } from "./client";
import type { DownloadRequest, ErrorResponse } from "./client.gen";
import * as Oazapfts from "@oazapfts/runtime";
import * as QS from "@oazapfts/runtime/query";

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
): Promise<{ data: ImageUploadResponse; status: number; }> {
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
                const response = xhr.response as ImageUploadResponse;
                resolve({ data: response, status: xhr.status });
            } else {
                reject({ data: xhr.response as ErrorResponse, status: xhr.status });
            }
        });

        if (onUploadProgress) {
            xhr.upload.addEventListener('progress', (event) => onUploadProgress(event));
        }

        const formData = new FormData();
        for (const [key, value] of Object.entries(data)) {
            formData.append(key, value);
        }

        const base = API_BASE_URL;
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
    const base = API_BASE_URL;
    return `${base}${path}`;
}

// --- Realtime helpers (custom) ---

export async function getJobsSnapshot(): Promise<{ data: any; status: number; }> {
    const base = API_BASE_URL;
    const res = await fetch(`${base}/jobs/snapshot`, {
        credentials: "include"
    });
    const data = await res.json().catch(() => ({}));
    return { data, status: res.status };
}

export async function updateJobTypeConcurrency(
    jobType: string,
    body: { concurrency: number; }
): Promise<{ data: any; status: number; }> {
    const base = API_BASE_URL;
    const url = `${base}/jobs/types/${encodeURIComponent(jobType)}/concurrency`;
    try {
        const res = await fetch(url, {
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
    downloadRequest: DownloadRequest,
    password?: string,
    opts?: Oazapfts.RequestOpts
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
    const fetch = opts?.fetch || defaults.fetch || window.fetch.bind(window);

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

        const response = await fetch(url, {
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
    | { status: 304; } // Add 304 Not Modified status
    | { status: 400; data: ErrorResponse; }
    | { status: 401; data: ErrorResponse; }
    | { status: 403; data: ErrorResponse; }
    | { status: 500; data: ErrorResponse; }
> {
    const baseUrl = API_BASE_URL;
    const queryParams = QS.query(QS.explode(params));
    const url = `${baseUrl}/images/${encodeURIComponent(uid)}/file${queryParams}`;
    const fetch = opts?.fetch || defaults.fetch || window.fetch.bind(window);

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

        const response = await fetch(url, {
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
