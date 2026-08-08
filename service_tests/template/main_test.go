//go:build integration

// Package templatetests holds the service tests for the template resource.
package templatetests

import (
	"os"
	"testing"

	"github.com/snc-software/go-template-service/service_tests/platform"
)

func TestMain(m *testing.M) {
	os.Exit(platform.Run(m))
}
