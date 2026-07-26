const { chromium } = require('@playwright/test');
const { execSync } = require('child_process');

(async () => {
  let winIp = 'localhost';
  try {
    winIp = execSync("ip route show | grep -i default | awk '{ print $3}'").toString().trim();
  } catch(e) {}
  
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  const url = `http://${winIp}:8080/character/8ba9e6b971492cf0`;
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
    const ssPath = '/home/mrp/.gemini/antigravity-ide/brain/ec6fed3a-b9f4-4a3e-8c52-92b6ada6ba80/character_screenshot.png';
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
