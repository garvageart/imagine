import type { ImageUploadFileData } from "$lib/upload/manager.svelte";
import { API_BASE_URL } from "$lib/api";

type RequestInitOptions = { fetch?: typeof fetch; } & RequestInit;

export async function sendAPIRequest<T>(path: string, options: RequestInitOptions, form: true): Promise<Response>;
export async function sendAPIRequest<T>(path: string, options?: RequestInitOptions, form?: false): Promise<T>;
export async function sendAPIRequest<T>(path: string, options?: RequestInitOptions, form: boolean = false): Promise<T | Response> {
    if (path.startsWith("/")) {
        path = path.substring(1);
    }

    if (form) {
        const base = API_BASE_URL;
        if (options?.fetch) {
            return options.fetch(`${base}/${path}`, options);
        }
        return fetch(`${base}/${path}`, options);
    }

    const base = API_BASE_URL;
    if (options?.fetch) {
        const res = await options.fetch(`${base}/${path}`, options);
        return res.json() as Promise<T>;
    }

    return fetch(`${base}/${path}`, options).then(res => res.json() as Promise<T>);
}

// From https://github.com/immich-app/immich/main/web/src/lib/utils.ts#L55
export interface UploadRequestOptions {
    path?: string;
    data: ImageUploadFileData;
    method?: "POST" | "PUT";
    onUploadProgress?: (event: ProgressEvent<XMLHttpRequestEventTarget>) => void;
}

export const uploadRequest = async <T>(options: UploadRequestOptions): Promise<{ data: T; status: number; }> => {
    const { onUploadProgress, data, path = "/images" } = options;

    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.addEventListener('error', (error) => reject(error));
        xhr.addEventListener('load', () => {
            if (xhr.readyState === 4 && xhr.status >= 200 && xhr.status < 300) {
                resolve({ data: xhr.response as T, status: xhr.status });
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

        const base = API_BASE_URL;
        xhr.open(options.method || 'POST', `${base}${path}`);
        xhr.responseType = 'json';
        xhr.send(formData);
    });
};