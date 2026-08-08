//go:build integration

// Package templatetests holds the service tests for the template resource.
package templatetests

import (
	"os"
	"testing"

	testinfrastructure "github.com/snc-software/go-template-service/service_tests/test_infrastructure"
)

func TestMain(m *testing.M) {
	os.Exit(testinfrastructure.Run(m))
}
