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
    localApi: ({ host = "localhost", port = "7770" }: {
        host: string | number | boolean;
        port: string | number | boolean;
    }) => `http://${host}:${port}`,
    productionApi: "/api"
};
export type UserCreate = {
    /** Full name */
    name: string;
    /** Email address */
    email: string;
    /** Password */
    password: string;
};
export type User = {
    /** User UID */
    uid: string;
    /** First name */
    first_name: string;
    /** Last name */
    last_name: string;
    /** Username */
    username: string;
    /** Email */
    email: string;
    /** User role */
    role: "user" | "admin" | "superadmin" | "guest";
    /** Creation time */
    created_at: string;
    /** Update time */
    updated_at: string;
};
export type ErrorResponse = {
    /** Error message */
    error: string;
};
export type ApiKey = {
    /** API Key UID */
    uid: string;
    /** API Key name */
    name?: string;
    /** API Key description */
    description?: string | null;
    /** Hashed key */
    key_hashed: string;
    user?: User;
    /** List of scopes */
    scopes: string[];
    /** Last used time */
    last_used_at?: string | null;
    /** Is revoked */
    revoked: boolean;
    /** Revocation time */
    revoked_at?: string;
    /** Expiry time */
    expires_at?: string | null;
    /** Creation time */
    created_at: string;
    /** Update time */
    updated_at: string;
};
export type ApiKeyCreate = {
    /** API Key name */
    name?: string;
    /** API Key description */
    description?: string | null;
    /** List of scopes */
    scopes?: string[];
    /** Expiry time */
    expires_at?: string | null;
};
export type ApiKeyCreateResponse = {
    /** The consumer key (secret) */
    consumer_key: string;
    /** Expiry time */
    expires_at?: string | null;
};
export type ApiKeyListResponse = {
    /** List of API keys */
    items: ApiKey[];
    /** Total count */
    count: number;
};
export type MessageResponse = {
    /** Response message */
    message: string;
};
export type OAuthUserData = {
    /** User email */
    email: string;
    /** User name */
    name: string;
    /** User profile picture URL */
    picture: string;
};
export type Session = {
    /** Session UID */
    uid: string;
    /** Session token */
    token: string;
    /** User UID */
    user_uid: string;
    user?: User;
    /** Client ID */
    client_id?: string;
    /** Client name */
    client_name?: string;
    /** Client IP */
    client_ip?: string;
    /** Last active time */
    last_active?: string;
    /** Expiry time */
    expires_at?: string;
    /** Timeout in seconds */
    timeout?: number;
    /** User agent */
    user_agent?: string;
    /** Reference ID */
    ref_id?: string;
    /** Login IP */
    login_ip?: string;
    /** Login time */
    login_at?: string;
    /** Session status */
    status?: number;
    /** Creation time */
    created_at: string;
    /** Update time */
    updated_at: string;
};
export type UserUpdate = {
    /** First name */
    first_name?: string | null;
    /** Last name */
    last_name?: string | null;
    /** Username */
    username?: string | null;
    /** Email address */
    email?: string | null;
};
export type UserSetting = {
    /** Setting unique name */
    name: string;
    /** Readable name */
    display_name: string;
    /** The effective value (override if exists, else default). */
    value: string;
    /** Default value */
    default_value: string;
    /** Type of the value */
    value_type: string;
    /** Allowed values if enum */
    allowed_values?: string[] | null;
    /** Whether user can edit */
    is_user_editable?: boolean;
    /** Setting group */
    group: string;
    /** Setting description */
    description: string;
};
export type UserOnboardingBody = {
    /** First name */
    first_name: string;
    /** Last name */
    last_name: string;
    /** User-specific setting overrides */
    settings: UserSetting[];
};
export type UserPasswordUpdate = {
    /** Current password */
    current: string;
    /** New password */
    "new": string;
};
export type SessionUpdate = {
    /** New client name */
    clientName?: string | null;
    /** New status */
    status?: number | null;
};
export type UserSettingUpdateRequest = {
    /** List of settings to update */
    settings: {
        /** Setting name */
        name: string;
        /** Setting value */
        value: string;
    }[];
};
export type ImageExif = {
    /** EXIF version */
    exif_version?: string;
    /** Camera make */
    make?: string;
    /** Camera model */
    model?: string;
    /** Date and time */
    date_time?: string;
    /** Original date and time */
    date_time_original?: string;
    /** ISO speed rating */
    iso?: string;
    /** Focal length */
    focal_length?: string;
    /** Exposure time */
    exposure_time?: string;
    /** Aperture */
    aperture?: string;
    /** Exposure value */
    exposure_value?: string;
    /** F-number */
    f_number?: string;
    /** Flash fired */
    flash?: number;
    /** White balance */
    white_balance?: string;
    /** Lens make */
    lens_make?: string;
    /** Lens model */
    lens_model?: string;
    /** Focal length in 35mm format */
    focal_length_in_35mm_format?: string;
    /** Scene capture type */
    scene_capture_type?: string;
    /** Exposure program */
    exposure_program?: string;
    /** Metering mode */
    metering_mode?: string;
    /** Sensing method */
    sensing_method?: string;
    /** Light source */
    light_source?: string;
    /** Exposure bias value */
    exposure_bias_value?: string;
    /** Max aperture value */
    max_aperture_value?: string;
    /** Exposure mode */
    exposure_mode?: string;
    /** Digital zoom ratio */
    digital_zoom_ratio?: string;
    /** Focal plane X resolution */
    focal_plane_x_resolution?: string;
    /** Focal plane Y resolution */
    focal_plane_y_resolution?: string;
    /** Focal plane resolution unit */
    focal_plane_resolution_unit?: string;
    /** Copyright */
    copyright?: string;
    /** Modify date */
    modify_date?: string;
    /** Rating */
    rating?: string;
    /** Orientation */
    orientation?: string;
    /** Resolution */
    resolution?: string;
    /** Software */
    software?: string;
    /** GPS Longitude */
    longitude?: string;
    /** GPS Latitude */
    latitude?: string;
    /** GPS Altitude */
    gps_altitude?: string;
    /** GPS Image Direction */
    gps_img_direction?: string;
    /** GPS Image Direction Ref */
    gps_img_direction_ref?: string;
    /** GPS Speed */
    gps_speed?: string;
    /** GPS Speed Ref */
    gps_speed_ref?: string;
    /** Offset time */
    offset_time?: string;
    /** Offset time original */
    offset_time_original?: string;
    /** Offset time digitized */
    offset_time_digitized?: string;
};
export type ImageMetadata = {
    /** Original file name */
    file_name: string;
    /** File size in bytes */
    file_size?: number;
    /** Original file name */
    original_file_name?: string;
    /** File MIME type */
    file_type: string;
    /** Additional metadata */
    metadata?: string;
    /** User-assigned rating (0-5). Null = unrated */
    rating?: number | null;
    /** Keywords */
    keywords?: string[];
    /** Color space */
    color_space: string;
    /** Has ICC profile */
    has_icc_profile?: boolean;
    /** File modification time */
    file_modified_at: string;
    /** File creation time */
    file_created_at: string;
    /** Thumbhash */
    thumbhash?: string;
    /** User-assigned label for the image. Null = unlabeled */
    label?: ("Red" | "Orange" | "Yellow" | "Purple" | "Pink" | "Green" | "Blue" | "None") | null;
    /** File checksum */
    checksum: string;
};
export type ImagePaths = {
    /** Path to original file */
    original: string;
    /** Path to thumbnail */
    thumbnail: string;
    /** Path to preview */
    preview: string;
    /** Path to raw file */
    raw?: string;
};
export type Image = {
    /** Image UID */
    uid: string;
    /** Image name */
    name: string;
    uploaded_by?: User;
    owner?: User;
    /** Image description */
    description?: string;
    exif?: ImageExif;
    /** Is private */
    "private": boolean;
    /** Is favourited */
    favourited?: boolean;
    /** Image width */
    width: number;
    /** Image height */
    height: number;
    /** Is processed */
    processed: boolean;
    image_metadata?: ImageMetadata;
    image_paths: ImagePaths;
    /** Creation time */
    created_at: string;
    /** Update time */
    updated_at: string;
    /** Taken time */
    taken_at?: string | null;
};
export type CollectionImage = {
    /** Image UID */
    uid: string;
    /** Added timestamp */
    added_at: string;
    added_by?: User;
};
export type Collection = {
    /** Collection UID */
    uid: string;
    /** Collection name */
    name: string;
    /** Number of images */
    image_count: number;
    /** Is private */
    "private"?: boolean | null;
    /** Is favourited */
    favourited?: boolean;
    /** List of images */
    images?: CollectionImage[];
    created_by?: User;
    owner?: User;
    /** Collection description */
    description?: string;
    thumbnail?: Image;
    /** Creation time */
    created_at: string;
    /** Update time */
    updated_at: string;
};
export type SearchListResponse = {
    /** List of images found */
    images: Image[];
    /** List of collections found */
    collections: Collection[];
};
export type ImagesResponse = {
    /** Added timestamp */
    added_at: string;
    added_by?: User;
    image: Image;
};
export type ImagesListResponse = {
    /** Self link */
    href?: string;
    /** Previous page link */
    prev?: string;
    /** Next page link */
    next?: string;
    /** Items per page */
    limit: number;
    /** Current page */
    page: number;
    /** Total count */
    count?: number;
    /** List of items */
    items: ImagesResponse[];
};
export type ImageUploadRequest = {
    /** Image file data */
    data: Blob;
    /** Name of the file */
    file_name: string;
    /** Optional checksum of the file */
    checksum?: string;
};
export type ImageUploadResponse = {
    /** UID of the uploaded image */
    uid: string;
    /** Extracted metadata */
    metadata?: {
        [key: string]: any;
    };
};
export type DeleteAssetsResponse = {
    /** Results of deletion */
    results?: {
        /** UID of the asset */
        uid?: string;
        /** Whether it was deleted */
        deleted?: boolean;
        /** Error message if failed */
        error?: string;
    }[];
};
export type ImageUpdate = {
    /** Image name */
    name?: string;
    /** Owner UID */
    owner_uid?: string | null;
    /** Image description */
    description?: string | null;
    /** Is private */
    "private"?: boolean;
    /** Is favourited */
    favourited?: boolean;
    exif?: ImageExif;
    image_metadata?: {
        /** User-assigned label for the image. Null = unlabeled */
        label?: ("Red" | "Orange" | "Yellow" | "Purple" | "Pink" | "Green" | "Blue" | "None") | null;
        /** User-assigned rating (0-5). Null = unrated */
        rating?: number | null;
        /** Keywords */
        keywords?: string[];
    };
};
export type CollectionListResponse = {
    /** Self link */
    href?: string;
    /** Previous page link */
    prev?: string;
    /** Next page link */
    next?: string;
    /** Items per page */
    limit: number;
    /** Current page */
    page: number;
    /** Total count */
    count?: number;
    /** List of collections */
    items: Collection[];
};
export type CollectionCreate = {
    /** Collection name */
    name: string;
    /** Is private */
    "private"?: boolean | null;
    /** Collection description */
    description?: string;
};
export type CollectionDetailResponse = {
    /** Collection UID */
    uid: string;
    /** Collection name */
    name: string;
    /** Number of images */
    image_count?: number;
    /** Is private */
    "private"?: boolean | null;
    images: ImagesListResponse;
    created_by?: User;
    owner?: User;
    /** Collection description */
    description?: string;
    thumbnail?: Image;
    /** Creation time */
    created_at: string;
    /** Update time */
    updated_at: string;
};
export type CollectionUpdate = {
    /** Collection name */
    name?: string;
    /** Thumbnail image UID */
    thumbnailUID?: string;
    /** Collection description */
    description?: string;
    /** Is private */
    "private"?: boolean;
    /** Is favourited */
    favourited?: boolean;
    /** Owner UID */
    ownerUID?: string;
};
export type AddImagesResponse = {
    /** Whether images were added */
    added: boolean;
    /** Error message if failed */
    error?: string;
};
export type DeleteImagesResponse = {
    /** Whether images were deleted */
    deleted: boolean;
    /** Error message if failed */
    error?: string;
};
export type DownloadRequest = {
    /** List of UIDs to download */
    uids: string[];
    /** Desired filename for the archive */
    file_name?: string;
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
export type SettingDefault = {
    /** Unique name for the setting (primary key). */
    name: string;
    /** A readable and UI-friendly name for the setting (not required but highly recommended). */
    display_name: string;
    /** The default value everyone gets. */
    value: string;
    /** Data type of the setting. */
    value_type: "boolean" | "string" | "integer" | "enum" | "json";
    /** List of valid choices if type is enum. */
    allowed_values?: string[] | null;
    /** Describes whether a user can edit this setting. */
    is_user_editable: boolean;
    /** Category/group for the setting (e.g., General, Notifications). */
    group: string;
    /** Description for UI */
    description: string;
};
export type SettingOverride = {
    /** Links to the users table. */
    user_id: string;
    /** Links to SettingDefault.name. */
    name: string;
    /** The user's chosen value for the setting. */
    value: string;
};
export type LoggingConfig = {
    /** Logging level */
    level?: string;
};
export type UploadConfig = {
    /** Upload location */
    location?: string;
};
export type DatabaseConfig = {
    /** Database location/host */
    location?: string;
    /** Database user */
    user?: string;
    /** Masked password */
    password?: string;
    /** Database name */
    name?: string;
    /** Database port */
    port?: number;
};
export type QueueConfig = {
    /** Is queue enabled */
    enabled?: boolean;
    /** Queue host */
    host?: string;
    /** Queue port */
    port?: number;
    /** Queue username */
    username?: string;
    /** Masked password */
    password?: string;
    /** Redis DB index */
    db?: number;
    /** Use TLS */
    use_tls?: boolean;
    /** Connection pool size */
    pool_size?: number;
    /** Dial timeout */
    dial_timeout_seconds?: number;
    /** Read timeout */
    read_timeout_seconds?: number;
    /** Write timeout */
    write_timeout_seconds?: number;
};
export type LibvipsConfig = {
    /** Match system logging level */
    match_system_logging?: boolean;
    /** Cache max memory MB */
    cache_max_memory_mb?: number;
    /** Cache max files */
    cache_max_files?: number;
    /** Cache max operations */
    cache_max_operations?: number;
    /** Concurrency level */
    concurrency?: number;
    /** Vector enabled */
    vector_enabled?: boolean;
};
export type CacheConfig = {
    /** GC enabled */
    gc_enabled?: boolean;
};
export type UserManagementConfig = {
    /** Allow manual registration */
    allow_manual_registration?: boolean;
};
export type StorageMetricsConfig = {
    /** Metrics enabled */
    enabled?: boolean;
    /** Interval in seconds */
    interval_seconds?: number;
};
export type ImagineConfig = {
    /** Base URL of the application */
    baseUrl?: string;
    logging?: LoggingConfig;
    /** Base directory path */
    base_directory?: string;
    upload?: UploadConfig;
    database?: DatabaseConfig;
    redis?: QueueConfig;
    libvips?: LibvipsConfig;
    cache?: CacheConfig;
    user_management?: UserManagementConfig;
    storage_metrics?: StorageMetricsConfig;
};
export type SystemStatsResponse = {
    /** System uptime in seconds */
    uptime_seconds: number;
    /** Number of running goroutines */
    num_goroutine: number;
    /** Bytes of allocated heap objects */
    alloc_memory: number;
    /** Total bytes of memory obtained from the OS */
    sys_memory: number;
    /** Total size of files in the base directory */
    storage_used_bytes: number;
    /** Path to storage directory */
    storage_path: string;
    /** Total disk space on the system */
    total_system_space_bytes: number;
    /** Total available disk space on the system */
    total_available_space_bytes: number;
};
export type DatabaseStatsResponse = {
    /** Total number of users */
    user_count: number;
    /** Total number of images */
    image_count: number;
    /** Total number of collections */
    collection_count: number;
    /** Size of the database in bytes (Postgres only) */
    db_size_bytes?: number;
    /** Number of active connections (Postgres only) */
    active_connections?: number;
};
export type CacheStatusResponse = {
    /** Current size of the cache in bytes */
    size: number;
    /** Number of items currently in the cache */
    items: number;
    /** Number of cache hits */
    hits: number;
    /** Number of cache misses */
    misses: number;
    /** Cache hit ratio */
    hit_ratio: number;
};
export type AdminUserCreate = {
    /** User's full name */
    name: string;
    /** User's email address */
    email: string;
    /** User's password */
    password: string;
    /** User role */
    role?: "user" | "admin" | "superadmin" | "guest";
};
export type AdminUserUpdate = {
    /** First name */
    first_name?: string | null;
    /** Last name */
    last_name?: string | null;
    /** Username */
    username?: string | null;
    /** Email address */
    email?: string | null;
    /** User role */
    role?: ("user" | "admin" | "superadmin" | "guest") | null;
};
export type WorkerInfo = {
    /** Job topic/worker name (e.g., exif_process, image_process) */
    name: string;
    /** Human-readable worker name */
    display_name: string;
    /** Number of concurrent jobs for this worker */
    concurrency: number;
    /** Number of active worker jobs with this name/topic */
    count?: number;
};
export type WorkersListResponse = {
    /** List of workers */
    items: WorkerInfo[];
};
export type WorkerRegisterRequest = {
    /** Job topic/worker name (e.g., exif_process, image_process) */
    name: string;
    /** Number of concurrent jobs for this worker */
    concurrency?: number;
};
export type WorkerJobCreateRequest = {
    /** Job topic (e.g., exif_process, image_process) */
    "type": string;
    /** Command to execute (all=process all, missing=process missing) */
    command: "all" | "missing";
    /** Image UIDs to process (optional, if omitted all images are considered) */
    uids?: string[];
};
export type WorkerJobEnqueueResponse = {
    /** Response message */
    message: string;
    /** Count of enqueued jobs */
    count?: number;
};
export type WorkerJob = {
    /** Job UID */
    uid: string;
    /** Job type */
    "type": string;
    /** Job topic */
    topic: string;
    /** Job command */
    command?: string | null;
    /** Related image UID */
    image_uid?: string | null;
    /** Job status */
    status: string;
    /** Job payload */
    payload?: string | null;
    /** Error code if failed */
    error_code?: string | null;
    /** Error message if failed */
    error_msg?: string | null;
    /** Enqueued timestamp */
    enqueued_at: string;
    /** Started timestamp */
    started_at?: string | null;
    /** Completed timestamp */
    completed_at?: string | null;
};
export type WorkerJobsResponse = {
    /** List of jobs */
    items: WorkerJob[];
    /** Total count of jobs */
    total: number;
};
export type WorkerJobStatsResponse = {
    /** Total running jobs */
    running: number;
    /** Running jobs by topic */
    running_by_topic: {
        [key: string]: number;
    };
    /** Queued jobs by topic */
    queued_by_topic: {
        [key: string]: number;
    };
};
export type WsStatsResponse = {
    /** Number of connected clients */
    connectedClients: number;
    /** List of connected client IDs */
    clientIds: string[];
    /** Timestamp of the stats */
    timestamp: string;
};
export type WsMetricsResponse = {
    /** Number of connected clients */
    connectedClients: number;
    /** Total events processed */
    totalEvents: number;
    /** Count of events by type */
    eventsByType: {
        [key: string]: number;
    };
    /** Timestamp of the metrics */
    timestamp: string;
};
export type EventRecord = {
    /** Event timestamp */
    timestamp: string;
    /** Event name */
    event: string;
    /** Event data payload */
    data: {
        [key: string]: any;
    };
};
export type EventHistoryResponse = {
    /** List of historical events */
    events: EventRecord[];
    /** Count of events returned */
    count: number;
};
export type WsBroadcastRequest = {
    /** Event name to broadcast */
    event: string;
    /** Event data payload */
    data: {
        [key: string]: any;
    };
};
export type WsBroadcastResponse = {
    /** Whether broadcast was successful */
    success: boolean;
    /** Response message */
    message: string;
    /** Number of clients reached */
    clients: number;
};
export type SystemStatusResponse = {
    /** True if the system has been initialized (at least one superadmin exists and first_run_complete is true). */
    initialized: boolean;
    /** True if the current user (if authenticated) still needs to complete onboarding. */
    user_onboarding_required: boolean;
    /** True if the system requires initial superadmin setup (no superadmin exists). */
    needs_superadmin: boolean;
    /** True if user registration is enabled, let's users register accounts themselves */
    allow_manual_registration: boolean;
};
export type SuperadminSetupRequest = {
    /** Desired username */
    username: string;
    /** Email address */
    email: string;
    /** Password */
    password: string;
    /** First name */
    firstName?: string | null;
    /** Last name */
    lastName?: string | null;
};
export type SuperadminSetupResponse = {
    /** Response message */
    message: string;
    user: User;
    /** Session token for the newly created superadmin */
    sessionToken: string;
};
/**
 * Health ping
 */
export function ping(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: {
            /** Pong message */
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
    } | {
        status: 403;
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
    /** User email address */
    email: string;
    /** User password */
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
 * Update current authenticated user's profile
 */
export function updateCurrentUser(userUpdate: UserUpdate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: User;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    }>("/accounts/me", oazapfts.json({
        ...opts,
        method: "PATCH",
        body: userUpdate
    }));
}
/**
 * Onboard current authenticated user
 */
export function doUserOnboarding(userOnboardingBody: UserOnboardingBody, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: User;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    }>("/accounts/me/onboard", oazapfts.json({
        ...opts,
        method: "PUT",
        body: userOnboardingBody
    }));
}
/**
 * Update user password
 */
export function updatePassword(userPasswordUpdate: UserPasswordUpdate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: UserUpdate;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/accounts/me/password", oazapfts.json({
        ...opts,
        method: "PUT",
        body: userPasswordUpdate
    }));
}
/**
 * Get all sessions for the current user
 */
