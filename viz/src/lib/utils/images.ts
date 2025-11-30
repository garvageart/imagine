import { thumbHashToDataURL } from "thumbhash";
import type { Image } from "../api";

/**
 * Converts a date in EXIF format to a format that
 * can be parsed by the native ````Date```` object.
 * 
 * @param {String} EXIFDateTime A date in EXIF format that can be parsed by the function
 * @returns {Date} The parsed EXIF date as a native Date object
 */
export function convertEXIFDateTime(EXIFDateTime: string): Date {
    const EXIFDate = EXIFDateTime.split(" ")[0];
    const EXIFTime = EXIFDateTime.split(" ")[1];

    const EXIFDateFormated = EXIFDate.replaceAll(":", "/");

    const EXIFDateTimeString = `${EXIFDateFormated} ${EXIFTime}`;
    const EXIFDateObject = new Date(EXIFDateTimeString);

    return EXIFDateObject;
}

// Parse a variety of EXIF date formats similar to backend ConvertEXIFDateTime
const DATE_FORMATS = [
    /^(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})$/, // 2006:01:02 15:04:05
    /^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})$/, // 2006-01-02 15:04:05
    /^(\d{4}):(\d{2}):(\d{2})T(\d{2}):(\d{2}):(\d{2})(Z|[+\-]\d{2}:?\d{2})$/, // 2006:01:02T15:04:05Z07:00
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(Z|[+\-]\d{2}:?\d{2})$/, // 2006-01-02T15:04:05Z07:00
    /^(\d{4}):(\d{2}):(\d{2})$/, // date only 2006:01:02
    /^(\d{4})-(\d{2})-(\d{2})$/ // date only 2006-01-02
];

export function parseExifDate(raw?: string | null): Date | undefined {
    if (!raw) {
        return undefined;
    }

    let s = raw.trim();
    const parenIdx = s.indexOf(" (");
    if (parenIdx > 0) s = s.slice(0, parenIdx).trim();

    for (const rx of DATE_FORMATS) {
        const m = s.match(rx);
        if (!m) continue;
        try {
            // Normalise separators to ISO where needed
            if (rx === DATE_FORMATS[0]) {
                // 2006:01:02 15:04:05 -> 2006-01-02T15:04:05
                const [, Y, M, D, h, mm, ss] = m;
                return new Date(`${Y}-${M}-${D}T${h}:${mm}:${ss}Z`);
            }
            if (rx === DATE_FORMATS[1]) {
                const [, Y, M, D, h, mm, ss] = m;
                return new Date(`${Y}-${M}-${D}T${h}:${mm}:${ss}Z`);
            }
            if (rx === DATE_FORMATS[2] || rx === DATE_FORMATS[3]) {
                // Already close to ISO, ensure T separator
                return new Date(s.replace(/:(?=\d{2}T)/, ":"));
            }
            if (rx === DATE_FORMATS[4]) {
                const [, Y, M, D] = m;
                return new Date(`${Y}-${M}-${D}T00:00:00Z`);
            }
            if (rx === DATE_FORMATS[5]) {
                const [, Y, M, D] = m;
                return new Date(`${Y}-${M}-${D}T00:00:00Z`);
            }
        } catch (e) {
        }
    }

    // Fallback: let Date try (may be unreliable, but last resort)
    const d = new Date(s);
    return isNaN(d.getTime()) ? undefined : d;
}

export function getTakenAt(image: Image): Date {
    if (image.taken_at) {
        return new Date(image.taken_at);
    }

    // Priority: EXIF Original -> EXIF Modify -> metadata file_created_at -> image.created_at
    const exif = image.exif;
    const dates: (string | undefined)[] = [
        exif?.date_time_original,
        exif?.date_time,
        exif?.modify_date,
        image.image_metadata?.file_created_at,
        image.image_metadata?.file_modified_at,
        image.created_at
    ];

    for (const date of dates) {
        const parsed = parseExifDate(date);
        if (parsed) {
            return parsed;
        }
    }
    return new Date(image.created_at);
}

export function compareByTakenAtDesc(a: Image, b: Image): number {
    return getTakenAt(b).getTime() - getTakenAt(a).getTime();
}

export function getThumbhashURL(asset: Image): string | undefined {
    if (!asset.image_metadata?.thumbhash) {
        return undefined;
    }

    try {
        const binaryString = atob(asset.image_metadata.thumbhash);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }
        return thumbHashToDataURL(bytes);
    } catch (error) {
        console.warn("Failed to decode thumbhash:", error);
        return undefined;
    }
}

export function formatBytes(bytes?: number) {
    if (!bytes && bytes !== 0) {
        return null;
    }

    const units = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    let v = bytes as number;

    while (v >= 1024 && i < units.length - 1) {
        v = v / 1024;
        i++;
    }

    return `${v % 1 === 0 ? v.toFixed(0) : v.toFixed(2)} ${units[i]}`;
}

export interface TransformParams {
    format?: "webp" | "png" | "jpg" | "jpeg" | "avif" | "heif";
    flip?: string;
    kernel?: string;
    width?: number;
    height?: number;
    quality?: number;
    rotate?: number;
}

export function parseTransformParams(pathStr: string): TransformParams {
    const url = new URL(pathStr, window.location.origin); // not really true, just there so the thing doesn't break
    const q = url.searchParams;

    const params: TransformParams = {};

    const format = q.get("format");
    if (format) {
        params.format = format as TransformParams["format"];
    }

    const flip = q.get("flip");
    if (flip) {
        params.flip = flip;
    }

    const kernel = q.get("kernel");
    if (kernel) {
        params.kernel = kernel;
    }

    // Support both shorthand and longhand as parameter names
    // Note: just for now
    let widthParam = q.get("width");
    if (!widthParam) {
        widthParam = q.get("w");
    }

    if (widthParam) {
        const w = parseInt(widthParam, 10);
        if (!isNaN(w)) {
            params.width = w;
        }
    }

    let heightParam = q.get("height");
    if (!heightParam) {
        heightParam = q.get("h");
    }

    if (heightParam) {
        const h = parseInt(heightParam, 10);
        if (!isNaN(h)) {
            params.height = h;
        }
    }

    const qualityParam = q.get("quality");
    if (qualityParam) {
        const qn = parseInt(qualityParam, 10);
        if (!isNaN(qn)) {
            params.quality = qn;
        }
    }

    const rotateParam = q.get("rotate");
    if (rotateParam) {
        const r = parseInt(rotateParam, 10);
        if (!isNaN(r)) {
            params.rotate = r;
        }
    }

    return params;
}