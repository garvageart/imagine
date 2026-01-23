import type { Collection, ImageAsset } from "$lib/api";
import type { AssetSort } from "$lib/types/asset";
import { orderBy } from "lodash-es";
import { getTakenAt } from "$lib/utils/images";

function getCollectionDate(collection: Collection): Date {
    return new Date(collection.updated_at || collection.created_at);
}

function getMostRecentImageDate(img: ImageAsset): number {
    // For most recent: prioritize when it was added/uploaded to the system
    // created_at = when uploaded, updated_at = when modified, file_created_at = original file date
    return new Date(img.updated_at || img.created_at).getTime();
}

export function sortCollectionImages(assets: ImageAsset[], sort: AssetSort) {
    switch (sort.by) {
        case "name":
            return orderBy(assets, "name", sort.order);
        case "created_at":
            return orderBy(assets, (img) => new Date(img.created_at).getTime(), sort.order);
        case "updated_at":
            return orderBy(assets, (img) => new Date(img.updated_at).getTime(), sort.order);
        case "oldest":
            return orderBy(assets, (img) => getTakenAt(img).getTime(), sort.order);
        case "most_recent":
            return orderBy(assets, (img) => getMostRecentImageDate(img), sort.order);
        default:
            return assets;
    }
}

export function sortCollections(collections: Collection[], sort: AssetSort) {
    switch (sort.by) {
        case "name":
            return orderBy(collections, "name", sort.order);
        case "created_at":
            return orderBy(collections, (col) => new Date(col.created_at).getTime(), sort.order);
        case "updated_at":
            return orderBy(collections, (col) => new Date(col.updated_at).getTime(), sort.order);
        case "oldest":
            return orderBy(collections, (col) => getCollectionDate(col).getTime(), sort.order);
        case "most_recent":
            return orderBy(collections, (col) => getCollectionDate(col).getTime(), sort.order);
        default:
            return collections;
    }
}