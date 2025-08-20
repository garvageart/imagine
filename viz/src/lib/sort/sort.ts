import type { AssetSort } from "$lib/types/asset";
import type { IImageObjectData } from "$lib/types/images";
import { orderBy } from "lodash-es";

export function sortCollectionImages(assets: IImageObjectData[], sort: AssetSort) {
    switch (sort.by) {
        case "name":
            return orderBy(assets, "name", sort.order);
        case "created_at":
            return orderBy(assets, "uploaded_on", sort.order);
        case "updated_at":
            return orderBy(assets, "updated_on", sort.order);
        case "most_recent":
            return orderBy(assets, "updated_on", sort.order);
        default:
            return assets;
    }
}