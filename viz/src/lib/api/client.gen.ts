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
    scopes: string[];
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
export type UserUpdate = {
    first_name?: string | null;
    last_name?: string | null;
    username?: string | null;
    email?: string | null;
};
export type UserSetting = {
    name: string;
    display_name: string;
    /** The effective value (override if exists, else default). */
    value: string;
    default_value: string;
    value_type: string;
    allowed_values?: string[] | null;
    is_user_editable?: boolean;
    group: string;
    description: string;
};
export type UserOnboardingBody = {
    first_name: string;
    last_name: string;
    /** User-specific setting overrides */
    settings: UserSetting[];
};
export type UserPasswordUpdate = {
    current: string;
    "new": string;
};
export type SessionUpdate = {
    clientName?: string | null;
    status?: number | null;
};
export type UserSettingUpdateRequest = {
    settings: {
        name: string;
        value: string;
    }[];
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
    has_icc_profile?: boolean;
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
    owner?: User;
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
export type ImageUploadRequest = {
    data: Blob;
    file_name: string;
    checksum?: string;
};
export type ImageUploadResponse = {
    uid: string;
    metadata?: {
        [key: string]: any;
    };
};
export type DeleteAssetsResponse = {
    results?: {
        uid?: string;
        deleted?: boolean;
        error?: string;
    }[];
};
export type ImageUpdate = {
    name?: string;
    owner_uid?: string | null;
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
    owner?: User;
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
    owner?: User;
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
    level?: string;
};
export type UploadConfig = {
    location?: string;
};
export type DatabaseConfig = {
    location?: string;
    user?: string;
    /** Masked password */
    password?: string;
    name?: string;
    port?: number;
};
export type QueueConfig = {
    enabled?: boolean;
    host?: string;
    port?: number;
    username?: string;
    /** Masked password */
    password?: string;
    db?: number;
    use_tls?: boolean;
    pool_size?: number;
    dial_timeout_seconds?: number;
    read_timeout_seconds?: number;
    write_timeout_seconds?: number;
};
export type LibvipsConfig = {
    match_system_logging?: boolean;
    cache_max_memory_mb?: number;
    cache_max_files?: number;
    cache_max_operations?: number;
    concurrency?: number;
    vector_enabled?: boolean;
};
export type CacheConfig = {
    gc_enabled?: boolean;
};
export type UserManagementConfig = {
    allow_manual_registration?: boolean;
};
export type StorageMetricsConfig = {
    enabled?: boolean;
    interval_seconds?: number;
};
export type ImagineConfig = {
    baseUrl?: string;
    logging?: LoggingConfig;
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
    uptime_seconds: number;
    num_goroutine: number;
    /** Bytes of allocated heap objects */
    alloc_memory: number;
    /** Total bytes of memory obtained from the OS */
    sys_memory: number;
    /** Total size of files in the base directory */
    storage_used_bytes: number;
    storage_path: string;
    /** Path to the base directory being measured */
    total_system_space_bytes?: number;
};
export type DatabaseStatsResponse = {
    user_count: number;
    image_count: number;
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
    name: string;
    email: string;
    password: string;
    role?: "user" | "admin" | "superadmin" | "guest";
};
export type AdminUserUpdate = {
    first_name?: string | null;
    last_name?: string | null;
    username?: string | null;
    email?: string | null;
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
export type WorkerJobsResponse = {
    items: WorkerJob[];
    total: number;
};
export type WorkerJobStatsResponse = {
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
export type SystemStatusResponse = {
    /** True if the system has been initialized (at least one user exists or first_run_complete is true). */
    initialized: boolean;
    /** True if the authenticated user needs to complete onboarding. */
    user_onboarding_required: boolean;
    /** True if the system requires initial superadmin setup. */
    needs_superadmin: boolean;
    /** True if user registration is enabled, let's users register accounts themselves */
    allow_manual_registration: boolean;
};
export type SuperadminSetupRequest = {
    username: string;
    email: string;
    password: string;
    firstName?: string | null;
    lastName?: string | null;
};
export type SuperadminSetupResponse = {
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
        data: ImagesPage;
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
