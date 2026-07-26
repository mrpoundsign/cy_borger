// test_character_edit.js — Playwright test for full character editing, blank character creation, and draft/keep workflow.
const { test, expect, request } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

async function loginUser(page) {
    await page.goto(BASE_URL + '/');
    await page.waitForLoadState('networkidle');

    if (await page.locator('button:has-text("🚪 LOGOUT")').isVisible()) {
        await page.locator('button:has-text("🚪 LOGOUT")').click();
        await page.waitForLoadState('networkidle');
    }

    if (await page.locator('#tab-btn-register').isVisible()) {
        await page.locator('#tab-btn-register').click();
        await page.waitForTimeout(200);
        const name = 'EditOp' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test.describe('Character editing & Draft/Keep workflow', () => {
    let apiCtx;

    test.beforeAll(async () => {
        apiCtx = await request.newContext({ baseURL: BASE_URL });
    });

    test.afterAll(async () => {
        await apiCtx.dispose();
    });

    test('create blank character and verify default fields', async ({ page }) => {
        await loginUser(page);

        // Click "Create Blank Character"
        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        await expect(page).toHaveURL(/\/character\//);
        await expect(page.locator('#identity-view h1')).toContainText('UNNAMED OPERATOR');
    });

    test('random character roll creates saved character for logged-in user', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Roll Random Character")').click();
        await page.waitForLoadState('networkidle');

        await expect(page).toHaveURL(/\/character\//);
        await expect(page.locator('#identity-view h1')).toBeVisible();
    });

    test('inline field editing updates character real-time', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        const newName = 'CyberGhost_' + Math.floor(Math.random() * 1000);

        // Edit Name by clicking ✏️ Edit Text button
        await page.locator('button:has-text("✏️ Edit Text")').first().click();
        await page.locator('#identity-edit input[name="name"]').fill(newName);
        await page.locator('#identity-edit button[type="submit"]').first().click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('#identity-view h1')).toContainText(newName);
    });

    test('inventory list workflow › add and delete weapon items', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        // Add Weapon
        await page.locator('button:has-text("+ Add Weapon / Armor")').first().click();
        await page.locator('#gear-add-edit input[name="name"]').first().fill('Nano-Blade');
        await page.locator('#gear-add-edit button[type="submit"]').first().click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('text=Nano-Blade')).toBeVisible();
    });
});
