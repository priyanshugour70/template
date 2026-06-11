# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: communication/comm.spec.ts >> communication module >> login → create channel → send → see live message → typing → hook
- Location: e2e/communication/comm.spec.ts:36:7

# Error details

```
Error: expect(locator).toContainText(expected) failed

Locator: getByTestId('message-list')
Expected substring: "via-ws-1781210079994"
Received string:    "AAAcme Admin02:04 AMhello from playwright 1781210079936"
Timeout: 5000ms

Call log:
  - Expect "toContainText" with timeout 5000ms
  - waiting for getByTestId('message-list')
    14 × locator resolved to <div data-testid="message-list" class="flex-1 overflow-y-auto px-6 py-4 space-y-3">…</div>
       - unexpected value "AAAcme Admin02:04 AMhello from playwright 1781210079936"

```

```yaml
- article: AA Acme Admin 02:04 AM hello from playwright 1781210079936
```

# Test source

```ts
  1   | import { expect, test } from "@playwright/test";
  2   | 
  3   | // Phase 3 end-to-end. Drives a real browser through login → create channel
  4   | // → send a message → see the live render → typing indicator.
  5   | //
  6   | // Prereqs:
  7   | //   1. Backend running: `cd backend && go run ./cmd/api`
  8   | //   2. Frontend dev server: `cd frontend-web && pnpm dev`
  9   | //   3. Seeded admin user (migrations/postgres/seeds.sql).
  10  | //
  11  | // The tenant subdomain is acme.lvh.me which resolves to 127.0.0.1 — no
  12  | // /etc/hosts editing needed.
  13  | 
  14  | const EMAIL = process.env.E2E_EMAIL ?? "admin@acme.example";
  15  | const PASSWORD = process.env.E2E_PASSWORD ?? "Admin@123";
  16  | 
  17  | async function login(page: import("@playwright/test").Page) {
  18  |   await page.goto("/auth/login");
  19  | 
  20  |   // Step 1: email field. The form is multi-step; the first submit takes us
  21  |   // to the password step because the tenant is fixed by subdomain. We scope
  22  |   // queries to the form to avoid the dev-tools button matching "continue".
  23  |   await page.getByLabel(/email/i).fill(EMAIL);
  24  |   const form = page.locator("form").first();
  25  |   await form.getByRole("button", { name: /^continue$/i }).click();
  26  | 
  27  |   // Step 2: password.
  28  |   await page.getByLabel(/password/i).fill(PASSWORD);
  29  |   await form.getByRole("button", { name: /sign in|log in/i }).click();
  30  | 
  31  |   // Land on /dashboard.
  32  |   await page.waitForURL(/\/dashboard(\/|$)/, { timeout: 15_000 });
  33  | }
  34  | 
  35  | test.describe("communication module", () => {
  36  |   test("login → create channel → send → see live message → typing → hook", async ({
  37  |     page,
  38  |   }) => {
  39  |     test.setTimeout(60_000);
  40  | 
  41  |     await login(page);
  42  |     await page.goto("/dashboard/communication");
  43  | 
  44  |     // Sidebar present.
  45  |     await expect(page.getByTestId("comm-layout")).toBeVisible();
  46  | 
  47  |     // Create a fresh channel.
  48  |     const slug = `e2e-${Date.now().toString().slice(-6)}`;
  49  |     await page.getByTestId("new-channel-trigger").click();
  50  |     await page.getByTestId("channel-slug-input").fill(slug);
  51  |     await page.getByTestId("channel-name-input").fill(`E2E ${slug}`);
  52  |     await page.getByTestId("channel-create-submit").click();
  53  | 
  54  |     // Navigated to /dashboard/communication/<id>.
  55  |     await page.waitForURL(/\/dashboard\/communication\/[0-9a-f-]+/, {
  56  |       timeout: 15_000,
  57  |     });
  58  |     await expect(page.getByTestId("conversation-pane")).toBeVisible();
  59  |     await expect(page.getByTestId("conversation-title")).toContainText(`E2E ${slug}`);
  60  | 
  61  |     // Wait for WS to settle (hello + subscribe).
  62  |     await page.waitForTimeout(500);
  63  | 
  64  |     // Send a message and assert it appears in the live list.
  65  |     const body = `hello from playwright ${Date.now()}`;
  66  |     await page.getByTestId("composer-input").fill(body);
  67  |     await page.getByTestId("composer-send").click();
  68  | 
  69  |     await expect(page.getByTestId("message-list")).toContainText(body, {
  70  |       timeout: 5_000,
  71  |     });
  72  | 
  73  |     // Second assertion: prove the WS path is live. Use page.evaluate to
  74  |     // POST a message from inside the browser (so cookies are attached) WITHOUT
  75  |     // going through the React mutation. The only way this text reaches the
  76  |     // DOM is via the WS broadcast → useLiveMessages cache update.
  77  |     const wsBody = `via-ws-${Date.now()}`;
  78  |     const conversationUrl = page.url();
  79  |     const conversationId = conversationUrl.split("/").pop()!;
  80  |     const wsPostStatus = await page.evaluate(
  81  |       async ({ id, body }) => {
  82  |         const res = await fetch(`/api/v1/comm/conversations/${id}/messages`, {
  83  |           method: "POST",
  84  |           credentials: "include",
  85  |           headers: { "Content-Type": "application/json" },
  86  |           body: JSON.stringify({ body }),
  87  |         });
  88  |         return res.status;
  89  |       },
  90  |       { id: conversationId, body: wsBody },
  91  |     );
  92  |     expect(wsPostStatus).toBeLessThan(300);
  93  | 
> 94  |     await expect(page.getByTestId("message-list")).toContainText(wsBody, {
      |                                                    ^ Error: expect(locator).toContainText(expected) failed
  95  |       timeout: 5_000,
  96  |     });
  97  | 
  98  |     // Typing indicator: composer fires the event. Single-user means we won't
  99  |     // see our own indicator (same-user exclude rule). Just verify no crash.
  100 |     await page.getByTestId("composer-input").fill("more");
  101 |     await page.waitForTimeout(200);
  102 | 
  103 |     // Open the hooks drawer and create a hook.
  104 |     await page.getByTestId("hooks-trigger").click();
  105 |     await expect(page.getByTestId("hooks-panel")).toBeVisible();
  106 |     await page.getByTestId("hook-name-input").fill("e2e hook");
  107 |     await page.getByTestId("hook-display-input").fill("E2E Bot");
  108 |     await page.getByTestId("hook-create-submit").click();
  109 | 
  110 |     // The one-time URL banner appears.
  111 |     await expect(page.getByTestId("hook-secret-banner")).toBeVisible({ timeout: 5_000 });
  112 | 
  113 |     // And the hook shows up in the list.
  114 |     await expect(page.getByTestId("hook-row").first()).toContainText("e2e hook");
  115 |   });
  116 | });
  117 | 
```