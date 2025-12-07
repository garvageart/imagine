import { getUserSettings } from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
    const response = await getUserSettings();
    if (response.status !== 200) {
        error(response.status, {
            message: response.data.error
        });
    }

    return {
        settings: response.data,
        section: params.section || 'general'
    };
};
