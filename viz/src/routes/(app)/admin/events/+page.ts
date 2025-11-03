import { getSseStats, getSseMetrics, getEventHistory } from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async () => {
    try {
        const [statsRes, metricsRes, historyRes] = await Promise.all([
            getSseStats(),
            getSseMetrics(),
            getEventHistory()
        ]);

        return {
            stats: statsRes.status === 200 ? statsRes.data : {
                connectedClients: 0,
                clientIds: [],
                timestamp: new Date().toISOString()
            },
            metrics: metricsRes.status === 200 ? metricsRes.data : {
                connectedClients: 0,
                totalEvents: 0,
                eventsByType: {},
                timestamp: new Date().toISOString()
            },
            history: historyRes.status === 200 ? (historyRes.data.events || []) : []
        };
    } catch (e) {
        console.error("Failed to load events data:", e);
        return {
            stats: {
                connectedClients: 0,
                clientIds: [],
                timestamp: new Date().toISOString()
            },
            metrics: {
                connectedClients: 0,
                totalEvents: 0,
                eventsByType: {},
                timestamp: new Date().toISOString()
            },
            history: [],
            error: (e as Error).message
        };
    }
};
