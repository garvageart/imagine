// WebSocket Events API helpers
import { API_BASE_URL } from ".";

/**
 * WebSocket message structure from server
 */
export interface WSMessage {
    event: string;
    data: any;
    id?: number;
    timestamp?: string;
}

/**
 * WebSocket connection options
 */
export interface WSConnectionOptions {
    /** Callback for handling events */
    onEvent: (event: string, data: any) => void;
    /** Callback for handling errors */
    onError?: (error: Event) => void;
    /** Callback when connection opens */
    onOpen?: () => void;
    /** Callback when connection closes */
    onClose?: (code: number, reason: string) => void;
    /** Auto-reconnect on close (default: true) */
    autoReconnect?: boolean;
    /** Reconnect delay in ms (default: 1000) */
    reconnectDelay?: number;
    /** Maximum reconnect attempts (default: 5, 0 for infinite) */
    maxReconnectAttempts?: number;
}

/**
 * WebSocket client wrapper with auto-reconnect
 */
export class WSClient {
    private ws: WebSocket | null = null;
    private url: string;
    private options: Required<WSConnectionOptions>;
    private reconnectAttempts = 0;
    private reconnectTimeout: number | null = null;
    private isClosed = false;

    constructor(options: WSConnectionOptions) {
        let base = API_BASE_URL;

        // If the base URL is relative (e.g., "/api"), resolve it to an absolute URL
        // using the current window location. WebSockets require absolute URLs.
        if (base.startsWith('/') && typeof window !== 'undefined') {
            base = window.location.origin + base;
        }

        this.url = base.replace(/^http/, 'ws') + '/events';

        this.options = {
            onEvent: options.onEvent,
            onError: options.onError || (() => { }),
            onOpen: options.onOpen || (() => { }),
            onClose: options.onClose || (() => { }),
            autoReconnect: options.autoReconnect ?? true,
            reconnectDelay: options.reconnectDelay ?? 1000,
            maxReconnectAttempts: options.maxReconnectAttempts ?? 5
        };

        this.connect();
    }

    private handleText = (text: string) => {
        // Server may batch messages into a single frame separated by newlines.
        // Split and parse each line individually.
        const parts = text.split('\n');
        for (const part of parts) {
            const line = part.trim();
            if (!line) continue;
            try {
                const message: WSMessage = JSON.parse(line);

                if (message.event === 'ping') {
                    this.send('pong', {});
                    continue;
                }

                this.options.onEvent(message.event, message.data);
            } catch (error) {
                console.error('[WebSocket] Failed to parse message part:', error, 'rawPart:', line);
            }
        }
    };

    private connect() {
        if (this.isClosed) return;

        try {
            this.ws = new WebSocket(this.url);

            this.ws.onopen = () => {
                console.debug('[WebSocket] Connected');
                this.reconnectAttempts = 0;
                this.options.onOpen();
            };

            this.ws.onmessage = (event) => {
                // event.data can be a string or a Blob; handle both
                try {
                    if (typeof event.data === 'string') {
                        this.handleText(event.data);
                    } else if (event.data instanceof Blob) {
                        // convert blob to text asynchronously
                        (event.data as Blob).text().then((txt) => this.handleText(txt)).catch((err) => {
                            console.error('[WebSocket] Failed to read Blob message:', err);
                        });
                    } else {
                        // Fallback: try to coerce to string
                        try {
                            this.handleText(String(event.data));
                        } catch (err) {
                            console.error('[WebSocket] Unknown message type, cannot parse:', err, event.data);
                        }
                    }
                } catch (error) {
                    console.error('[WebSocket] Failed to handle incoming message:', error, event.data);
                }
            };

            this.ws.onerror = (error) => {
                console.error('[WebSocket] Error:', error);
                this.options.onError(error);
            };

            this.ws.onclose = (event: CloseEvent) => {
                console.debug('[WebSocket] Closed', event.code, event.reason);
                this.options.onClose(event.code, event.reason);

                if (this.options.autoReconnect && !this.isClosed) {
                    this.scheduleReconnect();
                }
            };
        } catch (error) {
            console.error('[WebSocket] Connection error:', error);
            if (this.options.autoReconnect && !this.isClosed) {
                this.scheduleReconnect();
            }
        }
    }

    private scheduleReconnect() {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
        }

        // Check if we've exceeded max attempts (0 means infinite)
        if (this.options.maxReconnectAttempts > 0 &&
            this.reconnectAttempts >= this.options.maxReconnectAttempts) {
            console.error('[WebSocket] Max reconnect attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.options.reconnectDelay * Math.min(this.reconnectAttempts, 5); // Exponential backoff capped at 5x

        console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

        this.reconnectTimeout = window.setTimeout(() => {
            this.connect();
        }, delay);
    }

    /**
     * Send a message to the server
     */
    send(event: string, data: any) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ event, data }));
        } else {
            console.warn('[WebSocket] Cannot send - not connected');
        }
    }

    /**
     * Close the WebSocket connection
     */
    close() {
        this.isClosed = true;

        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    /**
     * Get current connection state
     */
    get readyState(): number {
        return this.ws?.readyState ?? WebSocket.CLOSED;
    }

    /**
     * Check if connected
     */
    get isConnected(): boolean {
        return this.ws?.readyState === WebSocket.OPEN;
    }
}

/**
 * Create a WebSocket connection (convenience function)
 * @returns WSClient instance that can be closed later
 */
export function createWSConnection(
    onEvent: (event: string, data: any) => void,
    onError?: (error: Event) => void,
    onOpen?: () => void,
    onClose?: (code: number, reason: string) => void
): WSClient {
    return new WSClient({
        onEvent,
        onError,
        onOpen,
        onClose: onClose ? (code, reason) => onClose(code, reason) : undefined
    });
}
