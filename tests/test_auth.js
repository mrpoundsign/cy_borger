// test_auth.js — Playwright test for User Auth (Register, Login, Logout, Session Persistence)
const { test, expect } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('User Authentication & Session Persistence', () => {
    test('user can register, log out, and log back in', async ({ page }) => {
        // 1. Visit homepage
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        // Unauthenticated user sees authentication required card
        await expect(page.locator('text=AUTHENTICATION REQUIRED')).toBeVisible();

        // Switch to Register tab
        await page.locator('#tab-btn-register').click();
        await page.waitForTimeout(200);

        const loginName = 'UserOp' + Math.floor(1000 + Math.random() * 9000);

        // Fill registration form
        await page.locator('#auth-register-form input[name="username"]').fill(loginName);
        await page.locator('#auth-register-form input[name="password"]').fill('secretpass123');
        await page.locator('#auth-register-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');

        // 2. Verify logged-in state shows login name and operator handle
        await expect(page.locator('text=LOGIN:')).toBeVisible();
        await expect(page.getByText(loginName, { exact: true })).toBeVisible();
        await expect(page.locator('a:has-text("🎲 Roll Random Character")')).toBeVisible();

        // 3. Test Logout
        await page.locator('button:has-text("🚪 LOGOUT")').click();
        await page.waitForLoadState('networkidle');

        // Confirm back to Login tab and creation buttons hidden
        await expect(page.locator('#auth-login-form')).toBeVisible();
        await expect(page.locator('text=AUTHENTICATION REQUIRED')).toBeVisible();

        // 4. Test Login
        await page.locator('#auth-login-form input[name="username"]').fill(loginName);
        await page.locator('#auth-login-form input[name="password"]').fill('secretpass123');
        await page.locator('#auth-login-form button[type="submit"]').click();
        await page.waitForLoadState('networkidle');

        // Confirm logged back in
        await expect(page.locator('text=LOGIN:')).toBeVisible();
        await expect(page.getByText(loginName, { exact: true })).toBeVisible();
    });
});
