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

export interface ImageObjectData {
    id: string;
    name: string;
    uploaded_on: Date;
    uploaded_by: string;
    updated_on: Date;
    image_data: ImageData;
    collection_id: string;
    private?: boolean;
    dupes: ImageDupes[];
}

export interface ImageData {
    file_name: string;
    file_size: number;
    original_file_name: string;
    file_type: SupportedImageTypes | SupportedRAWFiles;
    keywords: string[];
    width: number;
    height: number;
};

export interface Collection {
    id: string;
    name: string;
    image_count: number;
    private?: boolean;
    images: ImageObjectData[];
    created_on: Date;
    updated_on: Date;
    created_by: string;
    description: string;
};

export interface ImageDupes {
    id: string;
    original_image_id: string;
    properties: ImageObjectData;
    created_on: Date;
}