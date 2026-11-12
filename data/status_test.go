package data

import (
	"net/http"
	"testing"
)

func TestGetStatus(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/status" {
			t.Errorf("path = %q, want /v2/status", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"data": {
				"computed_at": "2026-11-12T04:30:00Z",
				"age_seconds": 12,
				"serving": {
					"lag_seconds": 3,
					"worst": "custody_balances",
					"mechanisms": [
						{"name": "activity_feed", "age_seconds": 1},
						{"name": "custody_balances", "age_seconds": 3, "blocks_behind": 2}
					]
				},
				"ingestion": {
					"cursors": 2,
					"chain_id": 137,
					"lagging": [{"source": "ctf_transfer", "block": 100, "behind_max": 4}],
					"max_synced_block": 104,
					"min_synced_block": 100,
					"most_lagged": {"source": "ctf_transfer", "block": 100, "behind_max": 4},
					"network": "polygon"
				}
			}
		}`))
	})

	status, err := client.GetStatus(t.Context())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.AgeSeconds != 12 || status.ComputedAt.IsZero() {
		t.Fatalf("status = %+v, want age 12 and computed_at", status)
	}
	if status.Serving.LagSeconds == nil || *status.Serving.LagSeconds != 3 {
		t.Fatalf("serving lag = %v, want 3", status.Serving.LagSeconds)
	}
	if status.Serving.Worst == nil || *status.Serving.Worst != "custody_balances" {
		t.Fatalf("worst = %v, want custody_balances", status.Serving.Worst)
	}
	if len(status.Serving.Mechanisms) != 2 || status.Serving.Mechanisms[1].BlocksBehind == nil ||
		*status.Serving.Mechanisms[1].BlocksBehind != 2 {
		t.Fatalf("mechanisms = %+v", status.Serving.Mechanisms)
	}
	if status.Ingestion.ChainID != 137 || status.Ingestion.Cursors != 2 ||
		status.Ingestion.MostLagged == nil || status.Ingestion.MostLagged.BehindMax != 4 {
		t.Fatalf("ingestion = %+v", status.Ingestion)
	}
}

func TestGetStatusNullLag(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(
			[]byte(
				`{"data":{"computed_at":"2026-11-12T04:30:00Z","age_seconds":1,"serving":{"lag_seconds":null,"mechanisms":[]},"ingestion":{"cursors":0,"chain_id":137,"lagging":[]}}}`,
			),
		)
	})
	status, err := client.GetStatus(t.Context())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.Serving.LagSeconds != nil {
		t.Fatalf("lag = %v, want nil", status.Serving.LagSeconds)
	}
}
