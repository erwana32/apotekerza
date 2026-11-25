// Package backend is the single Go binary of apotekerza.
//
// Code generators are declared here so that `go generate ./...` works from the
// repository root. The real OpenAPI and sqlc generators are wired in Phase 0.3
// and in the slices that add SQL queries.
package backend

//go:generate echo "No generators wired yet (Phase 0.3 adds the OpenAPI generator)."
