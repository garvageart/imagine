import { getCacheStatus } from "$lib/api";
import { error } from "@sveltejs/kit";

export async function load() {
    const res = await getCacheStatus();
    if (res.status !== 200) {
        throw error(res.status, res.data.error || 'Failed to fetch cache status');
    }

    console.log(res.data);

    return {
        cacheStatus: res.data
    };
}