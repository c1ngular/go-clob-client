// Package data reads Polymarket's current Data API v2.
//
// It covers market and combo portfolios, trades, wallet activity, PnL and volume,
// holders, price history, resolutions, trader and builder leaderboards, and
// indexed approvals. No trading writes are made.
//
// List methods return Page values containing Items, HasMore, and an opaque
// NextCursor. Iter methods walk those exact cursors until completion, cancellation,
// or consumer break. Short and empty pages do not imply completion; a repeated
// cursor is reported as ErrCursorCycle. Reuse cursors only with their original
// filters. Each request preserves HTTP/API errors; retries are caller-controlled.
//
// Monetary and share values use DecimalString to retain the service's full
// precision. Timestamp embeds time.Time and normalizes instants to UTC.
// GetUserStats and GetLeaderboardStanding return nil, nil for absent rows.
//
// The retired v1 routes are isolated in
// github.com/nijaru/go-clob-client/data/legacy and return HTTP 410. There is no
// implicit API-version switch or fallback after a v2 failure.
package data
