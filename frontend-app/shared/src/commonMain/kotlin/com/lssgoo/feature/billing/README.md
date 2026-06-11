# Feature: Billing

Mobile counterpart of `frontend-web/src/app/(dashboard)/dashboard/billing/*` and
the backend module `internal/modules/billing`.

Per [billing-overhaul.md] in the project memory: the legacy `/subscription*`
routes are retired; the canonical flow is **Quotation → Invoice → Payment →
Receipt**, with a daily cycle cron on the backend.

## Screens (planned)

| Screen             | Maps to web route                              |
|--------------------|------------------------------------------------|
| `SubscriptionScreen` | `/dashboard/billing/subscription`            |
| `InvoiceListScreen`  | `/dashboard/billing/invoices`                |
| `InvoiceDetailScreen`| `/dashboard/billing/invoices/[id]`           |
| `QuotationScreen`    | `/dashboard/billing/quotations/[id]`         |
| `TransactionsScreen` | `/dashboard/billing/transactions`            |

## Endpoints

Same as the web; consult `backend/internal/modules/billing/handler.go` for the
authoritative list. Common ones:

| Method | Path                              |
|--------|-----------------------------------|
| GET    | `/billing/subscription`           |
| GET    | `/billing/invoices?page=&limit=`  |
| GET    | `/billing/invoices/{id}`          |
| POST   | `/billing/quotations`             |
| GET    | `/billing/transactions`           |

## BillingGate (402)

The backend returns HTTP **402 Payment Required** with code `BILLING_GATE` when
a feature is gated. Add an interceptor in `presentation/` that listens for this
and pushes the user to the subscription screen.
