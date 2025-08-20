export interface IPagination {
    limit: number;
    offset: number;
}

export type AssetGridArray<T> = {
    asset: T;
    row: number;
    rowSize: number;
    column: number;
    columnSize: number;
    size: number;
}[][];

export type AssetSortOrder = "asc" | "desc";

export type AssetSortBy = "name" | "created_at" | "updated_at" | "oldest" | "most_recent";

export type AssetGroupBy = "none" | "year" | "month" | "day";

export type AssetDisplay = "list" | "cover";

export interface AssetSort {
    order: AssetSortOrder;
    by: AssetSortBy;
    group: {
        by: AssetGroupBy;
        order: AssetSortOrder;
    };
    display: AssetDisplay;
}