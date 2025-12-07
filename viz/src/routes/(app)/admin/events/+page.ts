import { error } from "@sveltejs/kit";
import { getWsStats, getWsMetrics, getEventsSince } from "$lib/api";
import type { PageLoad } from "./$types";
import type { WsStatsResponse, WsMetricsResponse, EventRecord } from "$lib/api";

interface PageLoadData {
    stats: WsStatsResponse;
    metrics: WsMetricsResponse;
    history: EventRecord[];
}

export const load: PageLoad = async (): Promise<PageLoadData> => {
    const [statsRes, metricsRes, historyRes] = await Promise.all([
        getWsStats(),
        getWsMetrics(),
        getEventsSince({ limit: 50 })
    ]);

    // Throw error if any response is not 2xx
    if (statsRes.status !== 200) {
        throw error(statsRes.status, `Failed to load stats: ${statsRes.status}`);
    }

    if (metricsRes.status !== 200) {
        throw error(metricsRes.status, `Failed to load metrics: ${metricsRes.status}`);
    }

    if (historyRes.status !== 200) {
        throw error(historyRes.status, `Failed to load history: ${historyRes.status}`);
    }

    return {
        stats: statsRes.data,
        metrics: metricsRes.data,
        history: 'events' in historyRes.data ? (historyRes.data.events || []) : []
    };
};
