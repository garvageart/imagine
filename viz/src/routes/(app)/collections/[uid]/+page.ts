import type { PageLoad } from "./$types";
import { getCollection } from "$lib/api";
import { error } from "@sveltejs/kit";

export const load: PageLoad = async ({ params }) => {
    const collectionImages = await getCollection(params.uid);
    if (collectionImages.status !== 200) {
        error(collectionImages.status, {
            message: collectionImages.data?.error || "Failed to load collection"
        });
    }

    return collectionImages.data;
};