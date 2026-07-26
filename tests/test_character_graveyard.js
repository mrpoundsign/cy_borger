// test_character_graveyard.js — Playwright test for GM & Owner Kill/Flatline and Graveyard feature.
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
        const name = 'GraveOp' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test.describe('Graveyard & Flatline Workflow', () => {
    test('owner or GM can flatline a character with death note, moving to Graveyard', async ({ page }) => {
        await loginUser(page);

        await page.locator('button:has-text("Create Blank Character")').click();
        await page.waitForLoadState('networkidle');

        // Get character ID
        const url = page.url();
        const charID = url.split('/character/')[1];
        expect(charID).toBeTruthy();

        // 2. Click 💀 FLATLINE button in toolbar
        await page.locator('button.btn-danger-small[onclick*="kill-char"]').click();
        await page.waitForTimeout(300);

        // Fill death note
        const deathNoteInput = page.locator('#kill-char-edit textarea[name="death_note"]');
        await deathNoteInput.fill('Crushed by rogue cyber-mech in Sector 7');

        // Submit flatline inside modal
        await page.locator('#kill-char-edit button[type="submit"]:has-text("💀 FLATLINE")').click();
        await page.waitForLoadState('networkidle');

        // 3. Confirm GRAVEYARD banner appears on character sheet
        await expect(page.locator('text=OPERATOR FLATLINED')).toBeVisible();
        await expect(page.locator('text=Crushed by rogue cyber-mech in Sector 7')).toBeVisible();

        // 4. Test Revive
        await page.locator('button:has-text("⚡ REVIVE OPERATOR")').click();
        await page.waitForLoadState('networkidle');

        // Confirm flatline banner is gone
        await expect(page.locator('text=OPERATOR FLATLINED')).not.toBeVisible();
    });

    test('kill button on game page opens modal and moves character to Graveyard', async ({ page }) => {
        await loginUser(page);

        await page.locator('input[name="name"]').fill('Sector 4 Campaign');
        await page.locator('.card button:has-text("Create Game as GM")').click();
        await page.waitForLoadState('networkidle');

        const gameUrl = page.url();

        // 2. Roll a character into game
        await Promise.all([
            page.waitForNavigation(),
            page.locator('button:has-text("🎲 Roll New Character")').click()
        ]);

        const keepBtn = page.locator('button:has-text("KEEP THIS CHARACTER")');
        await keepBtn.click();
        await expect(keepBtn).toBeHidden({ timeout: 10000 });
        await page.waitForLoadState('networkidle');

        // Return to game page
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');

        // 3. Click 💀 KILL button on game page
        const killBtn = page.locator('button:has-text("💀 KILL")').first();
        await expect(killBtn).toBeVisible({ timeout: 10000 });
        await killBtn.click();
        await page.waitForTimeout(300);

        // Submit kill form modal
        const killSubmit = page.locator('#kill-game-edit button[type="submit"]:has-text("💀 FLATLINE"), div[id^="kill-game-"] button[type="submit"]:has-text("💀 FLATLINE")').first();
        await killSubmit.click();
        await page.waitForLoadState('networkidle');

        // 4. Confirm character is in GRAVEYARD section on game page
        await expect(page.locator('text=GRAVEYARD')).toBeVisible();
        await expect(page.locator('button:has-text("REVIVE")')).toBeVisible();
    });
});
