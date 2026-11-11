package data

import (
	"context"
	"net/url"
)

// ServiceStatus is the Data API v2 freshness snapshot served by /v2/status.
// It is the service's own measure of how far the projections and ingestion
// cursors trail chain head, not a client-side timestamp comparison.
type ServiceStatus struct {
	ComputedAt Timestamp          `json:"computed_at"`
	AgeSeconds int64              `json:"age_seconds"`
	Serving    ServingFreshness   `json:"serving"`
	Ingestion  IngestionFreshness `json:"ingestion"`
}

// ServingFreshness reports the projections behind the reads: the worst
// mechanism lag and the name of the mechanism that produced it.
type ServingFreshness struct {
	LagSeconds *int64             `json:"lag_seconds"`
	Worst      *string            `json:"worst"`
	Mechanisms []ServingMechanism `json:"mechanisms"`
}

// ServingMechanism is one projection's freshness by its own clock.
type ServingMechanism struct {
	Name         string `json:"name"`
	AgeSeconds   int64  `json:"age_seconds"`
	BlocksBehind *int64 `json:"blocks_behind"`
}

// IngestionFreshness reports the live ingestion cursors and the furthest
// behind of them.
type IngestionFreshness struct {
	Cursors        int         `json:"cursors"`
	ChainID        int64       `json:"chain_id"`
	Lagging        []CursorLag `json:"lagging"`
	MaxSyncedBlock *int64      `json:"max_synced_block"`
	MinSyncedBlock *int64      `json:"min_synced_block"`
	MostLagged     *CursorLag  `json:"most_lagged"`
	Network        *string     `json:"network"`
}

// CursorLag is one ingestion stream's distance from the furthest-along stream.
type CursorLag struct {
	Source    string `json:"source"`
	Block     int64  `json:"block"`
	BehindMax int64  `json:"behind_max"`
}

// GetStatus reads the Data API's freshness snapshot. It is a non-paginated
// envelope read and follows the same indexed-read 429 policy as the other
// data methods.
func (c *Client) GetStatus(ctx context.Context) (*ServiceStatus, error) {
	return getValue[ServiceStatus](ctx, c, "/v2/status", url.Values{}, false)
}
