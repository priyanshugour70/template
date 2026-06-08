# src/actions/

Next.js Server Actions, grouped by feature. Use Server Actions for mutations that originate from React components (forms, optimistic updates) when you want the form to work without client JS.

Conventions:

- One folder per feature: `actions/sample/`, `actions/auth/`.
- Each action is a `"use server"` async function.
- Throw `Error` for unexpected failures, return a typed result for known failure modes (the form needs to render the error).
- Server Actions can call `services/*` directly since they execute on the server.
