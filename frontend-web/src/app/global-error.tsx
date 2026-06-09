"use client";

/** Last-resort error boundary. Catches errors thrown from the root layout
 * itself — e.g. an error before any provider mounts. Has to render its own
 * <html>/<body> because there's no layout above it. */
import { useEffect } from "react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Global error:", error);
  }, [error]);

  return (
    <html lang="en">
      <body
        style={{
          margin: 0,
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontFamily:
            'ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
          background: "#0a0a0a",
          color: "#fafafa",
        }}
      >
        <div style={{ maxWidth: 480, padding: "0 1rem", textAlign: "center" }}>
          <p
            style={{
              fontSize: 11,
              fontWeight: 500,
              letterSpacing: "0.08em",
              textTransform: "uppercase",
              color: "rgba(250,250,250,0.6)",
              margin: 0,
            }}
          >
            Critical error
          </p>
          <h1 style={{ marginTop: 8, fontSize: 28, fontWeight: 600 }}>The app couldn&apos;t load</h1>
          <p
            style={{
              marginTop: 12,
              fontSize: 14,
              color: "rgba(250,250,250,0.7)",
              lineHeight: 1.55,
            }}
          >
            Something went wrong before the page could render. Try reloading. If the problem
            persists, contact support.
          </p>
          {error.digest && (
            <p
              style={{
                marginTop: 16,
                display: "inline-block",
                padding: "4px 8px",
                fontSize: 11,
                fontFamily:
                  'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Monaco, Consolas, monospace',
                background: "rgba(250,250,250,0.06)",
                border: "1px solid rgba(250,250,250,0.1)",
                borderRadius: 4,
                color: "rgba(250,250,250,0.6)",
              }}
            >
              ref: {error.digest}
            </p>
          )}
          <div style={{ marginTop: 24 }}>
            <button
              type="button"
              onClick={() => reset()}
              style={{
                padding: "8px 16px",
                background: "#fafafa",
                color: "#0a0a0a",
                border: "none",
                borderRadius: 6,
                fontSize: 14,
                fontWeight: 500,
                cursor: "pointer",
              }}
            >
              Try again
            </button>
          </div>
        </div>
      </body>
    </html>
  );
}
