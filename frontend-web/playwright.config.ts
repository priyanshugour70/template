import { defineConfig, devices } from "@playwright/test";

// Playwright config — runs e2e/ against a dev server on http://acme.lvh.me:3000.
// The dev server is started externally (run `pnpm dev` first). For CI we
// would lift `webServer` to spawn it automatically; for now keeping it
// external avoids racing with hot-reload.
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: [["list"]],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://acme.lvh.me:3000",
    trace: "retain-on-failure",
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
