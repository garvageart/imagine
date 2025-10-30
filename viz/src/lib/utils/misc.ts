import { dev } from "$app/environment";

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

export function isObject(obj: any) {
    return obj !== null && typeof obj === 'object';
}

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
        if (window.debug) {
            console.info(`saving "viz:${this.key}" to local storage`);
        }

        this.value = value;
        let tempStr: string;

        if (isObject(value)) {
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