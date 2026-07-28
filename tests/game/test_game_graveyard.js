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
    test('flatline button on game page opens modal and moves character to Graveyard', async ({ page }) => {
        await loginUser(page);

        await page.locator('input[name="name"]').fill('Sector 4 Campaign');
        await page.locator('.card button:has-text("Create Game as GM")').click();
        await page.waitForSelector('#game-data', { state: 'attached', timeout: 10000 });
        const gameId = await page.locator('#game-data').getAttribute('data-id');
        const gameUrl = (process.env.BASE_URL || 'http://localhost:8080') + '/game/' + gameId;

        // 2. Roll a character into game
        await page.locator('button:has-text("🎲 Roll New Character")').click();
        await page.waitForLoadState('networkidle');

        const keepBtn = page.locator('button:has-text("KEEP THIS CHARACTER")');
        await keepBtn.click();
        await expect(keepBtn).toBeHidden({ timeout: 10000 });
        await page.waitForLoadState('networkidle');

        // Return to game page
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(1500);

        // 3. Click 💀 FLATLINE button on game page
        const killBtn = page.locator('button:has-text("💀 FLATLINE")').first();
        await expect(killBtn).toBeVisible({ timeout: 10000 });
        await killBtn.click();
        await page.waitForTimeout(300);

        // Fill death note
        const deathNoteInputGame = page.locator('div[id^="kill-game-"] textarea[name="death_note"]').first();
        await deathNoteInputGame.fill('Killed by orbital strike');

        // Submit kill form modal
        const killSubmit = page.locator('div[id^="kill-game-"] button[type="submit"]').first();
        await killSubmit.click();
        await page.waitForLoadState('networkidle');

        // 4. Confirm character is in GRAVEYARD section on game page
        await expect(page.locator('text=GRAVEYARD')).toBeVisible();
        await expect(page.locator('text=Killed by orbital strike')).toBeVisible();
        await expect(page.locator('button:has-text("REVIVE")')).toBeVisible();
    });
});
