// @ts-check
const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
    testDir: '.',
    testMatch: 'test_*.js',
    workers: 1,
    timeout: 30000,
    use: {
        baseURL: process.env.BASE_URL || 'http://localhost:8080',
        headless: true,
    },
    reporter: 'list',
});
