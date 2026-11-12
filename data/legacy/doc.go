// Package legacy preserves the retired v1 Polymarket Data API contract for
// historical protocol work.
//
// The v1 routes returned HTTP 410 after the 2026-11-02 retirement. New
// integrations should use github.com/nijaru/go-clob-client/data for the v2
// API. This package retains the v1 response models and offset-based endpoints
// exposed by the Rust SDK; it never acts as an automatic fallback after a v2
// request fails.
//
// Create a client with New, or stream list endpoints with their Iter methods.
// No signing or authenticated trading operations are performed by this package.
//
// Deprecated: Data API v1 is retired; the routes this package calls return
// HTTP 410. Use package data instead.
package legacy
