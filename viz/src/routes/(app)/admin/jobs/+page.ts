import { listJobs } from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch }) => {
    try {
        const res = await listJobs({ fetch });
        if (res.status === 200) {
            return {
                jobTypes: res.data.items || []
            };
        }
        return {
            jobTypes: [],
            error: "Failed to fetch job types"
        };
    } catch (e) {
        return {
            jobTypes: [],
            error: (e as Error).message
        };
    }
};
