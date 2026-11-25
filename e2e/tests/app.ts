import { Page, expect } from "@playwright/test";

// The state the app publishes on the page (internal/app/game.go
// publishState).
export type AppState = { mode: "loading" | "roms" | "game"; message: string; demo: string[] | null };

export async function appState(page: Page): Promise<AppState> {
  const json = await page.evaluate(() => (window as any).invadersState);
  return json ? JSON.parse(json) : { mode: "loading", message: "", demo: null };
}

// Waits until the loader has started the app.
export async function waitForApp(page: Page) {
  await page.waitForFunction(() => (window as any).invadersTimings !== undefined, null, { timeout: 90_000 });
  await expect(page.locator("#loader")).toBeHidden();
}

// Holds a key long enough for Ebitengine, which reads the keyboard once per
// frame, to see it.
export async function pressKey(page: Page, key: string) {
  await page.keyboard.down(key);
  await page.waitForTimeout(200);
  await page.keyboard.up(key);
}
