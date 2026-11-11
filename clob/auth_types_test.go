package clob

import "testing"

// TestBalanceAllowanceQueryConditionalV2 pins the V2 position selector added
// by the Protocol V2 migration. CTF positions keep using CONDITIONAL.
func TestBalanceAllowanceQueryConditionalV2(t *testing.T) {
	q := balanceAllowanceQuery(BalanceAllowanceParams{
		AssetType: AssetTypeConditionalV2,
		TokenID:   "123456789",
	}, SignatureTypePoly1271)
	if got := q.Get("asset_type"); got != "CONDITIONAL-V2" {
		t.Fatalf("asset_type = %q, want CONDITIONAL-V2", got)
	}
	if got := q.Get("token_id"); got != "123456789" {
		t.Fatalf("token_id = %q", got)
	}
	if got := q.Get("signature_type"); got != "3" {
		t.Fatalf("signature_type = %q, want 3", got)
	}
}
