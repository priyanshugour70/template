# app/(public)/

Public marketing / storefront pages. Accessible without a session.

If a public page needs an optional user context (e.g. "Logged in as X" header), use `useAuth()` — it returns `loading=false, user=null` for unauthenticated visitors.
