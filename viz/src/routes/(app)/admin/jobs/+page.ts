import { listAvailableWorkers } from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch }) => {
    const res = await listAvailableWorkers({ fetch });
    if (res.status === 200) {
        return {
            jobTypes: res.data.items.sort((a, b) => a.name.localeCompare(b.name))
        };
    }

    error(res.status, {
        message: res.data.error || "Failed to load job types"
    });
};
