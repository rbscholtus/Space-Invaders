import { test, expect } from "@playwright/test";
import { appState, openFileChooser, waitForApp } from "./app";

// No ROMs in CI: these tests cover the loader and the ROM screen (DESIGN
// §15.5). Software rendering is slow, so they poll the app's published
// state instead of waiting fixed times.

test("the app starts and asks for ROMs, on the CRT monitor", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);

  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");
  // The shaders compiled: the app falls back to clean only if they don't.
  expect((await appState(page)).shader).toBe("crt");
});

test("?shader=clean draws the plain picture", async ({ page }) => {
  await page.goto("/?shader=clean");
  await waitForApp(page);

  await expect.poll(async () => (await appState(page)).shader, { intervals: [300] }).toBe("clean");
});

test("files that aren't the ROMs are explained", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);
  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");

  await page.mouse.click(10, 10); // focus the page

  // Choosing nothing changes nothing, and O still opens the chooser.
  await (await openFileChooser(page)).setFiles([]);
  await (await openFileChooser(page)).setFiles({ name: "holiday.jpg", mimeType: "image/jpeg", buffer: Buffer.from("not a rom") });

  await expect
    .poll(async () => (await appState(page)).message, { intervals: [300] })
    .toBe("These aren't the invaders ROMs. Found: holiday.jpg");
  expect((await appState(page)).mode).toBe("roms");
});
