/** Format an ISO timestamp using the browser's locale. */
export function fmtDate(iso: string, opts: Intl.DateTimeFormatOptions = { dateStyle: "medium" }) {
  if (!iso) return "";
  try {
    return new Intl.DateTimeFormat(undefined, opts).format(new Date(iso));
  } catch {
    return iso;
  }
}

/** Format a number as currency. Defaults to en-IN / INR — change per project. */
export function fmtCurrency(value: number, currency = "INR", locale = "en-IN") {
  if (!Number.isFinite(value)) return "";
  return new Intl.NumberFormat(locale, { style: "currency", currency }).format(value);
}

export function truncate(text: string, max = 80) {
  if (text.length <= max) return text;
  return text.slice(0, max - 1) + "…";
}
