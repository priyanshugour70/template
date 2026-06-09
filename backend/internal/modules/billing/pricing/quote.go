package pricing

import (
	"fmt"
	"sort"
)

// Feature is the pricing-layer view of a catalog feature — only the fields the
// math actually needs. Keeping this struct local to the pricing package keeps
// it decoupled from gorm + DB columns.
type Feature struct {
	Key               string
	Name              string
	Description       string
	Category          string
	BasePriceCents    int64
	PerUserPriceCents int64
	IncludedUsers     int
	IsCore            bool
	Requires          []string
	SortOrder         int
}

// QuoteInput captures the user's selection. SelectedKeys excludes core features
// (which are added automatically); it CAN include "extra_users" but doesn't
// need to — we auto-add it when userCount exceeds the included quota.
type QuoteInput struct {
	SelectedKeys  []string
	UserCount     int
	CustomerState string
	HSNSAC        string
	Currency      string
	Rates         Rates
}

// Line is a single row on the quotation / invoice. Per-line tax breakdown is
// included so invoices show GST per item, which is the GST-compliant format.
type Line struct {
	FeatureKey         string       `json:"featureKey"`
	Description        string       `json:"description"`
	HSNSAC             string       `json:"hsnSac"`
	Quantity           int          `json:"quantity"`
	UnitPriceCents     int64        `json:"unitPriceCents"`
	TaxableAmountCents int64        `json:"taxableAmountCents"`
	Tax                TaxBreakdown `json:"tax"`
	TotalCents         int64        `json:"totalCents"`
	SortOrder          int          `json:"-"`
}

// Quote is the resolved output: line items + aggregated totals + tax split.
type Quote struct {
	Lines         []Line `json:"lines"`
	SubtotalCents int64  `json:"subtotalCents"`
	CGSTCents     int64  `json:"cgstCents"`
	SGSTCents     int64  `json:"sgstCents"`
	IGSTCents     int64  `json:"igstCents"`
	TotalCents    int64  `json:"totalCents"`
	PlaceOfSupply string `json:"placeOfSupply"`
	Currency      string `json:"currency"`
	UserCount     int    `json:"userCount"`
	IncludedUsers int    `json:"includedUsers"`
	ExtraUsers    int    `json:"extraUsers"`
}

