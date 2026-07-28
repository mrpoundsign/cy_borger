const { test, expect } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('Auth Failure Handling', () => {
    test('failed login does not load nested page inside login form', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        // Fill invalid login details
        await page.locator('#auth-login-form input[name="username"]').fill('invalid_user_12345');
        await page.locator('#auth-login-form input[name="password"]').fill('wrong_password');
        await page.locator('#btn-submit-login-index').click();

        await page.waitForTimeout(500);

        // Check if nested form or nested h1 exists inside auth-login-form
        const nestedForm = page.locator('#auth-login-form #auth-login-form');
        await expect(nestedForm).not.toBeVisible();

        // Check that error message is displayed
        await expect(page.locator('text=Invalid username or password')).toBeVisible();
    });

    test('failed registration with spaces in username shows error', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        await page.locator('#tab-btn-register').click();
        await page.waitForTimeout(200);

        await page.locator('#auth-register-form input[name="username"]').fill('user with spaces');
        await page.locator('#auth-register-form input[name="password"]').fill('password123');
        await page.locator('#btn-submit-register-index').click();

        await page.waitForTimeout(500);

        const nestedForm = page.locator('#auth-register-form #auth-register-form');
        await expect(nestedForm).not.toBeVisible();
        await expect(page.locator('text=Username cannot contain spaces')).toBeVisible();
    });

    test('registering duplicate user shows error without nesting forms', async ({ page }) => {
        await page.goto(BASE_URL + '/');
        await page.waitForLoadState('networkidle');

        const dupUser = 'DupUser' + Math.floor(1000 + Math.random() * 9000);

        // Register first time
        await page.locator('#tab-btn-register').click();
        await page.locator('#auth-register-form input[name="username"]').fill(dupUser);
        await page.locator('#auth-register-form input[name="password"]').fill('password123');
        await page.locator('#btn-submit-register-index').click();
        await page.waitForLoadState('networkidle');

        // Logout
        await page.locator('button:has-text("🚪 LOGOUT")').click();
        await page.waitForLoadState('networkidle');

        // Try registering same username again
        await page.locator('#tab-btn-register').click();
        await page.locator('#auth-register-form input[name="username"]').fill(dupUser);
        await page.locator('#auth-register-form input[name="password"]').fill('password123');
        await page.locator('#btn-submit-register-index').click();
        await page.waitForTimeout(500);

        const nestedForm = page.locator('#auth-register-form #auth-register-form');
        await expect(nestedForm).not.toBeVisible();
        await expect(page.locator('text=Username already registered')).toBeVisible();
    });
});
