// test_character_delete.js — Playwright test for owner-only character deletion with name confirmation.
const { test, expect, request } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

async function loginUser(page) {
    await page.goto(BASE_URL + '/');
    await page.waitForLoadState('networkidle');

    if (await page.locator('#tab-btn-register').isVisible()) {
        await page.locator('#tab-btn-register').click();
        await page.waitForTimeout(200);
        const name = 'DelOp' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test.describe('Character Deletion Workflow', () => {
    let apiCtx;

    test.beforeAll(async () => {
        apiCtx = await request.newContext({ baseURL: BASE_URL });
    });

    test.afterAll(async () => {
        await apiCtx.dispose();
    });

    test('owner can delete character by typing exact character name', async ({ page }) => {
        await loginUser(page);

        // Create a blank character
        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('#identity-view h1')).toContainText('UNNAMED OPERATOR');

        // Click DELETE button in toolbar
        const deleteBtn = page.locator('button:has-text("💣 DELETE")');
        await expect(deleteBtn).toBeVisible();
        await deleteBtn.click();

        // Deletion modal should open
        const modal = page.locator('#delete-char-edit');
        await expect(modal).toBeVisible();

        // Type exact character name
        await modal.locator('input[name="confirm_name"]').fill('UNNAMED OPERATOR');
        await modal.locator('button[type="submit"]:has-text("PERMANENTLY DELETE")').click();
        await page.waitForLoadState('networkidle');

        // Confirm redirected to home page
        await expect(page).toHaveURL(BASE_URL + '/');
    });
});
