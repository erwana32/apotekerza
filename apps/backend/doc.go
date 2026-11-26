// Package backend is the single Go binary of apotekerza.
//
// Code generators are declared here so that `go generate ./...` works from the
// repository root. The OpenAPI generator is wired below; sqlc is added with the
// first SQL queries.
package backend

//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../api/openapi.yaml
