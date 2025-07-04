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
    get: (key: string): string | undefined => {
        const allCookies = document?.cookie;
        const cookieValue = allCookies.split("; ").find(cookie => cookie.startsWith(`${key}`))?.split("=")[1];

        return cookieValue;
    },
    delete: (key: string) => {
        document.cookie = `${key}=; max-age=0; path =/`;
    }
};

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

export function createServerURL(serverURL: ServerURLConfig): string {
    if (!CLIENT_IS_PRODUCTION) {
        return serverURL.url;
    } else {
        return serverURL.subdomain;
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

export function debounce(func: () => any, wait: number | undefined) {
    let timeoutId: number | null = null;
    return (...args: any) => {
        if (timeoutId !== null) {
            window.clearTimeout(timeoutId);
        }

        timeoutId = window.setTimeout(() => {
            // @ts-ignore
            func(...args);
        }, wait);
    };
};

export class VizLocalStorage<V = string> {
    key: string;
    value: V | null = null;

    constructor(key: string, value?: V) {
        this.key = key;

        if (value) {
            this.value = value;
        }
    }

    get = (): V | null => {
        if (window.debug) {
            console.info(`getting "viz:${this.key}" from local storage`);
        }
        const item = localStorage.getItem("viz:" + this.key);

        if (!item || item === "undefined") {
            return null;
        }

        if (item?.startsWith("{") || item?.startsWith("[")) {
            return JSON.parse(item) as V;
        }

        return item !== null ? item as V : null;
    };

    set = (value: V) => {
        if (window.debug) {
            console.info(`saving "viz:${this.key}" to local storage`);
        }

        this.value = value;
        let tempStr: string;

        if (typeof value === 'object' && value !== null) {
            tempStr = JSON.stringify(value);
        } else {
            tempStr = value as unknown as string;
        }

        localStorage.setItem("viz:" + this.key, tempStr);
    };

    delete = () => {
        localStorage.removeItem("viz:" + this.key);
    };
}

export function checkDOMForID(id: string) {
    const el = document.getElementById(id);

    if (el) {
        return true;
    }

    return false;
}

export function swapArrayElements<A>(array: A[], index1: number, index2: number) {
    array[index1] = array.splice(index2, 1, array[index1])[0];
};

export function debugEvent(event: CustomEvent, printAsString: boolean = false) {
    console.log("Event:", event.type);

    if (printAsString) {
        console.log("Detail:", JSON.stringify(event.detail, null, 2));
        return;
    }

    console.log("Detail:", event.detail);
}

// TODO: Move utility functions to purpose made files
// instead of shoving them down all one utility file
export function generateKeyId(length = 10): string {
    return "sp-" + generateRandomString(length);
}

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

export function resetAndReloadLayout() {
    localStorage.removeItem("viz:layout");
    location.reload();
}