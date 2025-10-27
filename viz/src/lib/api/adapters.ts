/**
 * API Type Adapters
 * 
 * This file exports type aliases from the generated OpenAPI types (api.gen.ts)
 * to provide a cleaner, more maintainable way to reference API DTOs throughout the application.
 * 
 * ⚠️ IMPORTANT: Use these types instead of hardcoded types in images.ts, http.ts, etc.
 * These types are auto-generated from the OpenAPI spec and guarantee sync with the backend.
 */

import type { components, operations } from './types/api.gen';

// ===== Core Schema Types =====

export type APIPagination = components["schemas"]["Pagination"];
export type APIImage = components["schemas"]["Image"];
export type APIImageEXIF = components["schemas"]["ImageEXIF"];
export type APIImagePaths = components["schemas"]["ImagePaths"];
export type APIImageMetadata = components["schemas"]["ImageMetadata"];
export type APICollection = components["schemas"]["Collection"];
export type APICollectionImage = components["schemas"]["CollectionImage"];
export type APICollectionCreate = components["schemas"]["CollectionCreate"];
export type APICollectionDetailResponse = components["schemas"]["CollectionDetailResponse"];
export type APICollectionListResponse = components["schemas"]["CollectionListResponse"];
export type APIImagesResponse = components["schemas"]["ImagesResponse"];
export type APIImagesPage = components["schemas"]["ImagesPage"];

// ===== Operation Response Types =====

export type ListCollectionsResponse = operations["listCollections"]["responses"]["200"]["content"]["application/json"];
export type GetCollectionResponse = operations["getCollection"]["responses"]["200"]["content"]["application/json"];
export type CreateCollectionResponse = operations["createCollection"]["responses"]["201"]["content"]["application/json"];
export type ListCollectionImagesResponse = operations["listCollectionImages"]["responses"]["200"]["content"]["application/json"];

// ===== Helper Types =====

/**
 * Extended image type that includes the added_at timestamp from ImagesResponse
 * Use this when displaying images in a collection context
 */
export type ImageWithTimestamp = {
    image: APIImage;
    added_at: string;
    added_by?: string;
};

/**
 * Helper to extract the Image type from ImagesResponse
 */
export type ExtractImage<T extends APIImagesResponse> = T["image"];
