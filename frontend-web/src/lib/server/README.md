# src/lib/server/

Server-only helpers. Every file in this folder must start with:

```ts
import "server-only";
```

This ensures the build fails if a client component accidentally imports a server-only module (and leaks secrets to the browser).

Put here: server env readers, DB clients (if any), service-account credentials, anything sensitive.
