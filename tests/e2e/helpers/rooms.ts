import { expect, Page } from "@playwright/test";

/**
 * Helper functions for room management
 */

/**
 * Create a new room
 */
export async function createRoom(
  page: Page,
  roomName: string,
  draftNumber: number = 3,
  maxPlayers: number = 5,
) {
  await page.goto("/host");
  await page.fill('input[name="roomName"]', roomName);
  await page.selectOption('select[name="draftNumber"]', draftNumber.toString());
  await page.selectOption('select[name="maxplayers"]', maxPlayers.toString());
  await page.click('button[type="submit"]');

  // Wait for redirect to room page
  await page.waitForURL(`/room/${roomName}`);
}

/**
 * Join an existing room by name
 */
export async function joinRoom(page: Page, roomName: string) {
  await page.goto("/join");

  await page.waitForSelector("text=Join", { timeout: 5000 });

  await page.click("text=Join");
  await page.waitForURL(`/room/${roomName}`);
}

/**
 * Click the ready button in the lobby
 */
export async function clickReady(page: Page) {
  await page.click('button:has-text("Ready")');

  // Wait for button state to change (might become "Unready")
  await page.waitForTimeout(200);
}

/**
 * Start the game (host only)
 */
export async function startGame(page: Page) {
  await page.click('button:has-text("Start")');

  await page.waitForSelector(`text=/draft/i`, { timeout: 5000 });
}

/**
 * Wait for other players to appear in the lobby
 */
export async function waitForPlayer(
  page: Page,
  username: string,
  timeout: number = 5000,
) {
  await expect(page.locator(`text=${username}`)).toBeVisible({ timeout });
}

/**
 * Generate unique room name for testing
 */
export function generateRoomName(prefix: string = "TestRoom"): string {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 1000);
  return `${prefix}_${timestamp}_${random}`;
}

/**
 * Leave current room
 */
export async function leaveRoom(page: Page) {
  // Navigating out of the room cleans up the player and unsubscribes from room
  await page.goBack();
}
