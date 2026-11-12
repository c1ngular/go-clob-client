package clob

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/quagmt/udecimal"
)

func TestCreateOrderRejectsTokenIDOnV2Market(t *testing.T) {
	t.Parallel()

	client := newMarketVersionTestClient(t, newTradingFixtureServer(t))
	_, err := client.CreateOrder(t.Context(), OrderArgs{
		TokenID: "999",
		Price:   udecimal.MustParse("0.5"),
		Size:    udecimal.MustParse("10"),
		Side:    SideBuy,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "is v2; use PositionID") {
		t.Fatalf("TokenID on a v2 market: error = %v", err)
	}
}

func TestCreateOrderRejectsPositionIDOnV1Market(t *testing.T) {
	t.Parallel()

	client := newMarketVersionTestClient(t, newTradingFixtureServer(t))
	_, err := client.CreateOrder(t.Context(), OrderArgs{
		PositionID: "123",
		Price:      udecimal.MustParse("0.5"),
		Size:       udecimal.MustParse("10"),
		Side:       SideBuy,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "is v1; use TokenID") {
		t.Fatalf("PositionID on a v1 market: error = %v", err)
	}
}

func TestCreateMarketOrderRejectsTokenIDOnV2Market(t *testing.T) {
	t.Parallel()

	client := newMarketVersionTestClient(t, newTradingFixtureServer(t))
	_, err := client.CreateMarketOrder(t.Context(), MarketOrderArgs{
		TokenID: "999",
		Amount:  udecimal.MustParse("10"),
		Side:    SideBuy,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "is v2; use PositionID") {
		t.Fatalf("TokenID market order on a v2 market: error = %v", err)
	}
}

func TestCreateOrderRejectsUnsupportedMarketVersion(t *testing.T) {
	t.Parallel()

	server := newMarketVersionServer(t, "v3", "123")
	client := newMarketVersionTestClient(t, server)
	_, err := client.CreateOrder(t.Context(), OrderArgs{
		TokenID: "123",
		Price:   udecimal.MustParse("0.5"),
		Size:    udecimal.MustParse("10"),
		Side:    SideBuy,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), `unsupported protocol version "v3"`) {
		t.Fatalf("unsupported market version: error = %v", err)
	}
}

func TestCreateOrderRejectsMissingMarketVersion(t *testing.T) {
	t.Parallel()

	server := newMarketVersionServer(t, "", "123")
	client := newMarketVersionTestClient(t, server)
	_, err := client.CreateOrder(t.Context(), OrderArgs{
		TokenID: "123",
		Price:   udecimal.MustParse("0.5"),
		Size:    udecimal.MustParse("10"),
		Side:    SideBuy,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "did not report a protocol version") {
		t.Fatalf("missing market version: error = %v", err)
	}
}

func newMarketVersionTestClient(t *testing.T, server *httptest.Server) *SignerClient {
	t.Helper()
	t.Cleanup(server.Close)

	client, err := NewSignerClient(Config{
		Host:       server.URL,
		PrivateKey: "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
		RetryMax:   0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.saltGenerator = func() (uint64, error) { return 1, nil }
	return client
}

// newMarketVersionServer serves one market with the supplied compact version
// marker. An empty version omits the v field, mirroring an older deployment.
func newMarketVersionServer(t *testing.T, version, tokenID string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case versionEndpoint:
			_, _ = w.Write([]byte(`{"version":2}`))
		case marketsByTokenEndpoint + tokenID:
			_, _ = w.Write([]byte(`{"condition_id":"cid"}`))
		case clobMarketEndpoint + "/cid":
			versionField := ""
			if version != "" {
				versionField = fmt.Sprintf(`"v":%q,`, version)
			}
			_, _ = fmt.Fprintf(
				w,
				`{%s"c":"cid","mts":"0.01","nr":false,"t":[{"t":%q}]}`,
				versionField,
				tokenID,
			)
		case tickSizeEndpoint:
			_, _ = w.Write([]byte(`{"minimum_tick_size":"0.01"}`))
		case negRiskEndpoint:
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
