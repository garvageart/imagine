import { goto } from "$app/navigation";
import { ServerURLConfig, CLIENT_IS_PRODUCTION } from "$lib/constants";

export function updateURLParameter(name: string, value: string, keepFocus = true) {
    const url = new URL(window.location.href);
    const searchParams = url.searchParams;

    searchParams.set(name, value);
    const newUrl = url.origin + url.pathname + '?' + searchParams.toString() + url.hash;

    goto(newUrl, { replaceState: true, keepFocus, noScroll: true });
}

export function createServerURL(serverURL: ServerURLConfig): string {
    if (!CLIENT_IS_PRODUCTION) {
        return serverURL.url;
    } else {
        return serverURL.subdomain;
    }
}

/**
 * Reads a URL's hash and returns an object containing the query key/pair values as a properties
 * @param  {string url} URL query string
 */
export function getURLParams(url: string): any {
    // Ew
    const queryParams = Object.fromEntries(new URL(url).searchParams.entries());
    return queryParams;
}