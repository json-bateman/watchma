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
    // Timeout for each action (click, fill, etc.)
    actionTimeout: 10 * 1000,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    // For maximum speed
    // trace: "off",
    // screenshot: "off",
    // video: "off",

    // launchOptions: {
    //   args: [
    //     "--disable-gpu",
    //     "--disable-dev-shm-usage",
    //     "--no-sandbox",
    //   ],
    // },
  },

  // Configure projects for different browsers (if needed)
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    // Uncomment to test in other browsers:
    // {
    //   name: "firefox",
    //   use: { ...devices["Desktop Firefox"] },
    // },
    // {
    //   name: "webkit",
    //   use: { ...devices["Desktop Safari"] },
    // },
  ],

  // Run local dev server before tests
  webServer: {
    command:
      "templ generate && tailwindcss -i web/input.css -o public/styles.css && go run cmd/main.go",
    url: "http://localhost:58008",
    timeout: 10 * 1000, // 10 seconds to start server
    env: {
      PORT: "58008",
      LOG_LEVEL: "WARN",
      IS_DEV: "true",
      OPENAI_API_KEY: "",
    },
  },
});
