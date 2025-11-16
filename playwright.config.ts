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
  fullyParallel: false,
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

  // Run local dev server before tests
  webServer: {
    command: "go run cmd/main.go",
    url: "http://localhost:58008",
    timeout: 30 * 1000, // 30 seconds to start server
    env: {
      PORT: "58008",
      LOG_LEVEL: "WARN",
      IS_DEV: "true",
    },
  },
});
