const { test, expect } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('Authenticated Kill/Revive flow', () => {
    test('user can register, create game, create character, flatline and revive', async ({ page }) => {
        let errors = [];
        page.on('console', msg => {
            const text = msg.text();
            if (msg.type() === 'error' && !text.includes('ERR_INCOMPLETE_CHUNKED_ENCODING') && !text.includes('htmx:sendError') && !text.includes('htmx:afterRequest')) {
                errors.push(text);
                console.error(`[BROWSER ERROR] ${text}`);
            } else if (text.includes('TypeError') || text.includes('htmx:swapError')) {
                errors.push(text);
                console.error(`[BROWSER ERROR] ${text}`);
            }
        });
        page.on('pageerror', err => {
            errors.push(err.message);
            console.error(`[PAGE ERROR] ${err.message}`);
        });

        await page.goto(BASE_URL + '/');
        
        // 1. Register a new user
        await page.click('text=REGISTER ACCOUNT');
        const randomUser = 'testuser_' + Date.now();
        await page.locator('#auth-register-form input[name="username"]').fill(randomUser);
        await page.locator('#auth-register-form input[name="password"]').fill('password');
        await page.locator('#auth-register-form button:has-text("REGISTER ACCOUNT")').click();
        await page.waitForLoadState('networkidle');
        
        // 2. Create a Game
        await page.fill('input[name="name"]', 'Test Game');
        await Promise.all([
            page.waitForURL('**/game/**'),
            page.click('#btn-create-game-index')
        ]);
        
        // 3. Create a Character
        await page.goto(BASE_URL + '/');
        await Promise.all([
            page.waitForURL('**/characters/new**'),
            page.locator('#btn-generate-character-index').click()
        ]);
        
        const keepBtn = page.locator('button:has-text("KEEP THIS CHARACTER")');
        await Promise.all([
            page.waitForURL('**/character/**'),
            keepBtn.click()
        ]);
        await expect(keepBtn).toBeHidden();
        
        // 4. Join the Game
        const addToGameBtn = page.locator('button:has-text("+ ADD TO GAME")');
        if (await addToGameBtn.count() > 0) {
            await Promise.all([
                page.waitForURL('**/game/**'),
                addToGameBtn.click()
            ]);
            // Now on game page, we can see the character card
            await expect(page.locator('.char-card').first()).toBeVisible();
        }
        
        // 5. Flatline the character directly from the dashboard card
        await page.locator('button[id^="btn-flatline-"]').first().click();
        await page.locator('div[id^="kill-game-"] textarea[name="death_note"]').fill('TEST DEATH');
        await page.locator('div[id^="kill-game-"] button[type="submit"]:has-text("💀 FLATLINE")').click();
        
        // Wait for it to become dead
        await expect(page.locator('text="FLATLINED"')).toBeVisible();
        
        // 6. Revive the character
        const reviveBtn = page.locator('button[id^="btn-revive-"]').first();
        await reviveBtn.click();
        
        // Wait for it to be revived
        await expect(reviveBtn).toBeHidden();
        
        expect(errors).toHaveLength(0);
    });
});
