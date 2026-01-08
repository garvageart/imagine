import { test, expect } from '@playwright/test';

test.describe('Layout Regression Tests', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/');
        await expect(page.locator('.viz-workspace')).toBeVisible({ timeout: 15000 });
    });

    test('should correctly cleanup when moving last tab from a nested split', async ({ page }) => {
        // 1. Setup layout: Split initial group to get 2 groups
        const firstTab = page.locator('button[role="tab"]').first();
        await expect(firstTab).toBeVisible();
        await firstTab.click({ button: 'right' });
        await page.locator('text="Split Right"').click();
        
        // Wait for split
        await expect(page.locator('.splitpanes__splitter').first()).toBeVisible();

        // 2. Split the NEW group (the one on the right) again to get 3 groups
        // The new group should be the second .tab-group-panel
        const groups = page.locator('.tab-group-panel');
        await expect(groups).toHaveCount(2);

        const secondGroupTabs = groups.nth(1).locator('button[role="tab"]').first();
        await secondGroupTabs.click({ button: 'right' });
        await page.locator('text="Split Right"').click();

        // Wait for second split
        await expect(groups).toHaveCount(3);
        
        // Now we have [Group 1] | [Group 2] | [Group 3]
        // This likely creates a structure like: SplitNode(Group1, SplitNode(Group2, Group3))
        // or SplitNode(Group1, Group2, Group3) depending on how split works.
        // Based on "splitGroup" implementation:
        // if parent has same orientation, it adds as sibling.
        // Default split is horizontal (side-by-side).

        // 3. Move the tab from Group 3 to Group 2
        const group2 = groups.nth(1);
        const group3 = groups.nth(2);
        
        const tabInGroup3 = group3.locator('button[role="tab"]').first();
        const dropTargetInGroup2 = group2.locator('.tab-group-header');

        await expect(tabInGroup3).toBeVisible();
        await expect(dropTargetInGroup2).toBeVisible();

        // Drag and drop
        await tabInGroup3.dragTo(dropTargetInGroup2);

        // 4. Assertions
        
        // Group 3 should be gone (it was empty)
        await expect(page.locator('.tab-group-panel')).toHaveCount(2);

        // Group 1 should STILL exist (Regression check: it shouldn't be removed)
        // Check content or existence of first group
        const group1 = page.locator('.tab-group-panel').first();
        await expect(group1).toBeVisible();
        
        // Ensure the remaining structure is valid (Group 1 and Group 2 side by side)
        // Group 2 should now have multiple tabs (or at least the moved one)
        const tabsInGroup2 = group2.locator('button[role="tab"]');
        // Original tab + moved tab
        await expect(tabsInGroup2).toHaveCount(2);
    });
});
