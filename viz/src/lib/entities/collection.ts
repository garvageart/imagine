import type { Collection, ImageObjectData } from "$lib/types/images";
import type { User } from "$lib/types/users";

class CollectionData implements Collection {
    id: string;
    name: string;
    image_count: number;
    private?: boolean;
    images: ImageObjectData[];
    created_on: Date;
    updated_on: Date;
    created_by: User;
    description: string;
    owner: User;
    thumbnail?: ImageObjectData;

    constructor(data: Partial<Collection>) {
        this.id = data.id || '';
        this.name = data.name || '';
        this.image_count = data.image_count || 0;
        this.private = data.private || false;
        this.images = data.images || [];
        this.created_on = data.created_on || new Date();
        this.updated_on = data.updated_on || new Date();
        this.created_by = data.created_by || { id: '', first_name: '', last_name: '', username: '', email: '', created_on: new Date(), updated_on: new Date(), role: 'user' };
        this.description = data.description || '';
        this.owner = data.owner || { id: '', first_name: '', last_name: '', username: '', email: '', created_on: new Date(), updated_on: new Date(), role: 'user' };
        this.thumbnail = data.thumbnail;

        for (const [key, value] of Object.entries(data)) {
            if (value === undefined || value === null) {
                console.warn(`Collection: Missing value for ${key}`);
            }
        }
    }
}

export default CollectionData;