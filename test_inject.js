// test_inject.js — Tests that the INSPECT SHEET button loads a character sheet
// snippet into the modal without injecting a full HTML page (DOCTYPE, head, body).
//
// Run: npx playwright test test_inject.js
//
const { test, expect, request } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

/** Wait for air to finish rebuilding after a template change. */
const AIR_SETTLE_MS = 5000;

// ── Helpers ──────────────────────────────────────────────────────────────────

/** POST to generate a new character; returns the character page URL. */
async function generateCharacter(apiCtx) {
    const res = await apiCtx.post(`${BASE_URL}/character/generate`);
    // air serves the redirect, so follow and grab the final URL
    return res.url();
}

/** POST to create a game; returns the game page URL. */
async function createGame(apiCtx, name = 'Inspect Test Game') {
    const res = await apiCtx.post(`${BASE_URL}/game/create`, {
        form: { name },
    });
    return res.url();
}

/** Extract path segment after /character/ or /game/ */
function extractId(url, segment) {
    const match = url.match(new RegExp(`/${segment}/([^/?]+)`));
    return match ? match[1] : null;
}

/** Read the invite code from a game page. */
async function getInviteCode(page, gameUrl) {
    await page.goto(gameUrl);
    await page.waitForLoadState('networkidle');
    const el = page.locator('.invite-box strong').first();
    return el.innerText();
}

/** Join a character to a game by invite code (server-side POST). */
async function joinGame(apiCtx, charId, inviteCode) {
    return apiCtx.post(`${BASE_URL}/character/${charId}/join`, {
        form: { invite_code: inviteCode },
    });
}

// ── Tests ────────────────────────────────────────────────────────────────────

test.describe('INSPECT SHEET modal injection', () => {
    let apiCtx;
    let charUrl;
    let charId;
    let gameUrl;
    let gameId;
    let inviteCode;

    test.beforeAll(async ({ playwright }) => {
        apiCtx = await request.newContext({ baseURL: BASE_URL });

        // Generate a character
        charUrl = await generateCharacter(apiCtx);
        charId = extractId(charUrl, 'character');
        expect(charId, 'Character ID extracted').toBeTruthy();

        // Create a game
        gameUrl = await createGame(apiCtx);
        gameId = extractId(gameUrl, 'game');
        expect(gameId, 'Game ID extracted').toBeTruthy();
    });

    test.afterAll(async () => {
        await apiCtx.dispose();
    });

    test('reads invite code from game page', async ({ page }) => {
        inviteCode = await getInviteCode(page, gameUrl);
        expect(inviteCode, 'Invite code present').toBeTruthy();
    });

    test('joins character to game', async () => {
        const res = await joinGame(apiCtx, charId, inviteCode);
        // 200 or 303 redirect — either is success
        expect([200, 303]).toContain(res.status());
    });

    test('party grid shows character card after join', async ({ page }) => {
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');
        await expect(page.locator('.char-card')).toHaveCount(1, { timeout: 5000 });
    });

    test('INSPECT SHEET button is visible', async ({ page }) => {
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');
        await expect(page.locator('button:has-text("INSPECT SHEET")').first()).toBeVisible();
    });

    test('clicking INSPECT SHEET loads snippet — no DOCTYPE/html/head/body injected', async ({ page }) => {
        const consoleErrors = [];
        page.on('console', msg => {
            if (msg.type() === 'error') consoleErrors.push(msg.text());
        });

        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');

        // Click inspect
        await page.locator('button:has-text("INSPECT SHEET")').first().click();

        // Wait for the modal to appear
        await expect(page.locator('#char-modal')).toBeVisible({ timeout: 8000 });

        // Check the raw HTML inside the container for full-page pollution
        const containerHtml = await page.locator('#char-modal-container').innerHTML();
        expect(containerHtml, 'No <!DOCTYPE> injected').not.toContain('<!DOCTYPE');
        expect(containerHtml, 'No <html> tag injected').not.toContain('<html');
        expect(containerHtml, 'No <head> tag injected').not.toContain('<head>');
        expect(containerHtml, 'No <body> tag injected').not.toContain('<body');

        // Character sheet content should be present
        await expect(page.locator('#char-modal .toolbar')).toBeVisible();
        await expect(page.locator('#char-modal .vitals-box')).toBeVisible();
        await expect(page.locator('#char-modal h1')).toContainText('YOU ARE');

        // No JS errors during load
        expect(consoleErrors, `JS errors: ${consoleErrors.join(', ')}`).toHaveLength(0);
    });

    test('modal closes when CLOSE button clicked', async ({ page }) => {
        await page.goto(gameUrl);
        await page.waitForLoadState('networkidle');

        await page.locator('button:has-text("INSPECT SHEET")').first().click();
        await expect(page.locator('#char-modal')).toBeVisible({ timeout: 8000 });

        await page.locator('#char-modal button:has-text("CLOSE")').click();
        await expect(page.locator('#char-modal')).not.toBeVisible({ timeout: 3000 });
    });

    test('direct character page still renders full page correctly', async ({ page }) => {
        await page.goto(charUrl);
        await page.waitForLoadState('networkidle');

        await expect(page).toHaveTitle(/CY_BORG/);
        await expect(page.locator('body h1')).toContainText('YOU ARE');
        await expect(page.locator('.vitals-box')).toBeVisible();
        await expect(page.locator('.toolbar')).toBeVisible();
    });
});
