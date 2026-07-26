const { chromium } = require('@playwright/test');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  let errors = [];
  page.on('console', msg => {
      if (msg.type() === 'error' || msg.text().includes('TypeError') || msg.text().includes('htmx:swapError')) {
          errors.push(msg.text());
      }
  });
  page.on('pageerror', err => {
      errors.push(err.message);
  });

  try {
    await page.goto('http://localhost:8081/');
    // Login or register
    await page.click('text=REGISTER ACCOUNT');
    await page.locator('#auth-register-form input[name="username"]').fill('testuser123');
    await page.locator('#auth-register-form input[name="password"]').fill('password');
    await page.locator('#auth-register-form button:has-text("REGISTER ACCOUNT")').click();
    await page.waitForTimeout(1000);
    
    // Create blank character
    await page.click('text=Create Blank Character');
    await page.waitForTimeout(1000);
    
    // Trigger websocket update by adding gear
    await page.click('text=+ Add Weapon / Armor');
    await page.waitForTimeout(500);
    await page.fill('input[placeholder="Weapon Name"]', 'Knife');
    await page.click('button:has-text("+ Add")');
    await page.waitForTimeout(1500); // wait for websocket message and HTMX swap
    
    console.log("=== PLAYWRIGHT CONSOLE ERRORS ===");
    if (errors.length === 0) {
        console.log("No errors found!");
    } else {
        errors.forEach(e => console.log(e));
    }
    console.log("=================================");

  } catch (err) {
    console.error('Test script failed:', err);
  }
  await browser.close();
})();
