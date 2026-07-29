// test_inject.js — Tests that the INSPECT SHEET button loads a character sheet
// snippet into the modal without injecting a full HTML page (DOCTYPE, head, body).
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
        const name = 'InjOp' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test.describe('INSPECT SHEET modal injection', () => {
    test('game page allows GM to roll character and inspect sheet modal', async ({ page }) => {
        await loginUser(page);

        // Create a game as GM
        await page.locator('input[name="name"]').fill('Inspect Test Game');
        await page.locator('.card button:has-text("Create Game as GM")').click();
        await page.waitForSelector('#game-data', { state: 'attached', timeout: 10000 });
        const gameId = await page.locator('#game-data').getAttribute('data-id');
        const gameUrl = (process.env.BASE_URL || 'http://localhost:8080') + '/game/' + gameId;

        // Roll a character into game
        await page.locator('a:has-text("🎲 Roll New Character")').click();
        await page.waitForLoadState('networkidle');

        // Character is generated as a draft; keep it to save it to the game
        const keepBtn = page.locator('button:has-text("KEEP THIS CHARACTER")');
        await keepBtn.click();
        await expect(keepBtn).toBeHidden({ timeout: 10000 });
        await page.waitForLoadState('networkidle');

        const charUrl = page.url();

        // Go back to game page
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(1500);
        await page.waitForTimeout(500);


        // Party grid shows character card
        await page.waitForTimeout(1500);
        await expect(page.locator('.char-card')).toHaveCount(1, { timeout: 10000 });

        // INSPECT SHEET button is visible
        const inspectBtn = page.locator('button:has-text("INSPECT SHEET")').first();
        await expect(inspectBtn).toBeVisible();

        // Click INSPECT SHEET
        await inspectBtn.click();
        await expect(page.locator('#char-modal')).toBeVisible({ timeout: 8000 });

        // Verify HTML container does not inject full-page tags
        const containerHtml = await page.locator('#char-modal-container').innerHTML();
        expect(containerHtml).not.toContain('<!DOCTYPE');
        expect(containerHtml).not.toContain('<html');
        expect(containerHtml).not.toContain('<head');
        expect(containerHtml).not.toContain('<body');

        // Close modal
        await page.locator('button:has-text("✕ CLOSE")').click();
        await expect(page.locator('#char-modal')).not.toBeVisible();

        // Character page renders full page correctly
        await page.goto(charUrl);
        await page.waitForLoadState('networkidle');
        await expect(page).toHaveTitle(/CY_BORGER/);
        await expect(page.locator('body')).toBeVisible();
    });
});
