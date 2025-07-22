export interface VizConfig {
    environment: "dev" | "prod";
    version: string;
    debug: boolean;
    theme?: string;
}