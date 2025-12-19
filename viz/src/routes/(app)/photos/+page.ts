import type { PageLoad } from './$types';
import { listImages, type ImagesListResponse } from '$lib/api';
import { error } from "@sveltejs/kit";

export const load: PageLoad = async ({ url }) => {
    const limit = parseInt(url.searchParams.get('limit') || '100', 10);
    const page = parseInt(url.searchParams.get('page') || '0', 10);

    const response = await listImages({ limit, page });

    if (response.status === 200) {
        const data = response.data as ImagesListResponse;
        return {
            images: data.items?.map((item) => item.image) || [],
            count: data.count || 0,
            limit,
            page,
            next: data.next || null,
            prev: data.prev || null
        };
    }

    error(response.status, {
        message: response.data.error || "Failed to load images"
    });
};