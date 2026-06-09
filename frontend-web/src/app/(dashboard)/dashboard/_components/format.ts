// Money + date formatters for the dashboard charts. The backend ships minor
// units (cents/paise); the UI converts here. Currency defaults to INR because
// the seeded tax config is Karnataka — pass a code explicitly if you ship
// multi-currency support later.

const SYMBOLS: Record<string, string> = { INR: "₹", USD: "$", EUR: "€", GBP: "£" };

export function formatMoney(cents: number, currency = "INR"): string {
  const sym = SYMBOLS[currency.toUpperCase()] ?? `${currency} `;
  const amount = cents / 100;
  return `${sym}${amount.toLocaleString("en-IN", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })}`;
}

export function formatCompactMoney(cents: number, currency = "INR"): string {
  const sym = SYMBOLS[currency.toUpperCase()] ?? `${currency} `;
  const amount = cents / 100;
  if (Math.abs(amount) >= 10_000_000) return `${sym}${(amount / 10_000_000).toFixed(1)}Cr`;
  if (Math.abs(amount) >= 100_000) return `${sym}${(amount / 100_000).toFixed(1)}L`;
  if (Math.abs(amount) >= 1_000) return `${sym}${(amount / 1_000).toFixed(1)}k`;
  return `${sym}${amount.toFixed(0)}`;
}

export function formatNumber(n: number): string {
  if (Math.abs(n) >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (Math.abs(n) >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

export function monthLabel(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-IN", { month: "short" });
}

export function dayLabel(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-IN", { day: "2-digit", month: "short" });
}

export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  const seconds = Math.floor((Date.now() - then) / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}
