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
        const name = 'LayoutUser' + Math.floor(1000 + Math.random() * 9000);
        await page.locator('#auth-register-form input[name="username"]').fill(name);
        await page.locator('#auth-register-form input[name="password"]').fill('pass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
    }
}

test('Verify modal layout and flex classes', async ({ page }) => {
    await loginUser(page);
    
    await page.fill('input[name="name"]', 'Test Game');
    await page.click('#btn-create-game-index');
    await page.waitForURL(/\/game\//);
    
    // Give time for WebSocket to connect and logs to load
    await page.waitForTimeout(1000);

    // Create a blank character in the game
    await page.locator('button:has-text("Create Blank")').first().click();
    await page.waitForURL(/\/character\//);

    // Go back to the game page
    await Promise.all([
        page.waitForNavigation(),
        page.locator('text=⬅️ Back to Home').click()
    ]);
    
    // Click into the game we just created
    await Promise.all([
        page.waitForNavigation(),
        page.locator('text=[GM Mode]').first().click()
    ]);

    // Wait for the character card to appear
    await page.waitForSelector('.char-card');
    
    // Click 'INSPECT SHEET' on the first character
    await page.locator('button:has-text("⚡ INSPECT SHEET")').first().click();

    // Wait for modal to appear
    await page.waitForSelector('.char-modal-overlay');

    // Assert that the modal-content has the expected width (approx 1100px or full width if screen is smaller)
    const modalContent = page.locator('.char-modal-content');
    const box = await modalContent.boundingBox();
    console.log(`Modal width: ${box.width}px`);
    expect(box.width).toBeGreaterThan(800); // Should be wide

    // Assert that flex-1 classes are applied and working by checking the width of one of the columns
    const flex1Elements = page.locator('.flex-1');
    const count = await flex1Elements.count();
    console.log(`Found ${count} flex-1 elements`);
    expect(count).toBeGreaterThan(0);

    if (count > 0) {
        const flexBox = await flex1Elements.first().boundingBox();
        console.log(`First flex-1 column width: ${flexBox.width}px`);
        expect(flexBox.width).toBeGreaterThan(200); // Should be taking up significant space
    }
});
