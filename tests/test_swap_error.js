const { test, expect, chromium } = require('@playwright/test');

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
        const name = 'SwapUser' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test('Check HTMX swapError', async ({ page }) => {
    let errors = [];
    page.on('console', msg => {
        if (msg.type() === 'error' || msg.text().includes('htmx:swapError')) {
            errors.push(msg.text());
        }
    });
    
    await loginUser(page);

    // Create a new game
    await page.fill('input[name="name"]', 'Swap Test Game');
    await page.click('#btn-create-game-index');
    await page.waitForSelector('#game-data', { state: 'attached', timeout: 10000 });
    
    console.log("On game page, waiting to see if there are errors...");
    await page.waitForTimeout(2000);
    
    // Check checkboxes
    await page.click('text=STATS & HP');
    await page.waitForTimeout(1000);
    
    // We can't kill a character since there are none, but let's try creating one and killing it.
    
    if (errors.length > 0) {
        console.log("FOUND ERRORS:", errors);
    } else {
        console.log("No console errors found.");
    }
});
