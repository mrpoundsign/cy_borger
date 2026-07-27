const { test, expect } = require('@playwright/test');

test('Verify modal layout and flex classes', async ({ page }) => {
    // Navigate to homepage and create a new game
    await page.goto('http://localhost:8080/');
    
    await page.fill('input[name="name"]', 'Test Game');
    await Promise.all([
        page.waitForNavigation(),
        page.locator('button:has-text("Create Game as GM")').click()
    ]);
    
    // Give time for WebSocket to connect and logs to load
    await page.waitForTimeout(1000);

    // Create a blank character in the game
    await Promise.all([
        page.waitForNavigation(),
        page.locator('button:has-text("CREATE BLANK")').click()
    ]);

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
