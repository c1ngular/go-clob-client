package gamma

import (
	"strings"
	"testing"
)

func TestMarketTradingID(t *testing.T) {
	t.Parallel()

	v1 := ProtocolVersionV1
	v2 := ProtocolVersionV2
	unknown := ProtocolVersion("v3")

	tests := []struct {
		name    string
		market  Market
		index   int
		want    string
		wantErr string
	}{
		{
			name: "v2 selects the position ID",
			market: Market{
				ID:           "m",
				Version:      &v2,
				Outcomes:     []string{"Yes", "No"},
				PositionIDs:  []string{"111", "112"},
				CLOBTokenIDs: []string{"222", "223"},
			},
			index: 1,
			want:  "112",
		},
		{
			name: "v1 selects the CLOB token ID",
			market: Market{
				ID:           "m",
				Version:      &v1,
				Outcomes:     []string{"Yes", "No"},
				PositionIDs:  []string{"111", "112"},
				CLOBTokenIDs: []string{"222", "223"},
			},
			index: 1,
			want:  "223",
		},
		{
			name:    "missing version",
			market:  Market{ID: "m", Outcomes: []string{"Yes"}},
			wantErr: "missing protocol version",
		},
		{
			name: "empty version",
			market: Market{
				ID:      "m",
				Version: new(ProtocolVersion),
				Outcomes: []string{
					"Yes",
				},
			},
			wantErr: "missing protocol version",
		},
		{
			name:    "unsupported version",
			market:  Market{ID: "m", Version: &unknown, Outcomes: []string{"Yes"}},
			wantErr: `unsupported protocol version "v3"`,
		},
		{
			name:    "negative index",
			market:  Market{ID: "m", Version: &v2, Outcomes: []string{"Yes"}},
			index:   -1,
			wantErr: "out of range",
		},
		{
			name:    "index past outcomes",
			market:  Market{ID: "m", Version: &v2, Outcomes: []string{"Yes"}},
			index:   1,
			wantErr: "out of range",
		},
		{
			name: "missing ID array",
			market: Market{
				ID:       "m",
				Version:  &v2,
				Outcomes: []string{"Yes"},
			},
			wantErr: "has no v2 identifier",
		},
		{
			name: "empty identifier",
			market: Market{
				ID:          "m",
				Version:     &v2,
				Outcomes:    []string{"Yes"},
				PositionIDs: []string{" "},
			},
			wantErr: "empty v2 identifier",
		},
		{
			name: "non-decimal identifier",
			market: Market{
				ID:          "m",
				Version:     &v2,
				Outcomes:    []string{"Yes"},
				PositionIDs: []string{"12a"},
			},
			wantErr: "is not a decimal integer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.market.TradingID(tc.index)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("TradingID() error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("TradingID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("TradingID() = %q, want %q", got, tc.want)
			}
		})
	}
}
