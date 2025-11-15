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
export type ErrorResponse = {
    error: string;
};
export type ApiKey = {
    uid: string;
    name?: string;
    description?: string | null;
    key_hashed: string;
    user?: User;
    scopes?: string[];
    last_used_at?: string | null;
    revoked: boolean;
    revoked_at?: string;
    expires_at?: string | null;
    created_at: string;
    updated_at: string;
};
export type ApiKeyCreate = {
    name?: string;
    description?: string | null;
    scopes?: string[];
    expires_at?: string | null;
};
export type ApiKeyCreateResponse = {
    consumer_key: string;
    expires_at?: string | null;
};
export type ApiKeyListResponse = {
    items: ApiKey[];
    count: number;
};
export type MessageResponse = {
    message: string;
};
export type OAuthUserData = {
    email: string;
    name: string;
    picture: string;
};
export type Session = {
    uid: string;
    token: string;
    user_uid: string;
    user?: User;
    client_id?: string;
    client_name?: string;
    client_ip?: string;
    last_active?: string;
    expires_at?: string;
    timeout?: number;
    user_agent?: string;
    ref_id?: string;
    login_ip?: string;
    login_at?: string;
    status?: number;
    created_at: string;
    updated_at: string;
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
    exposure_value?: string;
    f_number?: string;
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
    metadata?: string;
    /** User-assigned rating (0-5). Null = unrated */
    rating?: number | null;
    keywords?: string[];
    color_space: string;
    file_modified_at: string;
    file_created_at: string;
    thumbhash?: string;
    label?: string;
    checksum: string;
};
export type ImagePaths = {
    original: string;
    thumbnail: string;
    preview: string;
    raw?: string;
};
export type Image = {
    uid: string;
    name: string;
    uploaded_by?: User;
    description?: string;
    exif?: ImageExif;
    "private": boolean;
    width: number;
    height: number;
    processed: boolean;
    image_metadata?: ImageMetadata;
    image_paths: ImagePaths;
    created_at: string;
    updated_at: string;
    taken_at?: string | null;
};
export type ImagesResponse = {
    added_at: string;
    added_by?: User;
    image: Image;
};
export type ImagesPage = {
    href?: string;
    prev?: string;
    next?: string;
    limit: number;
    page: number;
    count?: number;
    items: ImagesResponse[];
};
export type ImageUploadResponse = {
    id: string;
};
export type DeleteAssetsResponse = {
    results: {
        uid: string;
        deleted: boolean;
        error?: string;
    }[];
    message?: string;
};
export type ImageUpdate = {
    name?: string;
    description?: string | null;
    "private"?: boolean;
    exif?: ImageExif;
    image_metadata?: {
        label?: string | null;
        rating?: number | null;
        keywords?: string[];
    };
};
export type CollectionImage = {
    uid: string;
    added_at: string;
    added_by?: User;
};
export type Collection = {
    uid: string;
    name: string;
    image_count: number;
    "private"?: boolean | null;
    images?: CollectionImage[];
    created_by?: User;
    description?: string;
    thumbnail?: Image;
    created_at: string;
    updated_at: string;
};
export type CollectionListResponse = {
    href?: string;
    prev?: string;
    next?: string;
    limit: number;
    page: number;
    count?: number;
    items: Collection[];
};
export type CollectionCreate = {
    name: string;
    "private"?: boolean | null;
    description?: string;
};
export type CollectionDetailResponse = {
    uid: string;
    name: string;
    image_count?: number;
    "private"?: boolean | null;
    images: ImagesPage;
    created_by?: User;
    description?: string;
    thumbnail?: Image;
    created_at: string;
    updated_at: string;
};
export type CollectionUpdate = {
    name?: string;
    thumbnailUID?: string;
    description?: string;
    "private"?: boolean;
    ownerUID?: string;
};
export type AddImagesResponse = {
    added: boolean;
    error?: string;
};
export type DeleteImagesResponse = {
    deleted: boolean;
    error?: string;
};
export type DownloadRequest = {
    uids: string[];
    filename?: string;
};
export type SignDownloadRequest = {
    /** Array of image UIDs to include in the download token */
    uids?: string[];
    /** Time in seconds until the token expires (0 for no expiry, default 900 = 15 minutes) */
    expires_in?: number;
    /** Allow downloads using this token (default true) */
    allow_download?: boolean;
    /** Allow embedding images on external sites (default false to prevent hotlinking) */
    allow_embed?: boolean;
    /** Include EXIF and other metadata in responses (default true) */
    show_metadata?: boolean;
    /** Optional password protection for the token (will be bcrypt hashed) */
    password?: string;
    /** Optional description of this share/download link */
    description?: string;
};
export type DownloadToken = {
    /** 64-character hex token that serves as both unique identifier and authorization key */
    uid: string;
    /** Array of authorized image UIDs */
    image_uids: string[];
    /** Whether downloads are permitted with this token */
    allow_download: boolean;
    /** Whether embedding on external sites is allowed (false prevents hotlinking) */
    allow_embed: boolean;
    /** Whether to include EXIF and metadata in responses */
    show_metadata: boolean;
    /** Optional bcrypt hash of password (null if no password protection) */
    password?: string | null;
    /** Optional description of this download link */
    description?: string | null;
    /** When this token expires (null for no expiry) */
    expires_at?: string | null;
    /** When this token was created */
    created_at: string;
    /** When this token was last updated */
    updated_at: string;
};
export type JobInfo = {
    id: string;
    topic: string;
    status: string;
};
export type JobListResponse = {
    items: JobInfo[];
};
export type JobCreateRequest = {
    /** Job type (e.g., thumbnailGeneration, imageProcessing) */
    "type": string;
    /** Command to execute (all=process all, missing=process missing, single=process one image) */
    command?: "all" | "missing" | "single";
    /** Image UID (required when command=single) */
    image_uid?: string;
};
export type JobEnqueueResponse = {
    message: string;
    count?: number;
};
export type WorkerJob = {
    uid: string;
    "type": string;
    topic: string;
    command?: string | null;
    image_uid?: string | null;
    status: string;
    payload?: string | null;
    error_code?: string | null;
    error_msg?: string | null;
    enqueued_at: string;
    started_at?: string | null;
    completed_at?: string | null;
};
export type JobCountResponse = {
    running: number;
};
export type JobStatsResponse = {
    running: number;
    running_by_topic: {
        [key: string]: number;
    };
    queued_by_topic: {
        [key: string]: number;
    };
};
export type WsStatsResponse = {
    connectedClients: number;
    clientIds: string[];
    timestamp: string;
};
export type WsMetricsResponse = {
    connectedClients: number;
    totalEvents: number;
    eventsByType: {
        [key: string]: number;
    };
    timestamp: string;
};
export type EventRecord = {
    timestamp: string;
    event: string;
    data: {
        [key: string]: any;
    };
};
export type EventHistoryResponse = {
    events: EventRecord[];
    count: number;
};
export type WsBroadcastRequest = {
    event: string;
    data: {
        [key: string]: any;
    };
};
export type WsBroadcastResponse = {
    success: boolean;
    message: string;
    clients: number;
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
    } | {
        status: 400;
        data: ErrorResponse;
    }>("/accounts/", oazapfts.json({
        ...opts,
        method: "POST",
        body: userCreate
    }));
}
/**
 * Generate an API key
 */