export function getSessions(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Session[];
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/sessions", {
        ...opts
    });
}
/**
 * Delete all sessions for the current user
 */
export function deleteSessions(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/sessions", {
        ...opts,
        method: "DELETE"
    });
}
/**
 * Get a specific session by ID for the current user
 */
export function getSessionById(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Session;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    }>(`/sessions/${encodeURIComponent(uid)}`, {
        ...opts
    });
}
/**
 * Update a specific session
 */
export function updateSession(uid: string, sessionUpdate: SessionUpdate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: Session;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    }>(`/sessions/${encodeURIComponent(uid)}`, oazapfts.json({
        ...opts,
        method: "PUT",
        body: sessionUpdate
    }));
}
/**
 * Delete a specific session
 */
export function deleteSession(uid: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: MessageResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    }>(`/sessions/${encodeURIComponent(uid)}`, {
        ...opts,
        method: "DELETE"
    });
}
/**
 * Get all user settings
 */
export function getUserSettings(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: UserSetting[];
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/accounts/me/settings", {
        ...opts
    });
}
/**
 * Update a user setting (Override)
 */
export function updateUserSetting(name: string, body: {
    /** New setting value */
    value: string;
}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: UserSetting;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/accounts/me/settings${QS.query(QS.explode({
        name
    }))}`, oazapfts.json({
        ...opts,
        method: "PATCH",
        body
    }));
}
/**
 * Update multiple user settings (batch override)
 */
export function updateUserSettingsBatch(userSettingUpdateRequest: UserSettingUpdateRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: UserSetting[];
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/accounts/me/settings", oazapfts.json({
        ...opts,
        method: "PUT",
        body: userSettingUpdateRequest
    }));
}
/**
 * Search for images and collections
 */
export function executeSearch(q: string, { limit, page }: {
    limit?: number;
    page?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: SearchListResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/search${QS.query(QS.explode({
        q,
        limit,
        page
    }))}`, {
        ...opts
    });
}
/**
 * List all images with pagination
 */
