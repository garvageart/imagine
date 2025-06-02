import { expect, test, Page } from '@playwright/test';

test('home page has expected h1', async ({ page }: { page: Page }) => {
	await page.goto('/');
	await expect(page.locator('h1').toBeVisible());
});