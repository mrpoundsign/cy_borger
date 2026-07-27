const { test, expect, chromium } = require('@playwright/test');

test('Check HTMX swapError', async () => {
    const browser = await chromium.launch();
    const context = await browser.newContext();
    const page = await context.newPage();

    let errors = [];
    page.on('console', msg => {
        if (msg.type() === 'error' || msg.text().includes('htmx:swapError')) {
            errors.push(msg.text());
        }
    });
    
    // Create a new game
    await page.goto('http://localhost:8080/');
    await page.click('text=CREATE A NEW GAME');
    await page.waitForURL('**/game/*');
    
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
    
    await browser.close();
});
