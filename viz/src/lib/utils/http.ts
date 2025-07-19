import { MEDIA_SERVER } from "$lib/constants";
import { createServerURL } from "./url";

export async function sendAPIRequest<T>(path: string, options?: RequestInit, form: boolean = false) {
    if (path.startsWith("/")) {
        path = path.substring(1);
    }

    if (form) {
        return fetch(`${createServerURL(MEDIA_SERVER)}/${path}`, options);
    }

    return fetch(`${createServerURL(MEDIA_SERVER)}/${path}`, options).then(res => res.json() as Promise<T>).catch(console.error);
}