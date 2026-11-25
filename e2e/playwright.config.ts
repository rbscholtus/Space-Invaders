import { defineConfig, devices } from "@playwright/test";

// Serves ../web/dist (built by ../scripts/build.sh) and runs the browser
// tests in Chromium with software WebGL, as CI has no GPU (DESIGN §15.5).
// Adapted from Chippie's.
export default defineConfig({
  testDir: "./tests",
  timeout: 180_000,
  expect: { timeout: 30_000 },
  fullyParallel: false,
  workers: 1,
  reporter: process.env.CI ? [["github"], ["list"]] : "list",
  use: {
    baseURL: "http://127.0.0.1:8124/",
    launchOptions: { args: ["--use-angle=swiftshader", "--enable-unsafe-swiftshader"] },
  },
  webServer: {
    command: "python3 -m http.server 8124 --bind 127.0.0.1 --directory ../web/dist",
    url: "http://127.0.0.1:8124/",
    reuseExistingServer: !process.env.CI,
  },
  projects: [{ name: "desktop", use: { ...devices["Desktop Chrome"], viewport: { width: 900, height: 700 } } }],
});