export function listImages({ limit, page, sortBy, order }: {
    limit?: number;
    page?: number;
    sortBy?: "taken_at" | "created_at" | "updated_at" | "name";
    order?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ImagesListResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/images${QS.query(QS.explode({
        limit,
        page,
        sort_by: sortBy,
        order
    }))}`, {
        ...opts
    });
}
/**
 * Upload an image (multipart)
 */
export function uploadImage(imageUploadRequest: ImageUploadRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ImageUploadResponse;
    } | {
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
        body: imageUploadRequest
    }));
}
/**
 * Delete multiple asset UID directories (soft move to trash or force delete)
 */
export function deleteImagesBulk(body: {
    /** List of image UIDs */
    uids: string[];
    /** Force deletion */
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
        status: 200;
        data: ImageUploadResponse;
    } | {
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
export function getImageFile(uid: string, { format, width, height, quality, download, token, password }: {
    format?: "webp" | "png" | "jpg" | "jpeg" | "avif" | "heif";
    width?: number;
    height?: number;
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
        width,
        height,
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
 * Get image metadata
 */
export function getImage(uid: string, opts?: Oazapfts.RequestOpts) {
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
    }>(`/images/${encodeURIComponent(uid)}`, {
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
        data: ImagesListResponse;
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
    /** List of image UIDs */
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
    /** List of image UIDs */
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
 * List all setting definitions
 */
export function listSettingDefinitions(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: SettingDefault[];
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/admin/settings/definitions", {
        ...opts
    });
}
/**
 * List all overrides (debug/admin)
 */
export function listSettingOverrides(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: SettingOverride[];
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/admin/settings/overrides", {
        ...opts
    });
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
 * Get system configuration
 */
export function getSystemConfig(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: ImagineConfig;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    }>("/system/config", {
        ...opts
    });
}
/**
 * Get system statistics (uptime, memory)
 */
export function getSystemStats(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: SystemStatsResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    }>("/admin/system/stats", {
        ...opts
    });
}
/**
 * Get database statistics (counts)
 */
export function getDatabaseStats(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: DatabaseStatsResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    }>("/admin/db/stats", {
        ...opts
    });
}
/**
 * Get cache status
 */
export function getCacheStatus(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: CacheStatusResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/admin/cache/status", {
        ...opts
    });
}
/**
 * Clear the image cache
 */
export function clearImageCache(opts?: Oazapfts.RequestOpts) {
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
    }>("/admin/cache", {
        ...opts,
        method: "DELETE"
    });
}
/**
 * List all users
 */
export function listUsers(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: User[];
    } | {
        status: 401;
        data: ErrorResponse;
    } | {
        status: 403;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/admin/users", {
        ...opts
    });
}
/**
 * Create a new user (admin)
 */
export function adminCreateUser(adminUserCreate: AdminUserCreate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: User;
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
        status: 409;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/admin/users", oazapfts.json({
        ...opts,
        method: "POST",
        body: adminUserCreate
    }));
}
/**
 * Update user details (admin)
 */
export function adminUpdateUser(uid: string, adminUserUpdate: AdminUserUpdate, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: User;
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
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/admin/users/${encodeURIComponent(uid)}`, oazapfts.json({
        ...opts,
        method: "PATCH",
        body: adminUserUpdate
    }));
}
/**
 * Delete user (admin)
 */
