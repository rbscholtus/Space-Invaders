import { FileChooser, Page, expect } from "@playwright/test";

// The state the app publishes on the page (internal/app/game.go
// publishState).
export type AppState = { mode: "loading" | "roms" | "game"; message: string; hiscore: number };

export async function appState(page: Page): Promise<AppState> {
  const json = await page.evaluate(() => (window as any).invadersState);
  return json ? JSON.parse(json) : { mode: "loading", message: "", hiscore: 0 };
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

// Presses O until the app opens the file chooser. Ebitengine reads the
// keyboard once per frame, and with software rendering in CI a frame can
// be late, so the key is held until the chooser opens, and pressed again if
// it hasn't opened after 5 s (about how long a key press lets a page open a
// chooser).
export async function openFileChooser(page: Page): Promise<FileChooser> {
  for (let attempt = 1; ; attempt++) {
    const next = page.waitForEvent("filechooser", { timeout: 5_000 });
    await page.keyboard.down("o");
    try {
      return await next;
    } catch (e) {
      if (attempt === 5) throw e;
    } finally {
      await page.keyboard.up("o");
    }
  }
}
