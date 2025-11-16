import { expect, test } from "@playwright/test";
import { generateUsername, login, logout, signup } from "./helpers/auth";

/**
 * Authentication flow tests
 */
test.describe("Authentication", () => {
  test("user can sign up with valid credentials", async ({ page }) => {
    const username = generateUsername("newuser");
    const password = "Securepass123";

    await signup(page, username, password);

    await expect(page).toHaveURL((url) =>
      url.pathname === "/" || url.pathname.startsWith("/rooms")
    );
  });

  test("user can login with existing credentials", async ({ page }) => {
    const username = generateUsername("existinguser");
    const password = "Testpass123";

    await signup(page, username, password);

    await logout(page);

    await login(page, username, password);

    await expect(page).toHaveURL((url) =>
      url.pathname === "/" || url.pathname.startsWith("/rooms")
    );
  });

  test("user cannot login with wrong password", async ({ page }) => {
    const username = generateUsername("testuser");
    const correctPassword = "Correctpass123";
    const wrongPassword = "Wrongpass123";

    await signup(page, username, correctPassword);
    await logout(page);

    await page.goto("/login");
    await page.fill('input[name="username"]', username);
    await page.fill('input[name="password"]', wrongPassword);
    await page.click('button[type="submit"]');

    await page.waitForTimeout(1000);

    const isOnLoginPage = page.url().includes("/login");
    const hasErrorMessage = await page.locator(
      "text=/invalid|incorrect|wrong/i",
    ).isVisible();

    expect(isOnLoginPage || hasErrorMessage).toBe(true);
  });

  test("unauthenticated user is redirected to login", async ({ page }) => {
    await page.goto("/host");

    await expect(page).toHaveURL(/\/(login|signup)/);
  });

  test("user can logout", async ({ page }) => {
    const username = generateUsername("logouttest");
    const password = "Testpass123";

    await signup(page, username, password);

    await logout(page);

    await expect(page).toHaveURL("/login");

    // Try to access protected route - should redirect
    await page.goto("/join");
    await expect(page).toHaveURL(/\/(login|signup)/);
  });
});
