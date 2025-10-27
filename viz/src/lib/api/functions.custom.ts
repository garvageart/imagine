/**
* Custom API functions that require special handling beyond what openapi-fetch provides.
* These are manually written for specific use cases like upload progress tracking.
*/
import { createServerURL } from "$lib/utils/url";
import { MEDIA_SERVER } from "$lib/constants";
import type { ImageUploadFileData } from "$lib/upload/manager.svelte";

export interface UploadImageOptions {
    data: ImageUploadFileData;
    onUploadProgress?: (event: ProgressEvent<XMLHttpRequestEventTarget>) => void;
}

export interface UploadImageResponse {
    id: string;
}

export interface UploadImageResult {
    uid: string;
    metadata?: any;
}

/**
 * Upload an image with progress tracking using XMLHttpRequest.
 * This is a custom implementation because openapi-fetch doesn't support progress events.
 * 
 * Note: Maps the API response `id` to `uid` for consistency with the rest of the app.
 */
export async function uploadImageWithProgress(
    options: UploadImageOptions
): Promise<{ data: UploadImageResult; status: number; }> {
    const { onUploadProgress, data } = options;

    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.addEventListener('error', (error) => reject(error));
        xhr.addEventListener('load', () => {
            if (xhr.readyState === 4 && xhr.status >= 200 && xhr.status < 300) {
                const response = xhr.response as UploadImageResponse;
                // Map API response to app format (id -> uid)
                resolve({
                    data: { uid: response.id },
                    status: xhr.status
                });
            } else {
                reject({ data: xhr.response, status: xhr.status });
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
        xhr.responseType = 'json';
        xhr.send(formData);
    });
}
