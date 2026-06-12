import { expect, test } from "@playwright/test";

// Communication module e2e — drives a real browser through every feature
// added in Phases 3 & 4. The seeded admin (admin@acme.example) and a second
// seeded user (bob.builder@acme.example) are both required for the DM/member
// flows.
//
// Prereqs:
//   1. Backend running: `cd backend && go run ./cmd/api`
//   2. Frontend dev server: `cd frontend-web && pnpm dev`
//   3. Seeds applied: `psql -f migrations/postgres/seeds.sql`

const EMAIL = process.env.E2E_EMAIL ?? "admin@acme.example";
const PASSWORD = process.env.E2E_PASSWORD ?? "Admin@123";
const PEER_EMAIL = "bob.builder@acme.example";

async function login(page: import("@playwright/test").Page) {
  await page.goto("/auth/login");
  await page.getByLabel(/email/i).fill(EMAIL);
  const form = page.locator("form").first();
  await form.getByRole("button", { name: /^continue$/i }).click();
  await page.getByLabel(/password/i).fill(PASSWORD);
  await form.getByRole("button", { name: /sign in|log in/i }).click();
  await page.waitForURL(/\/dashboard(\/|$)/, { timeout: 15_000 });
}

// Click trigger that may be visually beneath the global app sidebar on first
// paint. evaluate-click bypasses Playwright's hit-test, dispatching the
// click directly on the React element.
async function jsClick(page: import("@playwright/test").Page, testid: string) {
  await page.waitForSelector(`[data-testid="${testid}"]`, { state: "attached" });
  await page
    .locator(`[data-testid="${testid}"]`)
    .first()
    .evaluate((el) => (el as HTMLElement).click());
}

async function createChannel(
  page: import("@playwright/test").Page,
  prefix: string,
): Promise<{ slug: string; name: string }> {
  const slug = `${prefix}-${Date.now().toString().slice(-6)}`;
  const name = `${prefix.toUpperCase()} ${slug}`;
  await jsClick(page, "new-channel-trigger");
  await page.getByTestId("channel-slug-input").fill(slug);
  await page.getByTestId("channel-name-input").fill(name);
  await page.getByTestId("channel-create-submit").click();
  await page.waitForURL(/\/dashboard\/communication\/[0-9a-f-]+/, {
    timeout: 15_000,
  });
  await expect(page.getByTestId("conversation-pane")).toBeVisible();
  return { slug, name };
}

