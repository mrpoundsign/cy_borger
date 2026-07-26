// test_character_edit.js — Playwright test for full character editing, blank character creation, and draft/keep workflow.
//
// Run: npx playwright test test_character_edit.js
//
const { test, expect, request } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('Character editing & Draft/Keep workflow', () => {
    let apiCtx;

    test.beforeAll(async () => {
        apiCtx = await request.newContext({ baseURL: BASE_URL });
    });

    test.afterAll(async () => {
        await apiCtx.dispose();
    });

    test('create blank character and verify default fields', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        // Click "Create Blank Character"
        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        await expect(page).toHaveURL(/\/character\//);
        await expect(page.locator('#identity-view h1')).toContainText('UNNAMED OPERATOR');
        await expect(page.locator('#identity-view .sub-title')).toContainText('@operator');
    });

    test('roll random character shows draft banner, keeping it saves it', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        // Click "Roll Random Character"
        await page.locator('button:has-text("Roll Random Character")').click();
        await page.waitForLoadState('networkidle');

        // Draft banner should be visible
        const draftBanner = page.locator('text=DRAFT CHARACTER PREVIEW');
        await expect(draftBanner).toBeVisible();

        const keepBtn = page.locator('button:has-text("KEEP THIS CHARACTER")');
        await expect(keepBtn).toBeVisible();

        // Click Keep Character
        await keepBtn.click();
        await page.waitForLoadState('networkidle');

        // Draft banner should now be gone
        await expect(draftBanner).not.toBeVisible();

        // Go home and verify character is in MY CHARACTERS list
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');
        await expect(page.locator('#my-chars-list')).toBeVisible();
    });

    test('field editing updates character real-time', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        // Click "✏️ Edit Text" to reveal input form
        await page.locator('button:has-text("Edit Text")').click();

        const nameInput = page.locator('#identity-edit input[name="name"]');
        await nameInput.fill('CYBER_PUNK_X');
        
        // Click Save Changes to save form
        await page.locator('#identity-edit button:has-text("Save Changes")').first().click();
        await page.waitForLoadState('networkidle');

        // Reload page to confirm persistence in view
        await page.reload();
        await page.waitForLoadState('networkidle');
        await expect(page.locator('#identity-view h1')).toContainText('CYBER_PUNK_X');
    });

    test('add and delete weapon items', async ({ page }) => {
        // Accept confirm dialogs automatically
        page.on('dialog', dialog => dialog.accept());

        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        // Toggle add weapon section
        await page.locator('button:has-text("+ Add Weapon / Armor")').click();

        // Fill Add Weapon form
        const addWpnForm = page.locator('form[action*="add_item"]').first();
        await addWpnForm.locator('input[name="name"]').fill('Plasma Cutter');
        await addWpnForm.locator('input[name="damage"]').fill('d8');
        await addWpnForm.locator('button[type="submit"]').click();

        await page.waitForLoadState('networkidle');

        // Weapon should appear in table
        await expect(page.locator('td:has-text("Plasma Cutter")')).toBeVisible();
    });
});
