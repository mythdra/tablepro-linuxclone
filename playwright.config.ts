import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for TablePro E2E tests.
 * 
 * Run with:
 *   npx playwright test                    # Run all tests
 *   npx playwright test query-execution    # Run specific test file
 *   npx playwright test --headed           # Run with visible browser
 *   npx playwright test --debug            # Debug mode
 * 
 * Environment variables:
 *   E2E_BASE_URL: Base URL of the application (default: http://localhost:34115)
 *   E2E_SLOW_MO: Slow motion delay in ms (default: 0)
 */

export default defineConfig({
  testDir: './tests/e2e',
  
  // Timeout for each test
  timeout: 60 * 1000, // 60 seconds
  
  // Timeout for each expectation
  expect: {
    timeout: 10 * 1000, // 10 seconds
  },
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // Retry on CI only
  retries: process.env.CI ? 2 : 0,
  
  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,
  
  // Reporter configuration
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
    ...(process.env.CI ? [['github'] as const] : []),
  ],
  
  // Shared settings for all the projects below
  use: {
    // Base URL for navigation
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:34115',
    
    // Collect trace when retrying the failed test
    trace: 'on-first-retry',
    
    // Screenshot on failure
    screenshot: 'only-on-failure',
    
    // Video on failure
    video: 'retain-on-failure',
    
    // Browser options
    viewport: { width: 1920, height: 1080 },
  },
  
  // Projects configuration
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    
    // Test against mobile viewports
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
    
    // Test against branded browsers
    {
      name: 'Microsoft Edge',
      use: { ...devices['Desktop Edge'], channel: 'msedge' },
    },
    {
      name: 'Google Chrome',
      use: { ...devices['Desktop Chrome'], channel: 'chrome' },
    },
  ],
  
  // Run your local dev server before starting the tests
  // webServer: {
  //   command: 'npm run dev',
  //   url: 'http://localhost:34115',
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 120 * 1000,
  // },
});
