package pricing

import "testing"

func TestComputeTax(t *testing.T) {
	rates := Rates{HomeState: "Karnataka", CGSTPct: 9, SGSTPct: 9, IGSTPct: 18}

	tests := []struct {
		name          string
		taxable       int64
		customerState string
		wantCGST      int64
		wantSGST      int64
		wantIGST      int64
	}{
		{"intra-state same case", 100_000, "Karnataka", 9_000, 9_000, 0},
		{"intra-state different case", 100_000, "karnataka", 9_000, 9_000, 0},
		{"intra-state whitespace tolerated", 100_000, "  Karnataka  ", 9_000, 9_000, 0},
		{"inter-state", 100_000, "Maharashtra", 0, 0, 18_000},
		{"empty state defaults to intra", 100_000, "", 9_000, 9_000, 0},
		{"zero taxable amount", 0, "Karnataka", 0, 0, 0},
		{"negative taxable amount", -500, "Karnataka", 0, 0, 0},
		// 333 paise at 9% = 29.97 → rounds to 30; 18% = 59.94 → rounds to 60
		{"rounds nearest", 333, "Karnataka", 30, 30, 0},
		{"rounds nearest IGST", 333, "Tamil Nadu", 0, 0, 60},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeTax(tc.taxable, tc.customerState, rates)
			if got.CGSTCents != tc.wantCGST || got.SGSTCents != tc.wantSGST || got.IGSTCents != tc.wantIGST {
				t.Fatalf("got CGST=%d SGST=%d IGST=%d, want CGST=%d SGST=%d IGST=%d",
					got.CGSTCents, got.SGSTCents, got.IGSTCents, tc.wantCGST, tc.wantSGST, tc.wantIGST)
			}
			wantTotal := tc.wantCGST + tc.wantSGST + tc.wantIGST
			if got.TotalCents != wantTotal {
				t.Fatalf("total mismatch: got %d, want %d", got.TotalCents, wantTotal)
			}
		})
	}
}

func TestComputeTaxIntraStateOnlyHasCGSTSGST(t *testing.T) {
	rates := Rates{HomeState: "Karnataka", CGSTPct: 9, SGSTPct: 9, IGSTPct: 18}
	got := ComputeTax(50_000, "Karnataka", rates)
	if got.IGSTCents != 0 {
		t.Errorf("intra-state should have IGST=0, got %d", got.IGSTCents)
	}
	if got.CGSTPct != 9 || got.SGSTPct != 9 {
		t.Errorf("intra-state rates not populated correctly: CGSTPct=%v SGSTPct=%v", got.CGSTPct, got.SGSTPct)
	}
}

func TestComputeTaxInterStateOnlyHasIGST(t *testing.T) {
	rates := Rates{HomeState: "Karnataka", CGSTPct: 9, SGSTPct: 9, IGSTPct: 18}
	got := ComputeTax(50_000, "Delhi", rates)
	if got.CGSTCents != 0 || got.SGSTCents != 0 {
		t.Errorf("inter-state should have CGST=SGST=0, got CGST=%d SGST=%d", got.CGSTCents, got.SGSTCents)
	}
	if got.IGSTPct != 18 {
		t.Errorf("inter-state IGSTPct not populated: %v", got.IGSTPct)
	}
}
