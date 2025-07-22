import type { User } from "./users";

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

export interface IImageObjectData {
    id: string;
    name: string;
    uploaded_on: Date;
    uploaded_by: User;
    updated_on: Date;
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

export interface Collection {
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
};

export interface ImageDupes {
    id: string;
    original_image_id: string;
    properties: IImageObjectData;
    created_on: Date;
}