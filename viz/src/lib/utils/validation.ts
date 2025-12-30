export const UID_DEFAULT_LENGTH = 24;
const UID_ALPHANUMERIC_REGEX = /^[a-zA-Z0-9]+$/;

export function isValidUid(str: string): boolean {
    return str.length === UID_DEFAULT_LENGTH && UID_ALPHANUMERIC_REGEX.test(str);
}

export function isVizMimeType(mimeType: string): boolean {
    return mimeType.startsWith("application/x-viz.");
}

