import type { SystemStatusResponse } from "$lib/api/client.gen";

export interface ViewfinderConfig {
    environment: "dev" | "prod";
    version: string;
    theme?: string;
}

export interface VizConfig {
    system: SystemStatusResponse;
};