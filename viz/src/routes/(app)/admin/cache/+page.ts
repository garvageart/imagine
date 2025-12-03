import { getCacheStatus } from "$lib/api";
import { error } from "@sveltejs/kit";

export async function load() {
    const res = await getCacheStatus();
    if (res.status !== 200) {
        throw error(res.status, {
            message: res.data.error || "Failed to load cache status"
        });
    }

    return {
        cacheStatus: res.data
    };
}