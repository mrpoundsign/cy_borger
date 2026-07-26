// test_character_graveyard.js — Playwright test for GM & Owner Kill/Flatline and Graveyard feature.
const { test, expect } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('Graveyard & Flatline Workflow', () => {
    test('owner or GM can flatline a character with death note, moving to Graveyard', async ({ page }) => {
        // 1. Go home and create blank character
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

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
        await expect(page.locator('text=OPERATOR FLATLINED / IN GRAVEYARD')).toBeVisible();
        await expect(page.locator('text=Crushed by rogue cyber-mech in Sector 7')).toBeVisible();

        // 4. Test Revive
        await page.locator('button:has-text("⚡ REVIVE OPERATOR")').click();
        await page.waitForLoadState('networkidle');

        // Confirm flatline banner is gone
        await expect(page.locator('text=OPERATOR FLATLINED / IN GRAVEYARD')).not.toBeVisible();
    });

    test('kill button on game page opens modal and moves character to Graveyard', async ({ page }) => {
        // 1. Create a game as GM
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        await page.locator('input[name="name"]').fill('Sector 4 Campaign');
        await page.locator('.card button:has-text("Create Game as GM")').click();
        await page.waitForLoadState('networkidle');

        const gameUrl = page.url();
        expect(gameUrl).toContain('/game/');

        // 2. Click "Roll New Character" on game page
        await page.locator('button:has-text("🎲 Roll New Character")').click();
        await page.waitForLoadState('networkidle');

        // Go back to game page
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');

        // 3. Click 💀 KILL button on game page
        const killBtn = page.locator('button:has-text("💀 KILL")').first();
        await expect(killBtn).toBeVisible();
        await killBtn.click();
        await page.waitForTimeout(300);

        // Fill death note in game page modal
        const modal = page.locator('div[id^="kill-game-"]:not([style*="display: none"])');
        await expect(modal).toBeVisible();
        await modal.locator('textarea[name="death_note"]').fill('Shot by Corp assassin on game page');
        await modal.locator('button[type="submit"]').click();
        await page.waitForLoadState('networkidle');

        // 4. Verify character appears in GRAVEYARD section on game page
        await expect(page.locator('text=GRAVEYARD / FLATLINED OPERATORS')).toBeVisible();
        await expect(page.locator('text=Shot by Corp assassin on game page')).toBeVisible();
    });
});
