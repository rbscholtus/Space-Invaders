import { test, expect } from "@playwright/test";
import { appState, pressKey, waitForApp } from "./app";

// No ROMs in CI: these tests cover the loader, the ROM screen and the
// CPU-test demo (DESIGN §15.5). Software rendering is slow, so they poll
// the app's published state instead of waiting fixed times.

test("the app starts and asks for ROMs", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);

  await expect.poll(async () => (await appState(page)).mode, { intervals: [300] }).toBe("roms");
});

test("the CPU-test demo shows its first result", async ({ page }) => {
  await page.goto("/?demo=fast");
  await waitForApp(page);

  // The header first, then "dad <b,d,h,sp> ... OK" after about 1.1 G cycles.
  await expect.poll(async () => (await appState(page)).demo?.[0], { intervals: [500] }).toBe("8080 instruction exerciser");
  await expect
    .poll(async () => (await appState(page)).demo?.[1] ?? "", { intervals: [1000], timeout: 150_000 })
    .toMatch(/^dad <b,d,h,sp>\s+OK$/);
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
