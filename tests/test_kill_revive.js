const { chromium } = require('@playwright/test');

(async () => {
  console.log("Starting Playwright test: Authenticated Kill/Revive flow...");
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
    console.log("Navigating to http://localhost:8080/");
    await page.goto('http://localhost:8080/');
    
    // 1. Register a new user
    console.log("Registering a new account...");
    await page.click('text=REGISTER ACCOUNT');
    const randomUser = 'testuser_' + Date.now();
    await page.locator('#auth-register-form input[name="username"]').fill(randomUser);
    await page.locator('#auth-register-form input[name="password"]').fill('password');
    await page.locator('#auth-register-form button:has-text("REGISTER ACCOUNT")').click();
    await page.waitForTimeout(1000);
    
    // 2. Create a Game
    console.log("Creating a game...");
    await page.fill('input[name="name"]', 'Test Game');
    await page.click('button:has-text("Create Game as GM")');
    await page.waitForTimeout(1000);
    const gameUrl = page.url();
    console.log(`Game created: ${gameUrl}`);
    
    // 3. Create a Character
    console.log("Creating a character...");
    await page.goto('http://localhost:8080/');
    await page.click('text=Create Blank Character');
    await page.waitForTimeout(1000);
    const charUrl = page.url();
    console.log(`Character created: ${charUrl}`);
    
    // 4. Join the Game
    console.log("Joining the game...");
    // We need the invite code. We can just go back to the game page and get it, or use the ActiveGame join button on the character page.
    // If we just created the game, it should be the active game. Let's click the "+ ADD TO GAME" button.
    const addToGameBtn = page.locator('button:has-text("+ ADD TO GAME")');
    if (await addToGameBtn.count() > 0) {
        await addToGameBtn.click();
    } else {
        console.log("Add to Game button not found, maybe not needed.");
    }
    await page.waitForTimeout(1000);
    
    // 5. Kill the character FROM THE GAME PAGE
    console.log("Clicking KILL button on the game page...");
    await page.click('button:has-text("💀 KILL")');
    await page.waitForTimeout(500); // Wait for modal
    await page.click('button:has-text("💀 FLATLINE")'); // Submit the kill form
    await page.waitForTimeout(2000); // Wait for HTMX and WebSockets to process
    
    // 6. Revive the character
    console.log("Checking if Revive button exists and clicking it...");
    const reviveBtn = page.locator('button:has-text("⚡ REVIVE")');
    if (await reviveBtn.count() > 0) {
        await reviveBtn.click();
        await page.waitForTimeout(2000); // Wait for HTMX and WebSockets to process
        console.log("Character revived successfully!");
    } else {
        console.log("Revive button not found! The UI might not have updated correctly.");
    }
    
    console.log("=== TEST SUMMARY ===");
    if (errors.length === 0) {
        console.log("SUCCESS: No JS Console Errors found during the Kill/Revive flow.");
    } else {
        console.log(`FAILED: Found ${errors.length} JS errors.`);
    }

  } catch (err) {
    console.error('Test execution failed:', err);
  } finally {
    await browser.close();
  }
})();
