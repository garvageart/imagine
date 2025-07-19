import { dev } from "$app/environment";
import { goto } from "$app/navigation";
import { page } from "$app/state";
import { createTestUser } from "$lib/data/test";
import CollectionData from "$lib/entities/collection";
import { ImageObjectData } from "$lib/entities/image";
import { search } from "$lib/states/index.svelte";
import { generateRandomString, sleep } from "$lib/utils/misc";
import { updateURLParameter } from "$lib/utils/url";
import { faker } from "@faker-js/faker";
import { DateTime } from "luxon";

export function transformQueryString(queryStr: string) {
    return queryStr.replace(/\s/g, "+");
}
export function redirectToSearchWithQuery() {
    goto(`/search?q=${transformQueryString(search.value)}`);
}

export async function performSearch() {
    if (!search.value.trim()) {
        return;
    }

    // TODO: Create search results dropdown and have an option to go to the search page
    // if the results aren't sufficient for the user
    // For now we just redirect to the search page

    if (page.url.pathname !== "/search") {
        redirectToSearchWithQuery();
        return;
    }

    search.loading = true;
    search.executed = true;

    try {

        const randomLatency = dev ? Math.floor(Math.random() * 2000) + 500 : 0;
        await sleep(randomLatency);

        updateURLParameter("q", search.value);
        search.data.collections = Array.from(
            { length: Math.floor(Math.random() * 45) + 15 },
            (_, i) =>
                new CollectionData({
                    id: generateRandomString(16),
                    name: `${faker.word.adjective()} ${faker.word.noun()} Photos`
                        .split(" ")
                        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
                        .join(" "),
                    description: faker.lorem.sentence(),
                    created_on: DateTime.now().toJSDate(),
                    updated_on: DateTime.now().toJSDate(),
                    images: [],
                    private: faker.datatype.boolean(),
                    created_by: {
                        id: generateRandomString(8),
                        first_name: faker.person.firstName(),
                        last_name: faker.person.lastName(),
                        username: faker.internet.username(),
                        email: faker.internet.email(),
                        created_on: faker.date.past({ years: 2 }),
                        role: "user",
                        updated_on: faker.date.recent({ days: 60 })
                    },
                    image_count: Math.floor(Math.random() * 100),
                    owner: createTestUser(),
                    thumbnail: new ImageObjectData({
                        id: generateRandomString(16),
                        name: `${faker.word.adjective()} ${faker.word.noun()}`,
                        collection_id: generateRandomString(16),
                        uploaded_on: faker.date.past({ years: 1 }),
                        uploaded_by: createTestUser(),
                        dupes: [],
                        private: false,
                        thumbhash: generateRandomString(16),
                        updated_on: faker.date.recent({ days: 30 }),
                        urls: {
                            original: `https://picsum.photos/1920/1080?random=${Math.floor(Math.random() * 100)}`,
                            thumbnail: `https://picsum.photos/800/600?random=${Math.floor(Math.random() * 100)}`,
                            preview: `https://picsum.photos/800/600?random=${Math.floor(Math.random() * 100)}`,
                            raw: `https://picsum.photos/1920/1080?random=${Math.floor(Math.random() * 100)}`
                        },
                        image_data: {
                            file_name: `${faker.word.noun()}.jpg`,
                            file_size: Math.floor(Math.random() * 1000000) + 100000, // Random size between 100KB and 1MB
                            original_file_name: `${faker.word.noun()}_original.jpg`,
                            file_type: "jpg",
                            keywords: faker.lorem.words(Math.floor(Math.random() * 15)).split(" "),
                            width: Math.floor(Math.random() * 1920) + 800, // Random width between 800 and 1920
                            height: Math.floor(Math.random() * 1080) + 600, // Random height between 600 and 1080
                            colorSpace: "sRGB"
                        }
                    })
                })
        );

        search.data.images = Array.from(
            { length: Math.floor(Math.random() * 90) + 54 },
            (_, i) =>
                new ImageObjectData({
                    id: generateRandomString(16),
                    name: `${faker.word.adjective()} ${faker.word.noun()}`,
                    collection_id: generateRandomString(16),
                    uploaded_on: faker.date.past({ years: 1 }),
                    uploaded_by: createTestUser(),
                    dupes: [],
                    private: false,
                    thumbhash: generateRandomString(16),
                    updated_on: faker.date.recent({ days: 30 }),
                    urls: {
                        original: `https://picsum.photos/1920/1080?random=${Math.floor(Math.random() * 100)}`,
                        thumbnail: `https://picsum.photos/800/600?random=${Math.floor(Math.random() * 100)}`,
                        preview: `https://picsum.photos/800/600?random=${Math.floor(Math.random() * 100)}`,
                        raw: `https://picsum.photos/1920/1080?random=${Math.floor(Math.random() * 100)}`
                    },
                    image_data: {
                        file_name: `${faker.word.noun()}.jpg`,
                        file_size: Math.floor(Math.random() * 1000000) + 100000, // Random size between 100KB and 1MB
                        original_file_name: `${faker.word.noun()}_original.jpg`,
                        file_type: "jpg",
                        keywords: faker.lorem.words(Math.floor(Math.random() * 15)).split(" "),
                        width: Math.floor(Math.random() * 1920) + 800, // Random width between 800 and 1920
                        height: Math.floor(Math.random() * 1080) + 600, // Random height between 600 and 1080
                        colorSpace: "sRGB"
                    }
                })
        );
    } catch (error) {
        console.error("Search failed:", error);
    } finally {
        search.loading = false;
    }
}