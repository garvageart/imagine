/**
 * Imagine API
 * 0.1.0
 * DO NOT MODIFY - This file has been generated using oazapfts.
 * See https://www.npmjs.com/package/oazapfts
 */
import * as Oazapfts from "@oazapfts/runtime";
import * as QS from "@oazapfts/runtime/query";
export const defaults: Oazapfts.Defaults<Oazapfts.CustomHeaders> = {
    headers: {},
    baseUrl: "http://localhost:7770",
};
const oazapfts = Oazapfts.runtime(defaults);
export const servers = {
    localApi: "http://localhost:7770"
};
export type UserCreate = {
    name: string;
    email: string;
    password: string;
    passwordConfirm: string;
};
export type User = {
    uid: string;
    first_name: string;
    last_name: string;
    username: string;
    email: string;
    role: "user" | "admin" | "superadmin" | "guest";
    created_at: string;
    updated_at: string;
};
export type CollectionImage = {
    uid: string;
    added_at: string;
    added_by?: string;
};
export type Collection = {
    uid: string;
    name: string;
    image_count: number;
    "private"?: boolean | null;
    images?: CollectionImage[];
    created_by?: string;
    description?: string;
    created_at: string;
    updated_at: string;
};
export type CollectionListResponse = {
    href?: string;
    prev?: string;
    next?: string;
    limit: number;
    offset: number;
    count?: number;
    items: Collection[];
};
export type CollectionCreate = {
    name: string;
    "private"?: boolean | null;
    description?: string;
};
export type ImageExif = {
    exif_version?: string;
    make?: string;
    model?: string;
    date_time?: string;
    date_time_original?: string;
    iso?: string;
    focal_length?: string;
    exposure_time?: string;
    aperture?: string;
    flash?: string;
    white_balance?: string;
    lens_model?: string;
    modify_date?: string;
    rating?: string;
    orientation?: string;
    resolution?: string;
    software?: string;
    longitude?: string;
    latitude?: string;
};
export type ImageMetadata = {
    file_name: string;
    file_size?: number;
    original_file_name?: string;
    file_type: string;
    keywords?: string[];
    color_space: string;
    file_modified_at: string;
    file_created_at: string;
    thumbhash?: string;
    label?: string;
    checksum: string;
};
export type ImagePaths = {
    original_path: string;
    thumbnail_path: string;
    preview_path: string;
    raw_path?: string;
};
export type Image = {
    uid: string;
    name: string;
    uploaded_by?: string;
    description?: string;
    exif?: ImageExif;
    "private": boolean;
    width: number;
    height: number;
    processed: boolean;
    image_metadata?: ImageMetadata;
    image_paths?: ImagePaths;
    created_at: string;
    updated_at: string;
};
export type ImagesResponse = {
    added_at: string;
    added_by?: string;
    image: Image;
};
export type ImagesPage = {
    href?: string;
    prev?: string;
    next?: string;
    limit: number;
    offset: number;
    count?: number;
    items: ImagesResponse[];
};
export type CollectionDetailResponse = {
    uid: string;
    name: string;
    image_count?: number;
    "private"?: boolean | null;
    images: ImagesPage;
    created_by?: string;
    description?: string;
};
/**
 * Health ping
 */
export function ping(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: {
            message: string;
        };
    }>("/ping", {
        ...opts
    });
}
/**
 * Register a new user
 */
export function registerUser(userCreate: UserCreate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: User;
    }>("/user", oazapfts.json({
        ...opts,
        method: "POST",
        body: userCreate
    }));
}
/**
 * Upload an image (multipart)
 */
export function uploadImage(body: {
    filename: string;
    data: Blob;
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: {
            id: string;
        };
    }>("/images", oazapfts.multipart({
        ...opts,
        method: "POST",
        body
    }));
}
/**
 * Upload an image by URL
 */
export function uploadImageByUrl(body: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: {
            id: string;
        };
    }>("/images/url", {
        ...opts,
        method: "POST",
        body
    });
}
/**
 * Get a processed image file
 */
export function getImageFile(uid: string, { format, w, h, quality }: {
    format?: "webp" | "png" | "jpg" | "jpeg" | "avif" | "heif";
    w?: number;
    h?: number;
    quality?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchBlob<{
        status: 200;
        data: Blob;
    }>(`/images/${encodeURIComponent(uid)}/file${QS.query(QS.explode({
        format,
        w,
        h,
        quality
    }))}`, {
        ...opts
    });
}
/**
 * List collections
 */
export function listCollections({ limit, offset }: {
    limit?: number;
    offset?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: CollectionListResponse;
    }>(`/collections${QS.query(QS.explode({
        limit,
        offset
    }))}`, {
        ...opts
    });
}
/**
 * Create a collection
 */
export function createCollection(collectionCreate: CollectionCreate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: Collection;
    }>("/collections", oazapfts.json({
        ...opts,
        method: "POST",
        body: collectionCreate
    }));
}
/**
 * Get collection detail
 */
export function getCollection(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: CollectionDetailResponse;
    }>(`/collections/${encodeURIComponent(uid)}`, {
        ...opts
    });
}
/**
 * List images in a collection
 */
export function listCollectionImages(uid: string, { limit, offset }: {
    limit?: number;
    offset?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ImagesPage;
    }>(`/collections/${encodeURIComponent(uid)}/images${QS.query(QS.explode({
        limit,
        offset
    }))}`, {
        ...opts
    });
}
/**
 * Add images to a collection
 */
export function addCollectionImages(uid: string, body: {
    uids: string[];
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: {
            added: boolean;
        };
    }>(`/collections/${encodeURIComponent(uid)}/images`, oazapfts.json({
        ...opts,
        method: "PUT",
        body
    }));
}
