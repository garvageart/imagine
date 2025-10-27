import { dev } from "$app/environment";
import { goto } from "$app/navigation";
import { page } from "$app/state";
import { createTestUser, createTestImageObject, createTestCollection } from "$lib/data/test";
import CollectionData from "$lib/entities/collection";
import { ImageObjectData } from "$lib/entities/image";
import { search } from "$lib/states/index.svelte";
import { generateRandomString, sleep } from "$lib/utils/misc";
import { updateURLParameter } from "$lib/utils/url";
import { faker } from "@faker-js/faker";

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

        // Generate mock collections
        search.data.collections.data = Array.from(
            { length: Math.floor(Math.random() * 45) + 15 },
            () => createTestCollection()
        );

        // Generate mock images
        search.data.images.data = Array.from(
            { length: Math.floor(Math.random() * 90) + 54 },
            () => createTestImageObject()
        );
    } catch (error) {
        console.error("Search failed:", error);
    } finally {
        search.loading = false;
    }
}