export function generateApiKey(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ApiKey;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/auth/apikey", {
        ...opts
    });
}
/**
 * Create a new API key
 */
export function createApiKey(apiKeyCreate: ApiKeyCreate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: ApiKeyCreateResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    }>("/api-keys", oazapfts.json({
        ...opts,
        method: "POST",
        body: apiKeyCreate
    }));
}
/**
 * List API keys for the authenticated user
 */
export function listApiKeys(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ApiKeyListResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/api-keys", {
        ...opts
    });
}
/**
 * Get API key details
 */
export function getApiKey(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ApiKey;
    } | {
        status: 404;
        data: ErrorResponse;
    }>(`/api-keys/${encodeURIComponent(uid)}`, {
        ...opts
    });
}
/**
 * Delete an API key
 */
export function deleteApiKey(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>(`/api-keys/${encodeURIComponent(uid)}`, {
        ...opts,
        method: "DELETE"
    });
}
/**
 * Revoke an API key
 */
export function revokeApiKey(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    }>(`/api-keys/${encodeURIComponent(uid)}/revoke`, {
        ...opts,
        method: "POST"
    });
}
/**
 * Rotate an API key (revoke old, create new)
 */
export function rotateApiKey(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: ApiKeyCreateResponse;
    }>(`/api-keys/${encodeURIComponent(uid)}/rotate`, {
        ...opts,
        method: "POST"
    });
}
/**
 * Login with email and password
 */
export function login(body: {
    email: string;
    password: string;
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/auth/login", oazapfts.json({
        ...opts,
        method: "POST",
        body
    }));
}
/**
 * Initiate OAuth flow
 */
