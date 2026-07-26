// test_character_delete.js — Playwright test for owner-only character deletion with name confirmation.
//
// Run: npx playwright test test_character_delete.js
//
const { test, expect, request } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('Character Deletion Workflow', () => {
    let apiCtx;

    test.beforeAll(async () => {
        apiCtx = await request.newContext({ baseURL: BASE_URL });
    });

    test.afterAll(async () => {
        await apiCtx.dispose();
    });

    test('owner can delete character by typing exact character name', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        // Create a blank character
        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        const charUrl = page.url();
        await expect(page.locator('#identity-view h1')).toContainText('UNNAMED OPERATOR');

        // Click DELETE button in toolbar
        const deleteBtn = page.locator('button:has-text("💣 DELETE")');
        await expect(deleteBtn).toBeVisible();
        await deleteBtn.click();

        // Deletion modal should open
        const modal = page.locator('#delete-char-edit');
        await expect(modal).toBeVisible();

        // Type exact character name
        const nameInput = modal.locator('input[name="confirm_name"]');
        await nameInput.fill('UNNAMED OPERATOR');
        await modal.locator('button:has-text("PERMANENTLY DELETE")').click();

        await page.waitForLoadState('networkidle');

        // Should redirect back to home page
        await expect(page).toHaveURL(BASE_URL + '/');

        // Navigating back to deleted character URL should return 404 / Not Found
        const res = await page.goto(charUrl);
        expect(res.status()).toBe(404);
    });
});
