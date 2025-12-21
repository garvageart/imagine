import { DateTime } from "luxon";

// try to parse a value as a date using Luxon; returns DateTime or undefined
export function tryParseDate(v: any): DateTime | undefined {
    if (v == null || v === undefined) {
        return undefined;
    }

    if (typeof v === "string") {
        const s = v.trim();
        // Try ISO first
        let dt = DateTime.fromISO(s);
        if (dt.isValid) {
            return dt;
        }

        // Try RFC2822
        dt = DateTime.fromRFC2822(s);
        if (dt.isValid) {
            return dt;
        }
        // If it's numeric string, try as epoch
        const n = Number(s);
        if (!Number.isNaN(n)) {
            if (n > 1e12) {
                return DateTime.fromMillis(n);
            }

            if (n > 1e10) {
                return DateTime.fromMillis(n);
            }

            if (n > 1e9) {
                return DateTime.fromSeconds(n);
            }

        }
        return undefined;
    }

    if (typeof v === "number") {
        // Treat only large numbers as epoch timestamps (seconds or milliseconds).
        // Small integers (like image width/height, counts) should NOT be parsed as dates.
        if (v > 1e12) {
            return DateTime.fromMillis(v);
        }

        if (v > 1e10) {
            return DateTime.fromMillis(v);
        }

        if (v > 1e9) {
            return DateTime.fromSeconds(v);
        }

        return undefined;
    }

    return undefined;
} 