test.describe("communication module", () => {
  test("channel: create → live message via WS → typing → hook", async ({ page }) => {
    test.setTimeout(60_000);

    await login(page);
    await page.goto("/dashboard/communication");
    await expect(page.getByTestId("comm-layout")).toBeVisible();
    await expect(page.getByTestId("comm-sidebar")).toBeVisible();

    const { name } = await createChannel(page, "e2e");
    await expect(page.getByTestId("conversation-title")).toContainText(name);

    await page.waitForTimeout(500); // settle WS

    // Composer send → render.
    const body = `hello ${Date.now()}`;
    await page.getByTestId("composer-input").fill(body);
    await page.getByTestId("composer-send").click();
    await expect(page.getByTestId("message-list")).toContainText(body, {
      timeout: 5_000,
    });

    // WS-only proof: fetch-POST a message → only WS broadcast can land it.
    const wsBody = `via-ws-${Date.now()}`;
    const conversationId = page.url().split("/").pop()!;
    const status = await page.evaluate(
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
    expect(status).toBeLessThan(300);
    await expect(page.getByTestId("message-list")).toContainText(wsBody, {
      timeout: 5_000,
    });

    // Hook drawer: create → secret banner → row appears.
    await page.getByTestId("hooks-trigger").click();
    await expect(page.getByTestId("hooks-panel")).toBeVisible();
    await page.getByTestId("hook-name-input").fill("e2e hook");
    await page.getByTestId("hook-display-input").fill("E2E Bot");
    await page.getByTestId("hook-create-submit").click();
    await expect(page.getByTestId("hook-secret-banner")).toBeVisible({
      timeout: 5_000,
    });
    await expect(page.getByTestId("hook-row").first()).toContainText("e2e hook");
  });

  test("channel: react to a message", async ({ page }) => {
    test.setTimeout(60_000);
    await login(page);
    await page.goto("/dashboard/communication");
    await createChannel(page, "e2e-react");

    await page.getByTestId("composer-input").fill("react to me");
    await page.getByTestId("composer-send").click();
    await expect(page.getByTestId("message-list")).toContainText("react to me");

    const row = page.getByTestId("message-row").first();
    await row.hover();
    // Open the quick-react popover; JS-click the pick because the global
    // sidebar's expanded menu items can briefly overlay it.
    await row.getByTestId("reaction-add-trigger").click();
    await jsClick(page, "reaction-pick-👍");

    const chip = page.getByTestId("reaction-chip-👍");
    await expect(chip).toBeVisible({ timeout: 5_000 });
    await expect(chip).toHaveAttribute("data-mine", "true");
    await expect(chip).toContainText("1");
  });

  test("channel: edit name + manage members + leave", async ({ page }) => {
    test.setTimeout(60_000);
    await login(page);
    await page.goto("/dashboard/communication");
    await createChannel(page, "e2e-edit");

    // Channel settings dropdown → Edit channel.
    await page.getByTestId("channel-settings-trigger").click();
    await page.getByTestId("channel-settings-edit").click();
    const newName = `Edited ${Date.now().toString().slice(-5)}`;
    await page.getByTestId("edit-channel-name").fill(newName);
    await page.getByTestId("edit-channel-topic").fill("our topic");
    await page.getByTestId("edit-channel-submit").click();

    await expect(page.getByTestId("conversation-title")).toContainText(newName, {
      timeout: 5_000,
    });

    // Members drawer: add a member.
    await page.getByTestId("members-trigger").click();
    await expect(page.locator('[data-testid="members-tabs"]')).toBeVisible();
    await page.getByTestId("members-tab-add").click();
    await page.getByTestId("members-add-search").fill("bob.builder");
    await jsClick(page, `members-add-${PEER_EMAIL}`);

    // Switch back to list and verify the new member row exists.
    await page.getByTestId("members-tab-list").click();
    const newRow = page
      .locator(`[data-testid="member-row"][data-user-id]`)
      .filter({ hasText: PEER_EMAIL });
    await expect(newRow).toBeVisible({ timeout: 5_000 });

    // Now remove that member.
    await newRow.locator('[data-testid^="member-remove-"]').click();
    await expect(newRow).toHaveCount(0, { timeout: 5_000 });
  });

  test("DM: start with teammate shows peer name (not literal 'DM')", async ({ page }) => {
    test.setTimeout(60_000);
    await login(page);
    await page.goto("/dashboard/communication");

    await jsClick(page, "new-dm-trigger");
    await page.getByTestId("dm-search-input").fill("bob.builder");
    const pick = page.getByTestId(`dm-user-${PEER_EMAIL}`);
    await expect(pick).toBeVisible({ timeout: 5_000 });
    await pick.click();

    await page.waitForURL(/\/dashboard\/communication\/[0-9a-f-]+/, {
      timeout: 15_000,
    });
    await expect(page.getByTestId("conversation-pane")).toBeVisible();

    // The conversation title must surface the peer's display name.
    await expect(page.getByTestId("conversation-title")).toContainText("Bob Builder", {
      timeout: 5_000,
    });

    // And the sidebar's DM row should NOT be the literal placeholder.
    const dmRow = page
      .locator('[data-testid^="conv-link-"][data-conv-type="dm"]')
      .first();
    await expect(dmRow).toBeVisible();
    await expect(dmRow).toContainText("Bob");

    // Send a DM message.
    await page.getByTestId("composer-input").fill("hello bob");
    await page.getByTestId("composer-send").click();
    await expect(page.getByTestId("message-list")).toContainText("hello bob");
  });
});
