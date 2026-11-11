package clob

import "testing"

// TestDefaultHost pins the documented CLOB host. The old clob-v2 host is not
// reachable and is not used by any current API reference or SDK example.
func TestDefaultHost(t *testing.T) {
	if DefaultHost != "https://clob.polymarket.com" {
		t.Fatalf("DefaultHost = %q, want the documented clob.polymarket.com", DefaultHost)
	}
}
