import { Page } from "@playwright/test";

/**
 * Helper functions for authentication flows
 */

/**
 * Sign up a new user
 */
export async function signup(page: Page, username: string, password: string) {
  await page.goto("/login");
  await page.fill('input[name="username"]', username);
  await page.fill('input[name="password"]', password);
  await page.click('button[type="submit"]');

  await page.waitForURL((url) =>
    url.pathname === "/" || url.pathname.startsWith("/rooms")
  );
}

/**
 * Log in an existing user
 */
export async function login(page: Page, username: string, password: string) {
  await page.goto("/login");
  await page.fill('input[name="username"]', username);
  await page.fill('input[name="password"]', password);
  await page.click('button[type="submit"]');

  await page.waitForURL((url) =>
    url.pathname === "/" || url.pathname.startsWith("/rooms")
  );
}

export async function logout(page: Page) {
  await page.goto("/");
  await page.click("#userDropdown");
  await page.click("#logout");

  await page.waitForURL("**\/login");
}

/**
 * Create a unique username for testing
 */
export function generateUsername(prefix: string = "user"): string {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 1000);
  return `${prefix}_${timestamp}_${random}`;
}

/**
 * Check if user is authenticated by seeing if they can access protected routes
 */
export async function isAuthenticated(page: Page): Promise<boolean> {
  try {
    // Wait for any authenticated-only element
    await page.waitForSelector(
      'a[href="/rooms/host"], a[href="/rooms/join"]',
      { timeout: 2000 },
    );
    return true;
  } catch {
    return false;
  }
}
