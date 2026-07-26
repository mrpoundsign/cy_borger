const { chromium } = require('@playwright/test');

(async () => {
  console.log("Starting Playwright test to trigger WebSocket HTMX error...");
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  let errors = [];
  page.on('console', msg => {
      if (msg.type() === 'error' || msg.text().includes('TypeError') || msg.text().includes('htmx:swapError')) {
          errors.push(msg.text());
          console.error(`[BROWSER ERROR] ${msg.text()}`);
      }
  });
  page.on('pageerror', err => {
      errors.push(err.message);
      console.error(`[PAGE ERROR] ${err.message}`);
  });

  try {
    await page.goto('http://localhost:8080/');
    
    // Register
    await page.click('text=REGISTER ACCOUNT');
    const randomUser = 'testuser_' + Date.now();
    await page.locator('#auth-register-form input[name="username"]').fill(randomUser);
    await page.locator('#auth-register-form input[name="password"]').fill('password');
    await page.locator('#auth-register-form button:has-text("REGISTER ACCOUNT")').click();
    await page.waitForTimeout(1000);
    
    // Create Character
    await page.goto('http://localhost:8080/');
    await page.click('text=Create Blank Character');
    await page.waitForTimeout(1000);
    const charUrl = page.url();
    
    console.log(`Navigating to character page: ${charUrl}`);
    await page.goto(charUrl);
    await page.waitForTimeout(1000);
    
    // Explicitly trigger the HTMX ajax call that the WebSocket handler uses
    console.log("Triggering htmx.ajax with target: 'body', swap: 'innerHTML' to simulate WebSocket update...");
    await page.evaluate(() => {
        // This is what my current fix does
        htmx.ajax('GET', window.location.pathname, { target: 'body', swap: 'innerHTML' });
    });
    
    await page.waitForTimeout(2000); // wait for HTMX to process

    console.log("Triggering htmx.ajax with target: 'body', swap: 'outerHTML' to simulate original bug...");
    await page.evaluate(() => {
        // This is the original buggy code
        htmx.ajax('GET', window.location.pathname, { target: 'body', swap: 'outerHTML' });
    });
    
    await page.waitForTimeout(2000);

    console.log("=== TEST SUMMARY ===");
    if (errors.length === 0) {
        console.log("SUCCESS: No errors found.");
    } else {
        console.log(`FAILED: Found ${errors.length} JS errors.`);
    }

  } catch (err) {
    console.error('Test execution failed:', err);
  } finally {
    await browser.close();
  }
})();
