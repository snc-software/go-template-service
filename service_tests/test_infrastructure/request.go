//go:build integration

package testinfrastructure

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"slices"
	"testing"
)

// Response is a completed HTTP response with its body already drained.
type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Get issues a GET against path and reads the whole response.
func (api *API) Get(t *testing.T, path string) Response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, api.BaseURL+path, http.NoBody)
	if err != nil {
		t.Fatalf("build GET %s: %v", path, err)
	}

	response, err := api.Client.Do(request)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}

	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body of GET %s: %v", path, err)
	}

	return Response{
		StatusCode:  response.StatusCode,
		ContentType: response.Header.Get("Content-Type"),
		Body:        body,
	}
}

// DecodeJSON unmarshals the response body into target.
func (response Response) DecodeJSON(t *testing.T, target any) {
	t.Helper()

	if err := json.Unmarshal(response.Body, target); err != nil {
		t.Fatalf("decode response body %q: %v", string(response.Body), err)
	}
}

// JSONFields returns the top-level field names of the response body, sorted.
// Decoding into an application type proves the values; this proves the names,
// which a rename of a struct tag would otherwise carry along unnoticed.
func (response Response) JSONFields(t *testing.T) []string {
	t.Helper()

	var fields map[string]json.RawMessage
	response.DecodeJSON(t, &fields)

	return slices.Sorted(maps.Keys(fields))
}
