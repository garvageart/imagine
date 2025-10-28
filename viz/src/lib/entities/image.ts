import type { Image, ImagesResponse, ImageMetadata, ImagePaths } from "$lib/api/client.gen";

export class ImageObjectData {
    uid: string;
    name: string;
    description?: string;
    uploaded_by?: string;
    image_metadata?: ImageMetadata;
    image_paths?: ImagePaths;
    private: boolean;
    width: number;
    height: number;
    processed: boolean;
    thumbhash?: string;
    created_at: Date;
    updated_at: Date;

    constructor(data: Partial<ImageObjectData> & Pick<ImageObjectData, 'uid' | 'name' | 'created_at' | 'updated_at'>) {
        this.uid = data.uid;
        this.name = data.name;
        this.description = data.description;
        this.uploaded_by = data.uploaded_by;
        this.image_metadata = data.image_metadata;
        this.image_paths = data.image_paths;
        this.private = data.private ?? false;
        this.width = data.width ?? 0;
        this.height = data.height ?? 0;
        this.processed = data.processed ?? false;
        this.thumbhash = data.thumbhash;
        this.created_at = data.created_at;
        this.updated_at = data.updated_at;
    }

    /**
     * Create ImageObjectData from API Image response
     */
    static fromAPI(apiImage: Image): ImageObjectData {
        return new ImageObjectData({
            uid: apiImage.uid,
            name: apiImage.name,
            description: apiImage.description,
            uploaded_by: apiImage.uploaded_by?.uid,
            image_metadata: apiImage.image_metadata,
            image_paths: apiImage.image_paths,
            private: apiImage.private,
            width: apiImage.width,
            height: apiImage.height,
            processed: apiImage.processed,
            thumbhash: apiImage.image_metadata?.thumbhash,
            created_at: new Date(apiImage.created_at),
            updated_at: new Date(apiImage.updated_at),
        });
    }
}