import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { downloadImagesZipBlob } from '$lib/api';

describe('API client integration (downloadImagesBlob)', () => {
    const origFetch = global.fetch;

    afterEach(() => {
        global.fetch = origFetch as any;
        vi.restoreAllMocks();
    });

    it('returns blob on success', async () => {
        const fakeBlob = { ok: true, blob: async () => ({ ok: true, content: 'zip' }) };
        global.fetch = vi.fn().mockResolvedValue(fakeBlob as any) as any;

        const res = await downloadImagesZipBlob('token123', { images: [] } as any);
        expect(res.status).toBe(200);
        // data should be the object returned from our fake blob()
        expect((res as any).data).toBeDefined();
        expect((res as any).data.content).toBe('zip');
    });

    it('parses JSON error response', async () => {
        const response = {
            ok: false,
            status: 400,
            headers: { get: (k: string) => (k === 'content-type' ? 'application/json' : '') },
            json: async () => ({ error: 'bad' })
        };

        global.fetch = vi.fn().mockResolvedValue(response as any) as any;

        const res = await downloadImagesZipBlob('t', { images: [] } as any);
        expect(res.status).toBe(400);
        expect((res as any).data).toBeDefined();
        expect((res as any).data.error).toBe('bad');
    });

    it('handles network errors', async () => {
        global.fetch = vi.fn().mockRejectedValue(new Error('network fail')) as any;

        const res = await downloadImagesZipBlob('t', { images: [] } as any);
        expect(res.status).toBe(500);
        expect((res as any).data.error).toMatch(/network/i);
    });
});
