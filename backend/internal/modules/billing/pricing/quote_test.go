package pricing

import (
	"strings"
	"testing"
)

// seedCatalog builds the canonical seed feature set as a map keyed by feature
// key, for tests that don't need DB access. Mirrors migration 012's seed.
func seedCatalog() map[string]Feature {
	feats := []Feature{
		{Key: "auth", Name: "Authentication", Category: "core", IsCore: true, SortOrder: 10},
		{Key: "settings", Name: "Settings hub", Category: "core", IsCore: true, SortOrder: 20},
		{Key: "user_management", Name: "User management", Category: "core", IsCore: true, SortOrder: 30},
		{Key: "basic_rbac", Name: "System roles", Category: "core", IsCore: true, SortOrder: 40},
		{Key: "notifications", Name: "In-app notifications", Category: "core", IsCore: true, SortOrder: 50},
		{Key: "starter_bundle", Name: "Starter package", Category: "limits", BasePriceCents: 1_000_000, IncludedUsers: 10, SortOrder: 100},
		{Key: "extra_users", Name: "Additional users", Category: "limits", PerUserPriceCents: 50_000, SortOrder: 110},
		{Key: "custom_roles", Name: "Custom roles", Category: "admin", BasePriceCents: 200_000, Requires: []string{"basic_rbac"}, SortOrder: 200},
		{Key: "multi_org", Name: "Multi-org", Category: "admin", BasePriceCents: 300_000, SortOrder: 210},
		{Key: "departments", Name: "Departments", Category: "admin", BasePriceCents: 250_000, SortOrder: 220},
		{Key: "groups", Name: "Groups", Category: "admin", BasePriceCents: 200_000, SortOrder: 230},
		{Key: "audit_log", Name: "Audit log", Category: "compliance", BasePriceCents: 150_000, SortOrder: 300},
		{Key: "audit_export", Name: "Audit export", Category: "compliance", BasePriceCents: 100_000, Requires: []string{"audit_log"}, SortOrder: 310},
		{Key: "webhooks", Name: "Webhooks", Category: "integrations", BasePriceCents: 250_000, SortOrder: 400},
		{Key: "api_keys", Name: "API keys", Category: "integrations", BasePriceCents: 250_000, SortOrder: 410},
	}
	m := make(map[string]Feature, len(feats))
	for _, f := range feats {
		m[f.Key] = f
	}
	return m
}

var karnatakaRates = Rates{HomeState: "Karnataka", CGSTPct: 9, SGSTPct: 9, IGSTPct: 18}

func TestBuildQuote_StarterIntraStateNoExtraUsers(t *testing.T) {
	q, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys:  []string{"starter_bundle"},
		UserCount:     5,
		CustomerState: "Karnataka",
		Rates:         karnatakaRates,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := q.SubtotalCents, int64(1_000_000); got != want {
		t.Errorf("subtotal: got %d want %d", got, want)
	}
	if got, want := q.CGSTCents, int64(90_000); got != want {
		t.Errorf("CGST: got %d want %d", got, want)
	}
	if got, want := q.SGSTCents, int64(90_000); got != want {
		t.Errorf("SGST: got %d want %d", got, want)
	}
	if q.IGSTCents != 0 {
		t.Errorf("intra-state should have IGST=0, got %d", q.IGSTCents)
	}
	if got, want := q.TotalCents, int64(1_180_000); got != want {
		t.Errorf("total: got %d want %d", got, want)
	}
	if q.ExtraUsers != 0 {
		t.Errorf("expected 0 extra users for headcount 5 with 10 included, got %d", q.ExtraUsers)
	}
	// Should have only one line (starter_bundle). Core features don't emit lines.
	if len(q.Lines) != 1 {
		t.Errorf("expected 1 line, got %d (%v)", len(q.Lines), lineKeys(q))
	}
}

