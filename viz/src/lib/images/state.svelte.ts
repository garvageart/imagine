import type { Image, ImagesListResponse } from "$lib/api";

// Helper class for managing gallery state
// This ensures we can mutate state (append images) while still initializing from data
export class ImagePaginationState {
    images = $state<Image[]>([]);
    pagination = $state({ limit: 100, page: 0 });
    totalCount = $state(0);
    hasMore = $state(false);

    constructor(data: ImagesListResponse) {
        if (!data) {
            return;
        }
        
        this.images = data.items?.map((i) => i.image) ?? [];
        this.pagination = {
            limit: data.limit ?? 100,
            page: data.page ?? 0
        };

        this.totalCount = data.count ?? 0;
        this.hasMore = !!data.next;
    }
}