// BuildQuote resolves a feature selection into a concrete priced quotation.
//
// Algorithm:
//  1. Catalog must contain every selected key. Unknown keys → error.
//  2. Every core feature is implicitly included (always charged ₹0, kept off
//     the line items unless it had a base price for some reason).
//  3. `requires` chains are walked transitively (e.g. audit_export → audit_log).
//  4. Total includedUsers = sum of included_users across the selected features.
//  5. If userCount > includedUsers and `extra_users` is in the catalog, it's
//     auto-added with quantity = (userCount - includedUsers).
//  6. For each selected feature, in sort_order, emit a line:
//       - features with base_price_cents > 0 → qty=1, unit=base
//       - "extra_users" → qty=extra, unit=per_user_price_cents
//       - other per-user features (per_user_price_cents > 0, key != extra_users)
//         → qty=userCount, unit=per_user_price_cents
//  7. Per-line tax via ComputeTax, totals summed.
//
// The pure-function design means tests can lock the math in without spinning
// up a DB.
func BuildQuote(catalog map[string]Feature, in QuoteInput) (Quote, error) {
	if in.UserCount < 1 {
		return Quote{}, fmt.Errorf("user count must be at least 1")
	}
	hsn := in.HSNSAC
	if hsn == "" {
		hsn = "998314"
	}
	currency := in.Currency
	if currency == "" {
		currency = "INR"
	}

	// 1+2: build the working set: explicit selection plus every core feature.
	selected := make(map[string]struct{}, len(in.SelectedKeys)+5)
	for _, key := range in.SelectedKeys {
		f, ok := catalog[key]
		if !ok {
			return Quote{}, fmt.Errorf("unknown feature %q", key)
		}
		_ = f
		selected[key] = struct{}{}
	}
	for key, f := range catalog {
		if f.IsCore {
			selected[key] = struct{}{}
		}
	}

	// 3: resolve `requires`. Walk until no new keys appear (transitive closure
	// with a small fixed-iteration cap to defeat hand-rolled cycles).
	for iter := 0; iter < 16; iter++ {
		added := false
		for key := range selected {
			for _, req := range catalog[key].Requires {
				if _, exists := selected[req]; !exists {
					if _, known := catalog[req]; !known {
						return Quote{}, fmt.Errorf("feature %q requires unknown feature %q", key, req)
					}
					selected[req] = struct{}{}
					added = true
				}
			}
		}
		if !added {
			break
		}
	}

	// 4: included users quota.
	includedUsers := 0
	for key := range selected {
		f := catalog[key]
		// extra_users contributes 0 included; everything else additive.
		if f.Key == "extra_users" {
			continue
		}
		includedUsers += f.IncludedUsers
	}

	// 5: auto-add extra_users if needed.
	extraUsers := in.UserCount - includedUsers
	if extraUsers < 0 {
		extraUsers = 0
	}
	if extraUsers > 0 {
		if _, ok := catalog["extra_users"]; ok {
			selected["extra_users"] = struct{}{}
		}
	}

	// 6: build line items in sort_order.
	orderedKeys := make([]string, 0, len(selected))
	for key := range selected {
		orderedKeys = append(orderedKeys, key)
	}
	sort.Slice(orderedKeys, func(i, j int) bool {
		return catalog[orderedKeys[i]].SortOrder < catalog[orderedKeys[j]].SortOrder
	})

	q := Quote{
		Lines:         make([]Line, 0, len(orderedKeys)),
		PlaceOfSupply: in.CustomerState,
		Currency:      currency,
		UserCount:     in.UserCount,
		IncludedUsers: includedUsers,
		ExtraUsers:    extraUsers,
	}

	for _, key := range orderedKeys {
		f := catalog[key]

		// Base line: any feature with a positive flat price.
		if f.BasePriceCents > 0 {
			line := buildLine(f, 1, f.BasePriceCents, hsn, in.CustomerState, in.Rates)
			q.Lines = append(q.Lines, line)
		}

		// Per-user line: only "extra_users" emits a line, and only when the
		// quota was exceeded. Other features with per_user_price are currently
		// ignored (no such features in the seed catalog) — when they appear,
		// extend this branch.
		if f.Key == "extra_users" && extraUsers > 0 && f.PerUserPriceCents > 0 {
			line := buildLine(f, extraUsers, f.PerUserPriceCents, hsn, in.CustomerState, in.Rates)
			q.Lines = append(q.Lines, line)
		}
	}

	// 7: aggregate totals.
	for _, l := range q.Lines {
		q.SubtotalCents += l.TaxableAmountCents
		q.CGSTCents += l.Tax.CGSTCents
		q.SGSTCents += l.Tax.SGSTCents
		q.IGSTCents += l.Tax.IGSTCents
	}
	q.TotalCents = q.SubtotalCents + q.CGSTCents + q.SGSTCents + q.IGSTCents
	return q, nil
}

func buildLine(f Feature, qty int, unitCents int64, hsn, customerState string, rates Rates) Line {
	taxable := unitCents * int64(qty)
	tax := ComputeTax(taxable, customerState, rates)
	desc := f.Name
	if qty > 1 {
		desc = fmt.Sprintf("%s × %d", f.Name, qty)
	}
	return Line{
		FeatureKey:         f.Key,
		Description:        desc,
		HSNSAC:             hsn,
		Quantity:           qty,
		UnitPriceCents:     unitCents,
		TaxableAmountCents: taxable,
		Tax:                tax,
		TotalCents:         taxable + tax.TotalCents,
		SortOrder:          f.SortOrder,
	}
}
