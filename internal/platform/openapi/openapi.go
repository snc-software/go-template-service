// Package openapi serves the generated OpenAPI document and the Scalar
// reference UI that renders it.
package openapi

import (
	"fmt"
	"net/http"

	scalargo "github.com/bdpiprava/scalar-go"
)

const (
	// SpecPath serves the OpenAPI document itself.
	SpecPath = "/openapi.json"
	// ReferencePath serves the Scalar reference UI.
	ReferencePath = "/docs"

	// scalarCDN is pinned so a CDN release cannot change the UI under a
	// deployed service. Replace it with a self-hosted URL to remove the
	// external request entirely.
	scalarCDN = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.64.1"
)

// Endpoints serves the OpenAPI document and its reference UI.
type Endpoints struct {
	spec []byte
	html string
}

// NewEndpoints renders the Scalar UI once at startup, so a request to
// ReferencePath is a buffer write and a bad configuration fails on boot.
//
// The spec is inlined into the page rather than fetched from SpecPath, because
// scalar-go only accepts absolute spec URLs and the host is not known here.
func NewEndpoints(title string, spec []byte) (*Endpoints, error) {
	html, err := scalargo.NewV2(
		scalargo.WithSpecBytes(spec),
		scalargo.WithCDN(scalarCDN),
		scalargo.WithMetaDataOpts(scalargo.WithTitle(title)),
		scalargo.WithTheme(scalargo.ThemeMoon),
		scalargo.WithLayout(scalargo.LayoutModern),
		scalargo.WithDefaultHTTPClient("go", "native"),
		scalargo.WithSidebarVisibility(true),
	)
	if err != nil {
		return nil, fmt.Errorf("render scalar reference: %w", err)
	}

	return &Endpoints{spec: spec, html: html}, nil
}

// Spec serves the OpenAPI document generated from the swag annotations.
func (endpoints *Endpoints) Spec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(endpoints.spec)
}

// Reference serves the Scalar UI, which fetches the document from SpecPath.
func (endpoints *Endpoints) Reference(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(endpoints.html))
}
