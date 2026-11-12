package gamma

import (
	"bytes"
	"encoding/json"
	"testing"
)

// TestMarketCapturesRawPayload pins the evidence contract: the decoded model
// keeps the venue's exact object, including fields the model does not know,
// and marshaling the model does not leak that raw copy back onto the wire.
func TestMarketCapturesRawPayload(t *testing.T) {
	raw := []byte(
		`{"id":"42","slug":"btc-updown-5m-1","version":"v2","outcomes":"[\"Up\",\"Down\"]","positionIds":["1","2"],"unknownField":{"keep":true}}`,
	)
	var market Market
	if err := json.Unmarshal(raw, &market); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !bytes.Equal(market.Raw, raw) {
		t.Fatalf("raw = %s, want the exact venue object", market.Raw)
	}

	encoded, err := json.Marshal(market)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(encoded, []byte(`"Raw"`)) {
		t.Fatalf("marshaled model leaked the raw field: %s", encoded)
	}
	if bytes.Contains(encoded, []byte(`"unknownField"`)) {
		t.Fatalf("marshaled model leaked an unmodeled field: %s", encoded)
	}
	if !bytes.Contains(encoded, []byte(`"slug":"btc-updown-5m-1"`)) {
		t.Fatalf("marshaled model lost the modeled fields: %s", encoded)
	}

	// A second decode must replace the evidence, not append to it.
	second := []byte(`{"id":"43","slug":"btc-updown-5m-2"}`)
	if err := json.Unmarshal(second, &market); err != nil {
		t.Fatalf("second unmarshal: %v", err)
	}
	if !bytes.Equal(market.Raw, second) {
		t.Fatalf("raw = %s, want the second venue object", market.Raw)
	}
}
