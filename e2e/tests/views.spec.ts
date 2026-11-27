import { test, expect } from "@playwright/test";
import { appLayout, pressKey, waitForApp } from "./app";

// The views (DESIGN §11.1). Auto shows the cabinet's frame (from its top to
// the control panel), with the picture inside it.

test("the cabinet is shown with the picture in it, and its art loads", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);

  await expect.poll(async () => (await appLayout(page))?.art, { intervals: [300] }).toBe(true);
  const lay = (await appLayout(page))!;
  expect(lay.view).toBe("cabinet");
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

// As the window narrows (a tablet, a phone, both upright), the frame's top
// stays at the top and more of the cabinet shows below; once all of it
// shows, it's centred with equal room above and below.
test("narrow windows show more of the cabinet, then centre it", async ({ page }) => {
  await page.goto("/");
  await waitForApp(page);
  await expect.poll(async () => (await appLayout(page))?.view, { intervals: [300] }).toBe("cabinet");

  const settle = async (width: number, height: number) => {
    await page.setViewportSize({ width, height });
    await expect
      .poll(async () => {
        const lay = await appLayout(page);
        return lay && Math.round(lay.frame.x * 2 + lay.frame.w) === width ? lay : undefined;
      }, { intervals: [300] })
      .toBeDefined();
    return (await appLayout(page))!;
  };

  // A tablet held upright: the frame fills the width, its top at the top,
  // the cabinet's lower part below it, running past the window.
  let lay = await settle(820, 1180);
  expect(lay.view).toBe("cabinet");
  expect(Math.abs(lay.frame.y)).toBeLessThanOrEqual(1);
  expect(lay.cabinet.y + lay.cabinet.h).toBeGreaterThan(1180);

  // A phone held upright: all of the cabinet shows, centred.
  lay = await settle(390, 844);
  expect(lay.view).toBe("cabinet");
  const above = lay.frame.y;
  const below = 844 - (lay.cabinet.y + lay.cabinet.h);
  expect(above).toBeGreaterThan(0);
  expect(Math.abs(above - below)).toBeLessThanOrEqual(1);
});
