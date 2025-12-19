import { browser } from "$app/environment";
import type { Cookies } from '@sveltejs/kit';

export const sleep = (time: number) => new Promise(resolve => setTimeout(resolve, time));

export function generateRandomString(length: number): string {
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }

    return result;
}

export const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen();
    } else {
        if (document.exitFullscreen) {
            document.exitFullscreen();
        }
    }
};

export function copyToClipboard(text: string) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(text);
        return;
    }

    const textArea = document.createElement("textarea");
    textArea.value = text;

    // Avoid scrolling to bottom
    textArea.style.top = "0";
    textArea.style.left = "0";
    textArea.style.position = "fixed";

    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    document.execCommand('copy');
    document.body.removeChild(textArea);
}

export function debounce<T extends (...args: any[]) => void>(
    fn: T,
    delay: number
): T {
    let timeoutID: ReturnType<typeof setTimeout> | undefined;
    return function (this: any, ...args: any[]) {
        clearTimeout(timeoutID);
        timeoutID = setTimeout(() => fn.apply(this, args), delay);
    } as T;
}

export function isObject(obj: any) {
    return obj !== null && typeof obj === 'object';
}

export class VizLocalStorage<V = string> {
    key: string;
    value: V | null = null;
    private prefix = "viz";

    constructor(key: string, value?: V) {
        this.key = key;

        if (value) {
            this.value = value;
        }
    }

    get = (): V | null => {
        const item = localStorage.getItem(this.prefix + ":" + this.key);

        if (!item || item === "undefined") {
            return null;
        }

        if ((item?.startsWith("{") && item?.endsWith("}")) || (item?.startsWith("[") && item?.endsWith("]"))) {
            return JSON.parse(item) as V;
        }

        if (item === "true" || item === "false") {
            if (item === "true") {
                return true as V;
            } else {
                return false as V;
            }
        }

        return item !== null ? item as V : null;
    };

    set = (value: V) => {
        this.value = value;
        let tempStr: string;

        if (isObject(value)) {
            tempStr = JSON.stringify(value);
        } else {
            tempStr = value as unknown as string;
        }

        localStorage.setItem(this.prefix + ":" + this.key, tempStr);
    };

    delete = () => {
        localStorage.removeItem(this.prefix + ":" + this.key);
    };
}

export class VizCookieStorage<V = string> {
    key: string;
    prefix = "viz";
    private serverCookies: Cookies | null = null;

    constructor(key: string, cookies?: Cookies) {
        this.key = key;

        if (!browser && cookies) {
            this.serverCookies = cookies;
        }
    }

    get = (): V | null => {
        const key = this.prefix + ":" + this.key;
        let item: string | null = null;
        if (this.serverCookies) {
            item = this.serverCookies.get(key) || null;
        } else if (browser) {
            const match = document.cookie.match(new RegExp('(^| )' + key + '=([^;]+)'));
            item = match ? decodeURIComponent(match[2]) : null;
        }

        if (!item || item === "undefined") {
            return null;
        }

        if ((item?.startsWith("{") && item?.endsWith("}")) || (item?.startsWith("[") && item?.endsWith("]"))) {
            return JSON.parse(item) as V;
        }

        if (item === "true" || item === "false") {
            if (item === "true") {
                return true as V;
            } else {
                return false as V;
            }
        }

        return item !== null ? item as V : null;
    };

    set = (value: V, options: { path?: string; maxAge?: number; sameSite?: 'strict' | 'lax' | 'none'; } = {}) => {
        const key = this.prefix + ":" + this.key;
        const { path = '/', maxAge = 31536000, sameSite = 'lax' } = options;
        let valueStr: string;

        if (isObject(value)) {
            valueStr = JSON.stringify(value);
        } else {
            valueStr = String(value);
        }

        if (this.serverCookies) {
            this.serverCookies.set(key, valueStr, { path, maxAge, sameSite });
        } else if (browser) {
            document.cookie = `${key}=${encodeURIComponent(valueStr)}; path=${path}; max-age=${maxAge}; SameSite=${sameSite}`;
        }
    };

    delete = (options: { path?: string; } = {}) => {
        const key = this.prefix + ":" + this.key;
        const { path = '/' } = options;

        if (this.serverCookies) {
            this.serverCookies.delete(key, { path });
        } else if (browser) {
            document.cookie = `${key}=; path=${path}; max-age=0`;
        }
    };
}

export function swapArrayElements<A>(array: A[], index1: number, index2: number) {
    array[index1] = array.splice(index2, 1, array[index1])[0];
};

export function arrayHasDuplicates(arr: any[]): { hasDuplicates: boolean, duplicates: any[]; } {
    let dupli: never[] = [];
    arr.reduce((acc, curr) => {
        if (acc.indexOf(curr) === -1 && arr.indexOf(curr) !== arr.lastIndexOf(curr)) {
            acc.push(curr);
        }
        return acc;
    }, dupli);

    if (dupli.length > 0) {
        return {
            hasDuplicates: true,
            duplicates: dupli
        };
    }

    return {
        hasDuplicates: false,
        duplicates: []
    };
}

export function normalizeBase64(str: string) {
    let normalized = str.replace(/-/g, "+").replace(/_/g, "/");
    const padding = normalized.length % 4;
    if (padding) {
        normalized += "=".repeat(4 - padding);
    }
    return normalized;
};