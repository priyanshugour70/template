# src/types/

TypeScript types and Zod schemas, grouped by feature. Use a folder per feature even when it currently holds a single file — keeps growth painless.

```
types/
├── auth/index.ts             # User, Session, LoginRequest, …
├── sample/index.ts           # Sample, SampleListResponse, SampleStatus
└── …
```

If a type is used by ≥2 features, surface it through a `shared/` folder; don't import across feature boundaries.
