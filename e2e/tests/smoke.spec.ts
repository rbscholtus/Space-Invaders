import { test, expect } from "@playwright/test";
import { appState, pressKey, waitForApp } from "./app";

// No ROMs in CI: these tests cover the loader and the ROM screen (DESIGN
// §15.5). Software rendering is slow, so they poll the app's published
// state instead of waiting fixed times.

test("the app starts and asks for ROMs", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);

  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");
});

test("files that aren't the ROMs are explained", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");

  await page.mouse.click(10, 10); // focus the page
  const chooser = page.waitForEvent("filechooser");
  await pressKey(page, "o");
  await (await chooser).setFiles({ name: "holiday.jpg", mimeType: "image/jpeg", buffer: Buffer.from("not a rom") });

  await expect
    .poll(async () => (await appState(page)).message, { intervals: [300] })
    .toBe("These aren't the invaders ROMs. Found: holiday.jpg");
  expect((await appState(page)).mode).toBe("roms");
});
