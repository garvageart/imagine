import { describe, it, expect, vi } from 'vitest';

// mock listImages from generated client
const mockImagesPage = { items: [{ image: { uid: 'i1', name: 'Img1' } }], count: 1, next: null, prev: null };
vi.mock('$lib/api', () => ({
    listImages: async ({ limit, page }: any) => ({ status: 200, data: mockImagesPage })
}));

import { load } from '../../routes/(app)/photos/+page';

describe('photos page load', () => {
    it('returns images list from API', async () => {
        const res = await load({ url: new URL('http://localhost/?limit=10&page=0') } as any);
        expect(res.images).toBeDefined();
        expect(res.images.length).toBe(1);
        expect(res.count).toBe(1);
    });
});
