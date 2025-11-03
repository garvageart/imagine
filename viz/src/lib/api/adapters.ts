// Minimal adapter types to satisfy legacy imports until full API-adapter migration is completed.
export type APIPagination<T = any> = {
    items: T[];
    total?: number;
    limit?: number;
    offset?: number;
};