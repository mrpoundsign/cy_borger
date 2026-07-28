const { chromium } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  // NOTE: This uses a hardcoded character ID which may not exist.
  const url = `${BASE_URL}/character/8ba9e6b971492cf0`;
  console.log('Navigating to', url);
  
  let errors = [];
  page.on('console', msg => {
      if (msg.type() === 'error') {
          errors.push(msg.text());
      }
  });

  try {
    await page.goto(url);
    await page.waitForTimeout(2000);
    const ssPath = 'tmp/character_screenshot.png';
    await page.screenshot({ path: ssPath, fullPage: true });
    console.log('Screenshot saved to', ssPath);
    if (errors.length > 0) {
      console.log('Console Errors found:', errors);
    } else {
      console.log('No console errors found.');
    }
  } catch (err) {
    console.error('Error navigating:', err);
  }
  await browser.close();
})();