export function initiateOAuth(provider: "google" | "github", opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 307;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/auth/oauth${QS.query(QS.explode({
        provider
    }))}`, {
        ...opts
    });
}
/**
 * Complete OAuth flow
 */
export function completeOAuth(provider: "google" | "github", code: string, state: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: OAuthUserData;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/auth/oauth/${encodeURIComponent(provider)}${QS.query(QS.explode({
        code,
        state
    }))}`, {
        ...opts,
        method: "POST"
    });
}
/**
 * Get current session information
 */
export function getCurrentSession(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Session;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/auth/session", {
        ...opts
    });
}
/**
 * Logout current session
 */
export function logout(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/auth/logout", {
        ...opts,
        method: "POST"
    });
}
/**
 * Get current authenticated user
 */
export function getCurrentUser(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: User;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/accounts/me", {
        ...opts
    });
}
/**
 * List all images with pagination
 */
export function listImages({ limit, page }: {
    limit?: number;
    page?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ImagesPage;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/images${QS.query(QS.explode({
        limit,
        page
    }))}`, {
        ...opts
    });
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
        data: ImageUploadResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/images", oazapfts.multipart({
        ...opts,
        method: "POST",
        body
    }));
}
/**
 * Delete multiple asset UID directories (soft move to trash or force delete)
 */
export function deleteImagesBulk(body: {
    uids: string[];
    force?: boolean;
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: DeleteAssetsResponse;
    } | {
        status: 207;
        data: DeleteAssetsResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/images", oazapfts.json({
        ...opts,
        method: "DELETE",
        body
    }));
}
/**
 * Upload an image by URL
 */
export function uploadImageByUrl(body: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: ImageUploadResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/images/url", {
        ...opts,
        method: "POST",
        body
    });
}
/**
 * Get a processed image file
 */
export function getImageFile(uid: string, { format, w, h, quality, download, token, password }: {
    format?: "webp" | "png" | "jpg" | "jpeg" | "avif" | "heif";
    w?: number;
    h?: number;
    quality?: number;
    download?: "1";
    token?: string;
    password?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Blob;
    } | {
        status: 304;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    }>(`/images/${encodeURIComponent(uid)}/file${QS.query(QS.explode({
        format,
        w,
        h,
        quality,
        download,
        token,
        password
    }))}`, {
        ...opts
    });
}
/**
 * Get EXIF data for an image
 */
export function getImageExif(uid: string, { simple }: {
    simple?: boolean;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ImageExif;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/images/${encodeURIComponent(uid)}/exif${QS.query(QS.explode({
        simple
    }))}`, {
        ...opts
    });
}
/**
 * Update image metadata
 */
export function updateImage(uid: string, imageUpdate: ImageUpdate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Image;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/images/${encodeURIComponent(uid)}`, oazapfts.json({
        ...opts,
        method: "PATCH",
        body: imageUpdate
    }));
}
/**
 * Create short-lived download token and redirect
 */
export function quickDownloadImage(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 302;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/images/${encodeURIComponent(uid)}/download`, {
        ...opts
    });
}
/**
 * List collections
 */