func TestBuildQuote_StarterPlusExtraUsersAutoAdded(t *testing.T) {
	q, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys:  []string{"starter_bundle"},
		UserCount:     15, // 10 included + 5 extra
		CustomerState: "Karnataka",
		Rates:         karnatakaRates,
	})
	if err != nil {
		t.Fatal(err)
	}
	if q.ExtraUsers != 5 {
		t.Fatalf("expected 5 extra users, got %d", q.ExtraUsers)
	}
	// 1_000_000 + (5 * 50_000) = 1_250_000
	if got, want := q.SubtotalCents, int64(1_250_000); got != want {
		t.Errorf("subtotal: got %d want %d", got, want)
	}
	// Verify extra_users line is present with correct quantity.
	var extra *Line
	for i := range q.Lines {
		if q.Lines[i].FeatureKey == "extra_users" {
			extra = &q.Lines[i]
			break
		}
	}
	if extra == nil {
		t.Fatalf("expected extra_users line, lines=%v", lineKeys(q))
	}
	if extra.Quantity != 5 || extra.UnitPriceCents != 50_000 {
		t.Errorf("extra_users line wrong: qty=%d unit=%d", extra.Quantity, extra.UnitPriceCents)
	}
}

func TestBuildQuote_CrossStateIGST(t *testing.T) {
	q, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys:  []string{"starter_bundle"},
		UserCount:     10,
		CustomerState: "Maharashtra",
		Rates:         karnatakaRates,
	})
	if err != nil {
		t.Fatal(err)
	}
	if q.CGSTCents != 0 || q.SGSTCents != 0 {
		t.Errorf("cross-state should have zero CGST/SGST, got %d/%d", q.CGSTCents, q.SGSTCents)
	}
	if got, want := q.IGSTCents, int64(180_000); got != want {
		t.Errorf("IGST: got %d want %d", got, want)
	}
}

func TestBuildQuote_RequiresChainResolved(t *testing.T) {
	// audit_export requires audit_log — we select only audit_export and expect
	// audit_log to be auto-included on the quote. Also: no starter_bundle ⇒
	// 0 included users, so userCount=1 auto-adds an extra_users line.
	q, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys:  []string{"audit_export"},
		UserCount:     1,
		CustomerState: "Karnataka",
		Rates:         karnatakaRates,
	})
	if err != nil {
		t.Fatal(err)
	}
	keys := lineKeys(q)
	if !contains(keys, "audit_log") || !contains(keys, "audit_export") {
		t.Errorf("expected both audit_log and audit_export lines, got %v", keys)
	}
	// audit_log 150k + audit_export 100k + extra_users 1 × 50k = 300k.
	if got, want := q.SubtotalCents, int64(300_000); got != want {
		t.Errorf("subtotal: got %d want %d", got, want)
	}
}

func TestBuildQuote_UnknownFeatureKey(t *testing.T) {
	_, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys: []string{"made_up_feature"},
		UserCount:    1,
		Rates:        karnatakaRates,
	})
	if err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("expected unknown-feature error, got %v", err)
	}
}

func TestBuildQuote_UserCountMustBePositive(t *testing.T) {
	_, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys:  []string{"starter_bundle"},
		UserCount:     0,
		CustomerState: "Karnataka",
		Rates:         karnatakaRates,
	})
	if err == nil {
		t.Fatal("expected error for userCount=0")
	}
}

func TestBuildQuote_LinesAreInSortOrder(t *testing.T) {
	q, err := BuildQuote(seedCatalog(), QuoteInput{
		SelectedKeys:  []string{"webhooks", "starter_bundle", "audit_log"},
		UserCount:     5,
		CustomerState: "Karnataka",
		Rates:         karnatakaRates,
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(q.Lines); i++ {
		if q.Lines[i].SortOrder < q.Lines[i-1].SortOrder {
			t.Errorf("lines not sorted: %v", lineKeys(q))
		}
	}
}

// helpers

func lineKeys(q Quote) []string {
	out := make([]string, len(q.Lines))
	for i, l := range q.Lines {
		out[i] = l.FeatureKey
	}
	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
