import { dev } from "$app/environment";
import { goto } from "$app/navigation";
import { page } from "$app/state";
import { executeSearch } from "$lib/api";
import { search } from "$lib/states/index.svelte";
import { sleep } from "$lib/utils/misc";
import { updateURLParameter } from "$lib/utils/url";

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
        const randomLatency = dev ? Math.floor(Math.random() * 200) + 100 : 0;
        await sleep(randomLatency);

        updateURLParameter("q", search.value);

        const res = await executeSearch(search.value, { limit: 100, page: 0 });
        if (res.status === 200) {
            search.data.images.data = res.data.images ?? [];
            search.data.collections.data = res.data.collections ?? [];
        } else {
            console.error("Search failed:", res.status);
            search.data.images.data = [];
            search.data.collections.data = [];
        }

    } catch (error) {
        console.error("Search failed:", error);
    } finally {
        search.loading = false;
    }
}