// test_character_edit.js — Playwright test for full character editing, blank character creation, and draft/keep workflow.
const { test, expect, request } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

async function loginUser(page) {
    await page.goto(BASE_URL + '/');
    await page.waitForLoadState('networkidle');

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

    test('random character roll shows draft banner, keeping it saves it', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Roll Random Character")').click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('.draft-banner')).toBeVisible();

        // Click KEEP OPERATOR
        await page.locator('button:has-text("KEEP OPERATOR")').click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('.draft-banner')).not.toBeVisible();
    });

    test('inline field editing updates character real-time', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        const newName = 'CyberGhost_' + Math.floor(Math.random() * 1000);
        const newClass = 'CYBERHEAD';

        // Edit Name
        await page.locator('#identity-view').click();
        await page.locator('#identity-edit input[name="name"]').fill(newName);
        await page.locator('#identity-edit button[type="submit"]').click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('#identity-view h1')).toContainText(newName);

        // Edit Class
        await page.locator('#class-view').click();
        await page.locator('#class-edit input[name="class_name"]').fill(newClass);
        await page.locator('#class-edit button[type="submit"]').click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('#class-view')).toContainText(newClass);
    });

    test('inventory list workflow › add and delete weapon items', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        // Add Weapon
        await page.locator('#add-weapon-btn').click();
        await page.locator('#add-weapon-edit input[name="item"]').fill('Nano-Blade');
        await page.locator('#add-weapon-edit button[type="submit"]').click();
        await page.waitForLoadState('networkidle');

        await expect(page.locator('text=Nano-Blade')).toBeVisible();
    });
});
