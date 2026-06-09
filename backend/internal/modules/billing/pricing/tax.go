// Package pricing holds pure-function pricing math for the billing module.
// No database, no I/O — everything is deterministic and testable in isolation.
//
// India-specific GST: when the customer's state matches the company's home
// state, the tax is split into CGST + SGST (typically 9% each). Otherwise it's
// IGST (typically 18%). The split is dictated by the place-of-supply rules
// in the GST Act; this calculator implements the SaaS-relevant subset.
package pricing

import (
	"math"
	"strings"
)

// Rates carries the configured GST percentages plus the home state for the
// place-of-supply comparison. Values are percentages, not fractions
// (i.e. 9.0 means 9%, not 0.09).
type Rates struct {
	HomeState string
	CGSTPct   float64
	SGSTPct   float64
	IGSTPct   float64
}

// TaxBreakdown is the result of applying GST to a taxable amount. Either the
// CGST/SGST pair is populated OR IGST is populated — never both.
type TaxBreakdown struct {
	CGSTPct   float64 `json:"cgstPct"`
	SGSTPct   float64 `json:"sgstPct"`
	IGSTPct   float64 `json:"igstPct"`
	CGSTCents int64   `json:"cgstCents"`
	SGSTCents int64   `json:"sgstCents"`
	IGSTCents int64   `json:"igstCents"`
	TotalCents int64  `json:"totalTaxCents"`
}

// ComputeTax applies GST to a taxable amount in cents. The decision between
// intra-state (CGST+SGST) and inter-state (IGST) is driven by case-insensitive
// trimmed equality of customerState and rates.HomeState. An empty customerState
// is treated as intra-state — that's the safe default when the customer
// hasn't yet provided a billing address (they're a local lead until proven
// otherwise).
//
// Cents-based math + math.Round on the per-component result avoids the drift
// you get from multiplying floats together. We round each component (CGST,
// SGST, IGST) independently and sum the rounded values, which matches how
// most invoicing tools display the breakdown.
func ComputeTax(taxableCents int64, customerState string, rates Rates) TaxBreakdown {
	if taxableCents <= 0 {
		return TaxBreakdown{}
	}
	if sameState(customerState, rates.HomeState) {
		cgst := roundPct(taxableCents, rates.CGSTPct)
		sgst := roundPct(taxableCents, rates.SGSTPct)
		return TaxBreakdown{
			CGSTPct:    rates.CGSTPct,
			SGSTPct:    rates.SGSTPct,
			CGSTCents:  cgst,
			SGSTCents:  sgst,
			TotalCents: cgst + sgst,
		}
	}
	igst := roundPct(taxableCents, rates.IGSTPct)
	return TaxBreakdown{
		IGSTPct:    rates.IGSTPct,
		IGSTCents:  igst,
		TotalCents: igst,
	}
}

// sameState — empty customerState falls back to "same as home" so a missing
// billing address doesn't trigger the IGST path on the first invoice.
func sameState(customerState, homeState string) bool {
	c := strings.TrimSpace(strings.ToLower(customerState))
	h := strings.TrimSpace(strings.ToLower(homeState))
	if c == "" {
		return true
	}
	return c == h
}

func roundPct(cents int64, pct float64) int64 {
	if pct <= 0 {
		return 0
	}
	return int64(math.Round(float64(cents) * pct / 100.0))
}
