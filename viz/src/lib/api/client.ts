/**
 * Configured oazapfts API client
 * Auto-generated functions with proper base URL and SvelteKit fetch integration
 */
import { MEDIA_SERVER } from "$lib/constants";
import { createServerURL } from "$lib/utils/url";
import { defaults } from "./client.gen";
import type { DownloadRequest, ErrorResponse } from "./client.gen";
import * as QS from "@oazapfts/runtime/query";

// Configure the base URL for the API client
defaults.baseUrl = createServerURL(MEDIA_SERVER);

// Include credentials (cookies) with all requests for authentication
defaults.credentials = "include";

/**
 * Download images as a ZIP blob using a token
 * This is a custom implementation because oazapfts doesn't properly handle binary responses
 */
export async function downloadImagesBlob(
    token: string,
    downloadRequest: DownloadRequest,
    password?: string
): Promise<
    | { status: 200; data: Blob; }
    | { status: 400; data: ErrorResponse; }
    | { status: 401; data: ErrorResponse; }
    | { status: 403; data: ErrorResponse; }
    | { status: 500; data: ErrorResponse; }
> {
    const baseUrl = defaults.baseUrl || "";
    const queryParams = QS.query(QS.explode({ token, password }));
    const url = `${baseUrl}/download${queryParams}`;

    try {
        const response = await fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                ...defaults.headers
            },
            credentials: "include",
            body: JSON.stringify(downloadRequest)
        });

        if (response.ok) {
            const blob = await response.blob();
            return { status: 200, data: blob };
        }

        // Try to parse error response as JSON
        const contentType = response.headers.get("content-type") || "";
        if (contentType.includes("application/json")) {
            const errorData = await response.json();
            return {
                status: response.status as 400 | 401 | 403 | 500,
                data: errorData
            };
        }

        // Fallback error
        return {
            status: response.status as 400 | 401 | 403 | 500,
            data: { error: `Request failed with status ${response.status}` }
        };
    } catch (error) {
        return {
            status: 500,
            data: { error: error instanceof Error ? error.message : "Network error" }
        };
    }
}

// Re-export everything from the generated client
export * from "./client.gen";
export { initApi } from "./init";
