/**
* Custom API functions that require special handling beyond what openapi-fetch provides.
* These are manually written for specific use cases like upload progress tracking.
*/
import { createServerURL } from "$lib/utils/url";
import { MEDIA_SERVER } from "$lib/constants";
import type { ImageUploadFileData } from "$lib/upload/manager.svelte";
import type { ErrorResponse, ImageUploadResponse } from "./client";

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

        xhr.open('POST', `${createServerURL(MEDIA_SERVER)}/images`);
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
    return `${createServerURL(MEDIA_SERVER)}${path}`;
}

// --- Realtime helpers (custom) ---

export async function getJobsSnapshot(): Promise<{ data: any; status: number; }> {
    const res = await fetch(`${createServerURL(MEDIA_SERVER)}/jobs/snapshot`, {
        credentials: "include"
    });
    const data = await res.json().catch(() => ({}));
    return { data, status: res.status };
}

export async function updateJobTypeConcurrency(
    jobType: string,
    body: { concurrency: number }
): Promise<{ data: any; status: number }> {
    const url = `${createServerURL(MEDIA_SERVER)}/jobs/types/${encodeURIComponent(jobType)}/concurrency`;
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