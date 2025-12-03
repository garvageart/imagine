import type { PageLoad } from "./$types";
import { listCollections } from "$lib/api";
import { error } from "@sveltejs/kit";

export const load: PageLoad = async () => {
    const response = await listCollections();

    if (response.status !== 200) {
        error(response.status, {
            message: response.data.error || "Failed to load collections"
        });
    }

    return response.data;
};
