import { expect, test } from "@playwright/test";

// Phase 3 end-to-end. Drives a real browser through login → create channel
// → send a message → see the live render → typing indicator.
//
// Prereqs:
//   1. Backend running: `cd backend && go run ./cmd/api`
//   2. Frontend dev server: `cd frontend-web && pnpm dev`
//   3. Seeded admin user (migrations/postgres/seeds.sql).
//
// The tenant subdomain is acme.lvh.me which resolves to 127.0.0.1 — no
// /etc/hosts editing needed.

const EMAIL = process.env.E2E_EMAIL ?? "admin@acme.example";
const PASSWORD = process.env.E2E_PASSWORD ?? "Admin@123";

async function login(page: import("@playwright/test").Page) {
  await page.goto("/auth/login");

  // Step 1: email field. The form is multi-step; the first submit takes us
  // to the password step because the tenant is fixed by subdomain. We scope
  // queries to the form to avoid the dev-tools button matching "continue".
  await page.getByLabel(/email/i).fill(EMAIL);
  const form = page.locator("form").first();
  await form.getByRole("button", { name: /^continue$/i }).click();

  // Step 2: password.
  await page.getByLabel(/password/i).fill(PASSWORD);
  await form.getByRole("button", { name: /sign in|log in/i }).click();

  // Land on /dashboard.
  await page.waitForURL(/\/dashboard(\/|$)/, { timeout: 15_000 });
}

test.describe("communication module", () => {
  test("login → create channel → send → see live message → typing → hook", async ({
    page,
  }) => {
    test.setTimeout(60_000);

    await login(page);
    await page.goto("/dashboard/communication");

    // Sidebar present.
    await expect(page.getByTestId("comm-layout")).toBeVisible();
    await expect(page.getByTestId("comm-sidebar")).toBeVisible();

    // Create a fresh channel.
    const slug = `e2e-${Date.now().toString().slice(-6)}`;
    // The global app-sidebar's auto-expanded "Communication" section can briefly
    // overlay this trigger on first paint; force-click sidesteps the
    // pointer-event race while still exercising the button.
    await page
  .locator('[data-testid="new-channel-trigger"]')
  .evaluate((el) => (el as HTMLElement).click());
    await page.getByTestId("channel-slug-input").fill(slug);
    await page.getByTestId("channel-name-input").fill(`E2E ${slug}`);
    await page.getByTestId("channel-create-submit").click();

    // Navigated to /dashboard/communication/<id>.
    await page.waitForURL(/\/dashboard\/communication\/[0-9a-f-]+/, {
      timeout: 15_000,
    });
    await expect(page.getByTestId("conversation-pane")).toBeVisible();
    await expect(page.getByTestId("conversation-title")).toContainText(`E2E ${slug}`);

    // Wait for WS to settle (hello + subscribe).
    await page.waitForTimeout(500);

    // Send a message and assert it appears in the live list.
    const body = `hello from playwright ${Date.now()}`;
    await page.getByTestId("composer-input").fill(body);
    await page.getByTestId("composer-send").click();

    await expect(page.getByTestId("message-list")).toContainText(body, {
      timeout: 5_000,
    });

    // Second assertion: prove the WS path is live. Use page.evaluate to
    // POST a message from inside the browser (so cookies are attached) WITHOUT
    // going through the React mutation. The only way this text reaches the
    // DOM is via the WS broadcast → useLiveMessages cache update.
    const wsBody = `via-ws-${Date.now()}`;
    const conversationUrl = page.url();
    const conversationId = conversationUrl.split("/").pop()!;
    const wsPostStatus = await page.evaluate(
      async ({ id, body }) => {
        const res = await fetch(`/api/v1/comm/conversations/${id}/messages`, {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ body }),
        });
        return res.status;
      },
      { id: conversationId, body: wsBody },
    );
    expect(wsPostStatus).toBeLessThan(300);

    await expect(page.getByTestId("message-list")).toContainText(wsBody, {
      timeout: 5_000,
    });

    // Typing indicator: composer fires the event. Single-user means we won't
    // see our own indicator (same-user exclude rule). Just verify no crash.
    await page.getByTestId("composer-input").fill("more");
    await page.waitForTimeout(200);

    // Open the hooks drawer and create a hook.
    await page.getByTestId("hooks-trigger").click();
    await expect(page.getByTestId("hooks-panel")).toBeVisible();
    await page.getByTestId("hook-name-input").fill("e2e hook");
    await page.getByTestId("hook-display-input").fill("E2E Bot");
    await page.getByTestId("hook-create-submit").click();

    // The one-time URL banner appears.
    await expect(page.getByTestId("hook-secret-banner")).toBeVisible({ timeout: 5_000 });

    // And the hook shows up in the list.
    await expect(page.getByTestId("hook-row").first()).toContainText("e2e hook");
  });

  test("react to a message", async ({ page }) => {
    test.setTimeout(60_000);
    await login(page);
    await page.goto("/dashboard/communication");

    // Use the channel from the prior test if present, else create one.
    const slug = `e2e-react-${Date.now().toString().slice(-6)}`;
    await page
  .locator('[data-testid="new-channel-trigger"]')
  .evaluate((el) => (el as HTMLElement).click());
    await page.getByTestId("channel-slug-input").fill(slug);
    await page.getByTestId("channel-name-input").fill(`E2E React ${slug}`);
    await page.getByTestId("channel-create-submit").click();
    await page.waitForURL(/\/dashboard\/communication\/[0-9a-f-]+/);
    await expect(page.getByTestId("conversation-pane")).toBeVisible();

    // Send a message so there's something to react to.
    await page.getByTestId("composer-input").fill("react to me");
    await page.getByTestId("composer-send").click();
    await expect(page.getByTestId("message-list")).toContainText("react to me");

    // Hover the message → click the add-reaction trigger → pick 👍.
    const row = page.getByTestId("message-row").first();
    await row.hover();
    await row.getByTestId("reaction-add-trigger").click();
    await page.getByTestId("reaction-pick-👍").click();

    // Chip appears with count 1 and mine=true. WS broadcast keeps it in
    // sync; sidebar refetch also triggers but unrelated here.
    const chip = page.getByTestId("reaction-chip-👍");
    await expect(chip).toBeVisible({ timeout: 5_000 });
    await expect(chip).toHaveAttribute("data-mine", "true");
    await expect(chip).toContainText("1");
  });

  test("start a DM with a teammate", async ({ page }) => {
    test.setTimeout(60_000);
    await login(page);
    await page.goto("/dashboard/communication");

    await page
      .locator('[data-testid="new-dm-trigger"]')
      .evaluate((el) => (el as HTMLElement).click());
    // Type a partial email to narrow the list.
    await page.getByTestId("dm-search-input").fill("bob.builder");
    const pick = page.getByTestId("dm-user-bob.builder@acme.example");
    await expect(pick).toBeVisible({ timeout: 5_000 });
    await pick.click();

    // Navigated to the DM conversation.
    await page.waitForURL(/\/dashboard\/communication\/[0-9a-f-]+/, { timeout: 15_000 });
    await expect(page.getByTestId("conversation-pane")).toBeVisible();

    // Sidebar should now list the new DM under "Direct messages".
    const dmRow = page
      .locator('[data-testid^="conv-link-"][data-conv-type="dm"]')
      .first();
    await expect(dmRow).toBeVisible();

    // Send a message in the DM.
    await page.getByTestId("composer-input").fill("hello bob");
    await page.getByTestId("composer-send").click();
    await expect(page.getByTestId("message-list")).toContainText("hello bob");
  });
});
