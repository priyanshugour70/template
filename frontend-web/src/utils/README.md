# src/utils/

Pure helpers without a stronger home: formatters (`fmtCurrency`, `fmtDate`), tiny data transforms, regex tests.

A util should:

- Be a pure function.
- Not depend on React or Next.js.
- Have a unit test.

For string-class helpers tied to Tailwind (e.g. `cn`), use `src/lib/cn.ts` instead — `lib/` is for cross-cutting infrastructure, `utils/` is for ad-hoc helpers.