export function adminDeleteUser(uid: string, body?: {
    /** If true, permanently deletes the user and all associated data (sessions, settings). */
    force?: boolean;
}, opts?: Oazapfts.RequestOpts) {
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
        status: 404;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/admin/users/${encodeURIComponent(uid)}`, oazapfts.json({
        ...opts,
        method: "DELETE",
        body
    }));
}
/**
 * List all workers
 */
export function listAvailableWorkers(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WorkersListResponse;
    } | {
        status: 401;
        data: ErrorResponse;
    }>("/jobs/workers", {
        ...opts
    });
}
/**
 * Register a new worker
 */
export function registerWorker(workerRegisterRequest: WorkerRegisterRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: WorkerInfo;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/jobs/workers", oazapfts.json({
        ...opts,
        method: "POST",
        body: workerRegisterRequest
    }));
}
/**
 * Create/enqueue a job
 */
export function createJob(workerJobCreateRequest: WorkerJobCreateRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 202;
        data: WorkerJobEnqueueResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/jobs", oazapfts.json({
        ...opts,
        method: "POST",
        body: workerJobCreateRequest
    }));
}
/**
 * List jobs with filtering and pagination
 */
export function listJobs({ status, topic, limit, page }: {
    status?: "queued" | "running" | "completed" | "failed" | "cancelled";
    topic?: string;
    limit?: number;
    page?: number;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WorkerJobsResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>(`/jobs${QS.query(QS.explode({
        status,
        topic,
        limit,
        page
    }))}`, {
        ...opts
    });
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
 * Get job stats by topic
 */
export function getJobStats(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: WorkerJobStatsResponse;
    }>("/jobs/stats", {
        ...opts
    });
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
            /** List of events */
            events: EventRecord[];
            /** Number of events */
            count: number;
            /** Next cursor ID */
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
        status: 500;
        data: ErrorResponse;
    }>(`/events/send/${encodeURIComponent(clientId)}`, oazapfts.json({
        ...opts,
        method: "POST",
        body: wsBroadcastRequest
    }));
}
/**
 * Get system and user onboarding status
 */
export function getSystemStatus(opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 200;
        data: SystemStatusResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/system/status", {
        ...opts
    });
}
/**
 * Initialize application with first superadmin user
 */
export function setupSuperadmin(superadminSetupRequest: SuperadminSetupRequest, opts?: Oazapfts.RequestOpts) {
    return oazapfts.fetchJson<{
        status: 201;
        data: SuperadminSetupResponse;
    } | {
        status: 400;
        data: ErrorResponse;
    } | {
        status: 409;
        data: ErrorResponse;
    } | {
        status: 500;
        data: ErrorResponse;
    }>("/setup/superadmin", oazapfts.json({
        ...opts,
        method: "POST",
        body: superadminSetupRequest
    }));
}
