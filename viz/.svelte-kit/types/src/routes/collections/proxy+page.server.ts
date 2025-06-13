// @ts-nocheck
import { MEDIA_SERVER } from "$lib/constants";
import type { Collection } from "$lib/types/images";
import { createServerURL, generateRandomString } from "$lib/utils";
import type { PageServerLoad } from "./$types";
import { DateTime } from "luxon";

export const load = async ({ fetch }: Parameters<PageServerLoad>[0]) => {
    // const collectionData: Collection[] = fetch(`${createServerURL(MEDIA_SERVER)}/collections/`, {
    //     method: "GET",
    //     headers: {
    //         "Content-Type": "application/json"
    //     }
    // });

    const collectionId = generateRandomString(16);

    const testData: Collection = {
        id: collectionId,
        created_by: "Les",
        created_on: DateTime.now().toJSDate(),
        name: "Fake Name Test Collection",
        description: "This is a fake collection",
        updated_on: DateTime.now().set({
            month: 8
        }).toJSDate(),
        images: [],
        image_count: 0,
        private: true,
    };

    return testData;
};
