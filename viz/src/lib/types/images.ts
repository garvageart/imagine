/**
 * ⚠️ DEPRECATION NOTICE ⚠️
 * 
 * The types in this file are being deprecated in favor of the auto-generated API types.
 * Please use types from './api-adapters' instead, which are generated from the OpenAPI spec
 * and guarantee synchronization with the backend API.
 * 
 * Migration guide:
 * - IImageObjectData → Use APIImage from api-adapters
 * - Collection → Use APICollection from api-adapters
 * - ImageData → Use APIImageMetadata from api-adapters
 * - ImagesResponse → Use APIImagesResponse from api-adapters
 * - CollectionResponse → Use APICollectionListResponse from api-adapters
 * - Pagination → Use APIPagination from api-adapters
 * 
 * This file will be removed in a future version.
 */

import type { User } from "./users";
import type { APIPagination } from "$lib/api/adapters";
import type CollectionData from "$lib/entities/collection";

export type SupportedImageTypes = "jpeg" | "jpg" | "png" | "tiff";
export const SUPPORTED_IMAGE_TYPES: SupportedImageTypes[] = [
    "jpeg",
    "jpg",
    "png",
    "tiff"
];

/**
 * Taken from https://docs.photoprism.app/developer-guide/media/raw/
 * 
 * Not the final supported files, this may eventually end up being removed
*/
export type SupportedRAWFiles = "3fr" | "ari" | "arw" | "bay" | "cap" | "cr2" | "cr3" | "crw" | "data" | "dcr" | "dcs" | "drf" | "eip" | "erf" | "fff" | "gpr" | "iiq" | "k25" | "kdc" | "mdc" | "mef" | "mos" | "mrw" | "nef" | "nrw" | "obm" | "orf" | "pef" | "ptx" | "pxn" | "r3d" | "raf" | "raw" | "rw2" | "rwl" | "rwz" | "sr2" | "srf" | "srw" | "x3f";
export const SUPPORTED_RAW_FILES: SupportedRAWFiles[] = [
    "3fr",
    "ari",
    "arw",
    "bay",
    "cap",
    "cr2",
    "cr3",
    "crw",
    "data",
    "dcr",
    "dcs",
    "drf",
    "eip",
    "erf",
    "fff",
    "gpr",
    "iiq",
    "k25",
    "kdc",
    "mdc",
    "mef",
    "mos",
    "mrw",
    "nef",
    "nrw",
    "obm",
    "orf",
    "pef",
    "ptx",
    "pxn",
    "r3d",
    "raf",
    "raw",
    "rw2",
    "rwl",
    "rwz",
    "sr2",
    "srf",
    "srw",
    "x3f"
];

/**
 * @deprecated Use ImageObjectData entity class instead
 */
export interface IImageObjectData {
    uid: string;
    name: string;
    uploaded_at: Date;
    uploaded_by: User;
    updated_at: Date;
    image_data: ImageData;
    collection_id: string;
    private: boolean;
    dupes?: ImageDupes[];
    thumbhash: string;
    urls: {
        original: string;
        thumbnail: string;
        preview: string;
        raw?: string;
    };
}

/**
 * @deprecated Use APIImageMetadata from api-adapters instead
 */
export interface ImageData {
    file_name: string;
    file_size: number;
    original_file_name: string;
    file_type: SupportedImageTypes | SupportedRAWFiles;
    keywords: string[];
    width: number;
    height: number;
    colorSpace: string;
};

/**
 * @deprecated Use CollectionData entity class instead
 */
export interface Collection {
    uid: string;
    name: string;
    image_count: number;
    private: boolean;
    images: IImageObjectData[];
    created_at: Date;
    updated_at: Date;
    created_by: User;
    description: string;
    owner: User;
    thumbnail?: IImageObjectData;
};

/**
 * @deprecated Not in API schema
 */
export interface ImageDupes {
    uid: string;
    original_image_id: string;
    properties: IImageObjectData;
    created_at: Date;
}

/**
 * @deprecated Use APIImagesResponse from api-adapters instead
 */
export type ImagesResponse<TImage = Record<string, any>> = TImage & {
    added_at: string; // time.Time → RFC3339 string
    added_by: string;
};

/**
 * @deprecated Use APICollectionListResponse from api-adapters instead
 */
export interface CollectionResponse extends APIPagination {
    items: CollectionData[];
}