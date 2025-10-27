/**
 * Configured oazapfts API client
 * Auto-generated functions with proper base URL and SvelteKit fetch integration
 */
import { MEDIA_SERVER } from "$lib/constants";
import { createServerURL } from "$lib/utils/url";
import { defaults } from "./client.gen";

// Configure the base URL for the API client
defaults.baseUrl = createServerURL(MEDIA_SERVER);

// Use global fetch (SvelteKit automatically enhances this in server contexts)
defaults.fetch = fetch;

// Re-export everything from the generated client
export * from "./client.gen";
