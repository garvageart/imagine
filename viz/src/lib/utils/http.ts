import type { ImageUploadFileData } from "$lib/upload/manager.svelte";
import { API_BASE_URL, downloadImagesZipBlob, getImageFile, getImageFileBlob, signDownload, type ImageAsset } from "$lib/api";
import { debugMode } from "$lib/states/index.svelte";

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

export async function downloadOriginalImageFile(img: ImageAsset) {
    const uid = img.uid;
    const fileRes = await getImageFileBlob(uid, {}, { cache: "no-cache" });
    if (fileRes.status === 304) {
        if (debugMode) {
            console.log(
                `Image ${uid} not modified, using cached version for download`
            );
        }
        return;
    } else if (fileRes.status !== 200) {
        throw new Error(`Failed to download image: ${fileRes.data.error}`);
    }

    // this should never happen man but hey
    const filename =
        img.name.trim() !== "" ? img.name : `image-${uid}-${Date.now()}`;
    const blob = fileRes.data;
    const url = URL.createObjectURL(blob);

    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();

    URL.revokeObjectURL(url);
}

// Helper to perform token-based bulk download given a list of UIDs.
// The server will create a download token and use it for authentication.
export async function performImageDownloads(images: ImageAsset[]) {
    if (!images || images.length === 0) {
        throw new Error("No image UIDs provided for download");
    }

    const uids = images.map(i => i.uid);

    if (uids.length === 1) {
        const uid = uids[0];
        const img = images.find((i) => i.uid === uid)!;

        return downloadOriginalImageFile(img);
    }

    const signRes = await signDownload({
        uids,
        expires_in: 300,
        allow_download: true,
        allow_embed: false,
        show_metadata: true
    });

    if (signRes.status !== 200) {
        throw new Error(signRes.data?.error ?? "Failed to sign download request");
    }

    const token = signRes.data.uid; // The token is stored in the uid field
    const dlRes = await downloadImagesZipBlob(token, { uids });

    if (dlRes.status !== 200) {
        throw new Error(dlRes.data?.error ?? "Failed to download archive");
    }

    // Extract filename from Content-Disposition header if available
    // Note: The custom function returns the blob directly, so we need to handle filename separately
    const filename = `images-${Date.now()}.zip`;

    const blob = dlRes.data;
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
}