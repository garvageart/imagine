import type { Collection } from "$lib/api";
import type { ImageObjectData } from "./image";

class CollectionData {
    uid: string;
    name: string;
    image_count: number;
    private: boolean;
    images: ImageObjectData[];
    created_by?: string;
    created_at: Date;
    updated_at: Date;
    description: string;
    thumbnail?: ImageObjectData;

    constructor(data: Partial<CollectionData> & Pick<CollectionData, 'uid' | 'name' | 'image_count' | 'created_at' | 'updated_at'>) {
        this.uid = data.uid;
        this.name = data.name;
        this.image_count = data.image_count;
        this.private = data.private ?? false;
        this.images = data.images ?? [];
        this.created_by = data.created_by;
        this.created_at = data.created_at;
        this.updated_at = data.updated_at;
        this.description = data.description ?? '';
        this.thumbnail = data.thumbnail;

        for (const [key, value] of Object.entries(data)) {
            if (value === undefined || value === null) {
                console.warn(`Collection: Missing value for ${key}`);
            }
        }
    }

    /**
     * Create a CollectionData instance from an API Collection response
     */
    static fromAPI(apiCollection: Collection): CollectionData {
        return new CollectionData({
            uid: apiCollection.uid,
            name: apiCollection.name,
            image_count: apiCollection.image_count,
            private: apiCollection.private ?? false,
            images: [], // Images are loaded separately or from CollectionDetailResponse
            created_by: apiCollection.created_by?.uid,
            created_at: new Date(apiCollection.created_at),
            updated_at: new Date(apiCollection.updated_at),
            description: apiCollection.description ?? '',
        });
    }
}

export default CollectionData;