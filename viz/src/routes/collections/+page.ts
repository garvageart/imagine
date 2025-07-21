import type { Collection } from "$lib/types/images";
import type { PageLoad } from "./$types";
import { createTestCollection } from "$lib/data/test";

export const load: PageLoad = ({ fetch }) => {
    let allCollections: Collection[] = [];
    const randomCollectionCount = Math.floor(Math.random() * 70) + 70; // Random number between 5 and 25
    for (let i = 0; i < randomCollectionCount; i++) {
        const testData = createTestCollection();

        allCollections.push(testData);
    };

    return {
        response: allCollections
    };
};
