// SSE Events API helpers
import { MEDIA_SERVER } from "$lib/constants";
import { createServerURL } from "$lib/utils/url";
import * as api from "./client.gen";

// Re-export types from generated client
export type {
    SseStatsResponse,
    SseMetricsResponse,
    EventRecord,
    EventHistoryResponse,
    SseBroadcastRequest,
    SseBroadcastResponse
} from "./client.gen";

/**
 * Get SSE connection statistics
 */
export async function getSSEStats() {
    return api.getSseStats();
}

/**
 * Get SSE metrics including event counts by type
 */
export async function getSSEMetrics() {
    return api.getSseMetrics();
}

/**
 * Get event history
 * @param limit - Maximum number of events to retrieve (default: 50, max: 100)
 */
export async function getEventHistory(limit: number = 50) {
    return api.getEventHistory({ limit },);
}

/**
 * Clear event history
 */
export async function clearEventHistory() {
    return api.clearEventHistory();
}

/**
 * Broadcast a message to all connected SSE clients
 */
export async function broadcastSSEEvent(event: string, data: any) {
    return api.broadcastSseEvent({ event, data },);
}

/**
 * Send a message to a specific SSE client
 */
export async function sendToSSEClient(clientId: string, event: string, data: any) {
    return api.sendToSseClient(clientId, { event, data },);
}

/**
 * Create an SSE connection
 * @param onEvent - Callback for handling events
 * @param onError - Callback for handling errors
 * @returns EventSource instance that can be closed later
 */
export function createSSEConnection(
    onEvent: (event: string, data: any) => void,
    onError?: (error: Event) => void
): EventSource {
    const url = createServerURL(MEDIA_SERVER) + "/events";
    const eventSource = new EventSource(url);

    const eventTypes = [
        // Generic lifecycle
        "connected",
        "job-started",
        "job-progress", // new generic progress event supported by server
        "job-completed",
        "job-failed",
        "test"
    ];

    eventTypes.forEach((eventType) => {
        eventSource.addEventListener(eventType, (e: MessageEvent) => {
            try {
                const data = JSON.parse(e.data);
                onEvent(eventType, data);
            } catch (error) {
                console.error(`Failed to parse ${eventType} event:`, error);
            }
        });
    });

    if (onError) {
        eventSource.onerror = onError;
    }

    return eventSource;
}
