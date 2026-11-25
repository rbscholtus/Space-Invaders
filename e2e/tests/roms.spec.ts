import { existsSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { test, expect } from "@playwright/test";
import { appState, openFileChooser, pressKey, waitForApp } from "./app";

// With your own ROM set (local only: ROMs are never in the repository or
// CI): ../roms/invaders.zip, or INVADERS_ROMS.
// The HI-SCORE's key in localStorage (platform.Prefs: "space-invaders." + key).
const hiScoreKey = "space-invaders.hiscore";

const roms = process.env.INVADERS_ROMS ?? fileURLToPath(new URL("../../roms/invaders.zip", import.meta.url));

test("your ROMs start the game, are remembered, and can be forgotten", async ({ page }) => {
  test.skip(!existsSync(roms), `no ROM set at ${roms}`);

  await page.goto("/");
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");

  await page.mouse.click(10, 10);
  await (await openFileChooser(page)).setFiles(roms);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("game");

  // Remembered in IndexedDB: a reload starts the game at once, with the
  // HI-SCORE remembered in localStorage.
  await page.evaluate((key) => localStorage.setItem(key, "1230"), hiScoreKey);
  await page.reload();
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("game");
  expect((await appState(page)).hiscore).toBe(1230);

  await page.mouse.click(10, 10);
  await pressKey(page, "f");
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");
  expect(await page.evaluate((key) => localStorage.getItem(key), hiScoreKey)).toBeNull();

  // Forgotten: a reload asks again.
  await page.reload();
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");
});
