import { test, expect } from "@playwright/test";
import { appLayout, pressKey, waitForApp } from "./app";

// The views (DESIGN §11.1). The desktop project's window is landscape, so
// auto shows the cabinet's upper part, with the picture inside it.

test("a landscape window shows the cabinet's upper part, and its art loads", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);

  await expect.poll(async () => (await appLayout(page))?.art, { intervals: [300] }).toBe(true);
  const lay = (await appLayout(page))!;
  expect(lay.view).toBe("cabinet");
  expect(lay.part).toBe("upper");
  // The picture is inside the cabinet, which runs on below the window.
  expect(lay.tube.x).toBeGreaterThan(lay.cabinet.x);
  expect(lay.tube.y).toBeGreaterThan(lay.cabinet.y);
  expect(lay.cabinet.y + lay.cabinet.h).toBeGreaterThan(page.viewportSize()!.height);
});

test("?view=screen shows the picture alone", async ({ page }) => {
  await page.goto("/?view=screen");
  await waitForApp(page);

  await expect.poll(async () => (await appLayout(page))?.view, { intervals: [300] }).toBe("screen");
  expect((await appLayout(page))!.tube.h).toBe(page.viewportSize()!.height);
});

test("V switches the view, and the choice is remembered", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);
  await expect.poll(async () => (await appLayout(page))?.view, { intervals: [300] }).toBe("cabinet");

  await page.mouse.click(10, 10); // focus the page
  await pressKey(page, "v");
  await expect.poll(async () => (await appLayout(page))?.view, { intervals: [300] }).toBe("screen");

  await page.reload();
  await waitForApp(page);
  await expect.poll(async () => (await appLayout(page))?.chosen, { intervals: [300] }).toBe("screen");
  expect((await appLayout(page))!.view).toBe("screen");

  await page.mouse.click(10, 10);
  await pressKey(page, "v");
  await expect.poll(async () => (await appLayout(page))?.view, { intervals: [300] }).toBe("cabinet");
});
