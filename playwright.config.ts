import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright configuration for Watchma E2E tests
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: "./tests/e2e",

  // Maximum time one test can run
  timeout: 60 * 1000,

  // Test execution settings
  fullyParallel: false, // Run tests sequentially to avoid database conflicts
  forbidOnly: !!process.env.CI, // Fail CI if test.only is used
  retries: process.env.CI ? 2 : 0, // Retry failed tests in CI
  workers: 1, // Run one test at a time (SQLite limitation)

  // Reporter settings
  reporter: [
    ["html"], // HTML report
    ["list"], // Console output
    ["json", { outputFile: "test-results/results.json" }],
  ],

  // Shared settings for all projects
  use: {
    // Base URL for all navigation
    baseURL: "http://localhost:58008",

    // Collect trace on failure for debugging
    trace: "on-first-retry",

    // Screenshot on failure
    screenshot: "only-on-failure",

    // Video on failure
    video: "retain-on-failure",

    // Timeout for each action (click, fill, etc.)
    actionTimeout: 10 * 1000,
  },

  // Configure projects for different browsers (if needed)
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    // Uncomment to test in other browsers:
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Run local dev server before tests (optional)
  webServer: {
    command: "go run cmd/main.go",
    url: "http://localhost:58008",
    timeout: 120 * 1000, // 2 minutes to start server
    reuseExistingServer: !process.env.CI, // Use existing server in dev
    env: {
      PORT: "58008",
      LOG_LEVEL: "WARN", // Reduce noise during tests
      IS_DEV: "true",
    },
  },
});
