import { existsSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { test, expect } from "@playwright/test";
import { appState, pressKey, waitForApp } from "./app";

// With your own ROM set (local only: ROMs are never in the repository or
// CI): ../roms/invaders.zip, or INVADERS_ROMS.
const roms = process.env.INVADERS_ROMS ?? fileURLToPath(new URL("../../roms/invaders.zip", import.meta.url));

test("your ROMs start the game, are remembered, and can be forgotten", async ({ page }) => {
  test.skip(!existsSync(roms), `no ROM set at ${roms}`);

  await page.goto("/");
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");

  await page.mouse.click(10, 10);
  const chooser = page.waitForEvent("filechooser");
  await pressKey(page, "o");
  await (await chooser).setFiles(roms);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("game");

  // Remembered in IndexedDB: a reload starts the game at once.
  await page.reload();
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("game");

  await page.mouse.click(10, 10);
  await pressKey(page, "f");
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");

  // Forgotten: a reload asks again.
  await page.reload();
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");
});
