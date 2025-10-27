import type { APIImage, APIImagesResponse, APIImageMetadata, APIImagePaths } from "$lib/types/api-adapters";

export class ImageObjectData {
    uid: string;
    name: string;
    description?: string;
    uploaded_by?: string;
    image_metadata?: APIImageMetadata;
    image_paths?: APIImagePaths;
    private: boolean;
    width: number;
    height: number;
    processed: boolean;
    thumbhash?: string;
    // Extended properties for UI/local use
    created_at: Date;
    updated_at: Date;
    added_at?: Date; // From ImagesResponse when in collection context
    added_by?: string; // From ImagesResponse when in collection context
    collection_id?: string; // Local tracking, not from API

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
        this.added_at = data.added_at;
        this.added_by = data.added_by;
        this.collection_id = data.collection_id;
    }

    /**
     * Create ImageObjectData from API Image response
     */
    static fromAPI(apiImage: APIImage): ImageObjectData {
        return new ImageObjectData({
            uid: apiImage.uid,
            name: apiImage.name,
            description: apiImage.description,
            uploaded_by: apiImage.uploaded_by,
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

    /**
     * Create ImageObjectData from API ImagesResponse (includes added_at timestamp)
     */
    static fromImagesResponse(response: APIImagesResponse): ImageObjectData {
        const imageData = ImageObjectData.fromAPI(response.image);
        imageData.added_at = new Date(response.added_at);
        imageData.added_by = response.added_by;
        return imageData;
    }

    /**
     * Get thumbnail URL
     */
    get thumbnailUrl(): string {
        return this.image_paths?.thumbnail_path
            ? `/images/${this.uid}/file?format=webp&w=800`
            : '';
    }

    /**
     * Get preview URL
     */
    get previewUrl(): string {
        return this.image_paths?.preview_path
            ? `/images/${this.uid}/file?format=webp`
            : '';
    }

    /**
     * Get original URL
     */
    get originalUrl(): string {
        return this.image_paths?.original_path
            ? `/images/${this.uid}/file`
            : '';
    }
}