import type { Collection, IImageObjectData } from "$lib/types/images";
import type { User } from "$lib/types/users";

class CollectionData implements Collection {
    id: string;
    name: string;
    image_count: number;
    private: boolean;
    images: IImageObjectData[];
    created_on: Date;
    updated_on: Date;
    created_by: User;
    description: string;
    owner: User;
    thumbnail?: IImageObjectData;

    constructor(data: Collection) {
        this.id = data.id;
        this.name = data.name;
        this.image_count = data.image_count;
        this.private = data.private;
        this.images = data.images;
        this.created_on = data.created_on;
        this.updated_on = data.updated_on;
        this.created_by = data.created_by;
        this.description = data.description;
        this.owner = data.owner;
        this.thumbnail = data.thumbnail;

        for (const [key, value] of Object.entries(data)) {
            if (value === undefined || value === null) {
                console.warn(`Collection: Missing value for ${key}`);
            }
        }
    }
}

export default CollectionData;