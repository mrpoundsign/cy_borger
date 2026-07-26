// @ts-check
const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
    testDir: './tests',
    outputDir: './tmp/test-results',
    testMatch: 'test_*.js',
    workers: 1,
    timeout: 30000,
    use: {
        baseURL: process.env.BASE_URL || 'http://localhost:8080',
        headless: true,
    },
    globalSetup: require.resolve('./tests/global-setup.js'),
    reporter: 'list',
});
