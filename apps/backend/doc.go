// Package backend is the single Go binary of apotekerza.
//
// Code generators are declared here so that `go generate ./...` works from the
// repository root. Both generators are wired below: oapi-codegen for the API and sqlc
// for each module's SQL queries.
package backend

//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../api/openapi.yaml
//go:generate go tool sqlc generate
