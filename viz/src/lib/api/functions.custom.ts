/**
* Custom API functions that require special handling beyond what openapi-fetch provides.
* These are manually written for specific use cases like upload progress tracking.
*/
import type { ImageUploadFileData } from "$lib/upload/manager.svelte";
import type { DownloadRequest, ErrorResponse } from "./client.gen";
import { API_BASE_URL, defaults, type ImageUploadResponse } from ".";
import * as Oazapfts from "@oazapfts/runtime";
import * as QS from "@oazapfts/runtime/query";

