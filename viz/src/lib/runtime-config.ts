export interface RuntimeConfig {
    mediaServer?: string;
    uiServer?: string;
    version?: string;
}

export async function loadRuntimeConfig(): Promise<RuntimeConfig> {
    try {
        const maybe = (globalThis as any).__RUNTIME_CONFIG__;
        if (maybe) {
            const parsed = typeof maybe === 'string' ? JSON.parse(maybe) : maybe;
            const servers = parsed.servers ?? {};

            let mediaServer: string | undefined;
            if (servers['api']) {
                const s = servers['api'];
                if (s.host) {
                    mediaServer = s.host;
                } else if (s.port) {
                    mediaServer = (typeof window !== 'undefined') ? `http://${window.location.hostname}:${s.port}` : `http://localhost:${s.port}`;
                }
            }

            let uiServer: string | undefined;
            if (servers.viz) {
                const s = servers.viz;
                if (s.host) {
                    uiServer = s.host;
                } else if (s.port) {
                    uiServer = (typeof window !== 'undefined') ? `http://${window.location.hostname}:${s.port}` : `http://localhost:${s.port}`;
                }
            }

            const out: RuntimeConfig = { version: parsed.version };
            if (mediaServer) out.mediaServer = mediaServer;
            if (uiServer) out.uiServer = uiServer;

            return out;
        }
    } catch (e) {
        // fall through to fetch fallback
    }

    return {};
}

export default loadRuntimeConfig;
