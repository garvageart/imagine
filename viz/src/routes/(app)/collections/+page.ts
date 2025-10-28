import type { PageLoad } from "./$types";
import { listCollections, type Collection } from "$lib/api";
import CollectionData from "$lib/entities/collection";

export const load: PageLoad = async () => {
    const response = await listCollections();

    if (response.data && 'items' in response.data) {
        const allCollections = {
            ...response.data,
            items: response.data.items.map((item: Collection) => CollectionData.fromAPI(item))
        };

        return allCollections;
    }

    // API returned an error shape - return it so the caller can handle it
    return response.data;
};
