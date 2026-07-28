// test_character_flatline.js — Playwright test for flatlining a character from the character page.
const { test, expect } = require('@playwright/test');

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
        const name = 'FlatlineOp' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test.describe('Standalone Character Flatline Workflow', () => {
    test('owner can flatline a character from the character sheet', async ({ page }) => {
        await loginUser(page);

        // 1. Create a standalone character
        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForSelector('#char-data-id', { state: 'attached', timeout: 10000 });
        const charIdData = await page.locator('#char-data-id').getAttribute('data-id');
        const url = (process.env.BASE_URL || 'http://localhost:8080') + '/character/' + charIdData;
        const charID = url.split('/character/')[1];
        expect(charID).toBeTruthy();

        // 2. Click 💀 FLATLINE button in toolbar
        await page.locator('button[id^="btn-flatline-sheet-"]').click();
        await page.waitForTimeout(300);

        // Fill death note
        const deathNoteInput = page.locator('#kill-char-edit textarea[name="death_note"]');
        await deathNoteInput.fill('Crushed by rogue cyber-mech in Sector 7');

        // Submit flatline inside modal
        await page.locator('#kill-char-edit button[type="submit"]').click();
        // Wait for WebSocket to trigger a page refresh via HTMX
        // Playwright's expect.toBeVisible() will automatically wait/retry up to 10s.
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(1000);

        // 3. Confirm OPERATOR FLATLINED banner appears on character sheet
        await expect(page.locator('.text-danger', { hasText: 'OPERATOR FLATLINED' }).first()).toBeVisible({ timeout: 10000 });
        await expect(page.locator('text=Crushed by rogue cyber-mech in Sector 7')).toBeVisible();

        // 4. Test Revive
        await page.locator('button:has-text("⚡ REVIVE OPERATOR")').click();
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(1000); // Wait for HTMX swap

        // Confirm flatline banner is gone
        await expect(page.locator('.text-danger', { hasText: 'OPERATOR FLATLINED' })).toHaveCount(0);
    });
});
