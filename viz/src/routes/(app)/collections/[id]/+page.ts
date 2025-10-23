import type { PageLoad } from "./$types";
import { createTestCollection, createTestImageObject } from "$lib/data/test";
import type { ImageObjectData } from "$lib/entities/image";

export const load: PageLoad = ({ fetch, params, url }) => {
    const collection = createTestCollection();
    const images: ImageObjectData[] = [];
    collection.id = params.id;

    for (let i = 0; i < collection.image_count; i++) {
        images.push(createTestImageObject());
    };
    collection.images = images;


    return {
        response: collection
    };
};