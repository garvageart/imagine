import fs from 'fs';
import { CLIENT_IS_PRODUCTION, type ServerURLConfig } from "./constants";

/**
 * Reads a URL's hash and returns an object containing the query key/pair values as a properties
 * @param  {string url} URL query string
 */
export function getURLParams(url: string): any {
    // Ew
    const queryParams = Object.fromEntries(new URL(url).searchParams.entries());
    return queryParams;
}

/**
 * Various methods for storing, retrieving and deleting cookies from the browser
 */
export const cookieMethods = {
    set: (key: string, value: string, expiresDate?: Date | string) => {
        document.cookie = `${key}=${value}; expires=${expiresDate}; Secure; path =/`;
    },
    get: (key: string): string | null => {
        if (!browser) {
            return null;
        }

        const allCookies = document?.cookie;
        const cookieValue = allCookies.split("; ").find(cookie => cookie.startsWith(`${key}`))?.split("=")[1]!;

        return cookieValue;
    },
    delete: (key: string) => {
        document.cookie = `${key}=; max-age=0; path =/`;
    }
};

export const sleep = (time: number) => new Promise(resolve => setTimeout(resolve, time));

export function readConfig(): any {
    const fileData = fs.readFileSync('../../../config/imagine.json');
    return JSON.parse(fileData.toString());
}

export function createServerURL(serverURL: ServerURLConfig): string {
    if (!CLIENT_IS_PRODUCTION) {
        return serverURL.url;
    } else {
        return serverURL.prod;
    }
}

export const fullscreen = {
    enter: () => {
        const documentEl = document.documentElement;
        if (documentEl.requestFullscreen && !document.fullscreenElement) {
            documentEl.requestFullscreen();
        }
    },
    exit: () => {
        if (document.exitFullscreen && document.fullscreenElement) {
            document.exitFullscreen();
        }
    }
};