export function listCollections({ limit, page }: {
    limit?: number;
    page?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: CollectionListResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/collections${QS.query(QS.explode({
        limit,
        page
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
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
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
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/collections/${encodeURIComponent(uid)}`, {
        ...opts
    });
}
/**
 * Update collection
 */
export function updateCollection(uid: string, collectionUpdate: CollectionUpdate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Collection;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/collections/${encodeURIComponent(uid)}`, oazapfts.json({
        ...opts,
        method: "PATCH",
        body: collectionUpdate
    }));
}
/**
 * Delete a collection
 */
export function deleteCollection(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 204;
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/collections/${encodeURIComponent(uid)}`, {
        ...opts,
        method: "DELETE"
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
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
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
        data: AddImagesResponse;
    } | {
        status: 400;
        data: AddImagesResponse;
    } | {
        status: 404;
        data: AddImagesResponse;
    }>(`/collections/${encodeURIComponent(uid)}/images`, oazapfts.json({
        ...opts,
        method: "PUT",
        body
    }));
}
/**
 * Remove images from a collection
 */
export function deleteCollectionImages(uid: string, body: {
    uids: string[];
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: DeleteImagesResponse;
    } | {
        status: 400;
        data: DeleteImagesResponse;
    } | {
        status: 404;
        data: DeleteImagesResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/collections/${encodeURIComponent(uid)}/images`, oazapfts.json({
        ...opts,
        method: "DELETE",
        body
    }));
}
/**
 * Download a set of images as a ZIP (requires token)
 */
export function downloadImages(token: string, downloadRequest: DownloadRequest, { password }: {
    password?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Blob;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/download${QS.query(QS.explode({
        token,
        password
    }))}`, oazapfts.json({
        ...opts,
        method: "POST",
        body: downloadRequest
    }));
}
/**
 * Create a download token with optional access controls
 */
export function signDownload(signDownloadRequest?: SignDownloadRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: DownloadToken;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/download/sign", oazapfts.json({
        ...opts,
        method: "POST",
        body: signDownloadRequest
    }));
}
/**
 * Admin-only healthcheck
 */
export function adminHealthcheck(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/admin/healthcheck", {
        ...opts,
        method: "POST"
    });
}
/**
 * List jobs
 */
export function listJobs(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: JobListResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/jobs", {
        ...opts
    });
}
/**
 * Create/enqueue a job
 */
export function createJob(jobCreateRequest: JobCreateRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 202;
        data: JobEnqueueResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/jobs", oazapfts.json({
        ...opts,
        method: "POST",
        body: jobCreateRequest
    }));
}
/**
 * Get job detail
 */
export function getJob(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WorkerJob;
    } | {
        status: 404;
        data: ErrorResponse;
    }>(`/jobs/${encodeURIComponent(uid)}`, {
        ...opts
    });
}
/**
 * Cancel job
 */
export function cancelJob(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    }>(`/jobs/${encodeURIComponent(uid)}`, {
        ...opts,
        method: "DELETE"
    });
}
/**
 * Retry job
 */
export function retryJob(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 501;
        data: ErrorResponse;
    }>(`/jobs/${encodeURIComponent(uid)}`, {
        ...opts,
        method: "POST"
    });
}
/**
 * Start job scheduler
 */
export function startScheduler(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/jobs/scheduler/start", {
        ...opts,
        method: "POST"
    });
}
/**
 * Shutdown job scheduler
 */
export function shutdownScheduler(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/jobs/scheduler/shutdown", {
        ...opts,
        method: "POST"
    });
}
/**
 * Get running jobs count
 */
export function getJobsCount(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: JobCountResponse;
    }>("/jobs/count", {
        ...opts
    });
}
/**
 * Get job stats by topic
 */
export function getJobStats(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: JobStatsResponse;
    }>("/jobs/stats", {
        ...opts
    });
}
/**
 * Stop a job type
 */
export function stopJobType($type: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/jobs/types/${encodeURIComponent($type)}/stop`, {
        ...opts,
        method: "POST"
    });
}
/**
 * Update job type concurrency
 */
export function updateJobTypeConcurrency($type: string, body: {
    concurrency: number;
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/jobs/types/${encodeURIComponent($type)}/concurrency`, oazapfts.json({
        ...opts,
        method: "PUT",
        body
    }));
}
/**
 * WebSocket connection for real-time updates
 */
export function connectWebSocket(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 101;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/events", {
        ...opts
    });
}
/**
 * Get WebSocket connection statistics
 */
export function getWsStats(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WsStatsResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/events/stats", {
        ...opts
    });
}
/**
 * Get WebSocket event metrics
 */
export function getWsMetrics(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WsMetricsResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/events/metrics", {
        ...opts
    });
}
/**
 * Get recent event history
 */
export function getEventHistory({ limit }: {
    limit?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: EventHistoryResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>(`/events/history${QS.query(QS.explode({
        limit
    }))}`, {
        ...opts
    });
}
/**
 * Clear event history
 */
export function clearEventHistory(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/events/history", {
        ...opts,
        method: "DELETE"
    });
}
/**
 * Get events since a cursor ID
 */
export function getEventsSince({ cursor, limit }: {
    cursor?: number;
    limit?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: {
            events: EventRecord[];
            count: number;
            nextCursor: number;
        };
    } | {
        status: 401;
        data: ErrorResponse;
    }>(`/events/since${QS.query(QS.explode({
        cursor,
        limit
    }))}`, {
        ...opts
    });
}
/**
 * Broadcast event to all connected WebSocket clients
 */
export function broadcastWsEvent(wsBroadcastRequest: WsBroadcastRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WsBroadcastResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/events/broadcast", oazapfts.json({
        ...opts,
        method: "POST",
        body: wsBroadcastRequest
    }));
}
/**
 * Send event to specific WebSocket client
 */
export function sendToWsClient(clientId: string, wsBroadcastRequest: WsBroadcastRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: {
            success: boolean;
            message: string;
            clientId: string;
        };
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/events/send/${encodeURIComponent(clientId)}`, oazapfts.json({
        ...opts,
        method: "POST",
        body: wsBroadcastRequest
    }));
}
