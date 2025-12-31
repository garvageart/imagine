import { generateRandomString } from "./misc";

export function generateKeyId(length = 10): string {
    return "sp-" + generateRandomString(length);
}

export function isTabData(obj: any): obj is any {
    const objKeys = Object.keys(obj);
    const hasValidAttrs = objKeys.includes("view") && objKeys.includes("index");
    return obj !== null && typeof obj === 'object' && hasValidAttrs;
}
