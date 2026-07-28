const { test, expect } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

async function registerUser(page, prefix) {
    await page.goto(BASE_URL + '/');
    await page.waitForLoadState('networkidle');

    if (await page.locator('button:has-text("🚪 LOGOUT")').isVisible()) {
        await page.locator('button:has-text("🚪 LOGOUT")').click();
        await page.waitForLoadState('networkidle');
    }

    if (await page.locator('#tab-btn-register').isVisible()) {
        await page.locator('#tab-btn-register').click();
        await page.waitForTimeout(200);
        const name = prefix + '_' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test('Real-time WebSocket updates on standalone character sheet from off-page GM actions', async ({ browser }) => {
    // 1. Create separate browser contexts for GM and Player
    const gmContext = await browser.newContext();
    const gmPage = await gmContext.newPage();

    const playerContext = await browser.newContext();
    const playerPage = await playerContext.newPage();

    try {
        // 2. GM registers and creates a game
        await registerUser(gmPage, 'GMUser');
        await gmPage.fill('input[name="name"]', 'WS Offpage Test Game');
        await gmPage.click('#btn-create-game-index');
        await gmPage.waitForSelector('#game-data-invite', { state: 'attached', timeout: 10000 });
        
        const inviteCode = await gmPage.getAttribute('#game-data-invite', 'data-code');
        expect(inviteCode).toBeTruthy();

        // 3. Player registers, creates a character, and joins GM's game using the invite code
        await registerUser(playerPage, 'PlayerUser');
        await playerPage.click('text=Create Blank Character');
        await playerPage.waitForURL(/\/character\//);
        const charUrl = playerPage.url();
        const charId = charUrl.split('/character/')[1].split('?')[0];

        // Join game via invite code on character sheet
        await playerPage.fill('input[name="invite_code"]', inviteCode);
        await playerPage.click('#btn-join-game-' + charId);
        await playerPage.waitForURL(/\/game\//, { timeout: 10000 });

        // Navigate back to standalone character sheet so player is watching /character/:id
        await playerPage.goto(charUrl);
        await playerPage.waitForLoadState('networkidle');
        await playerPage.waitForTimeout(1000);

        // Verify player is on character page and character is alive
        await expect(playerPage.locator('#character-banners')).not.toContainText('OPERATOR FLATLINED');

        // 4. GM (on gmPage) flatlines Player's character from the GM party grid
        await gmPage.reload();
        await gmPage.waitForSelector('.char-card', { timeout: 10000 });

        const flatlineCardBtn = gmPage.locator('#btn-flatline-' + charId + '-card');
        await flatlineCardBtn.click();
        await gmPage.waitForTimeout(500);

        const deathNoteInput = gmPage.locator('#kill-game-' + charId + '-edit textarea[name="death_note"]');
        await deathNoteInput.fill('Flatlined via off-page GM action');
        
        const submitKillBtn = gmPage.locator('#btn-submit-flatline-' + charId + '-modal');
        await submitKillBtn.click();
        await gmPage.waitForTimeout(1500);

        // Verify GM page updated and character moved to Graveyard
        await expect(gmPage.locator('.party-container')).toContainText('GRAVEYARD / FLATLINED OPERATORS', { timeout: 5000 });

        // 5. ASSERTION: Player's page (playerPage) automatically updates via WebSocket to show FLATLINED banner without page reload
        await expect(playerPage.locator('#character-banners')).toContainText('OPERATOR FLATLINED', { timeout: 10000 });
        await expect(playerPage.locator('#character-banners')).toContainText('Flatlined via off-page GM action');

        // 6. GM (on gmPage) revives Player's character from the Graveyard
        const reviveBtn = gmPage.locator('button:has-text("⚡ REVIVE OPERATOR")').first();
        if (await reviveBtn.isVisible()) {
            await reviveBtn.click();
            await gmPage.waitForTimeout(1500);
        } else {
            // Alternatively revive via player page or inspect modal
            const inspectBtn = gmPage.locator('button:has-text("⚡ INSPECT SHEET")').first();
            await inspectBtn.click();
            await gmPage.waitForSelector('#char-modal', { state: 'visible', timeout: 5000 });
            await gmPage.locator('#char-modal button:has-text("⚡ REVIVE OPERATOR")').click();
            await gmPage.waitForTimeout(1500);
        }

        // 7. ASSERTION: Player's page (playerPage) automatically updates via WebSocket to clear FLATLINED banner
        await expect(playerPage.locator('#character-banners')).not.toContainText('OPERATOR FLATLINED', { timeout: 10000 });

    } finally {
        await gmContext.close();
        await playerContext.close();
    }
});
