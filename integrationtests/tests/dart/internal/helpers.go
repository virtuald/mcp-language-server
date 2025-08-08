// Package internal contains shared helpers for Dart tests
package internal

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
)

// GetTestSuite returns a test suite for Dart language server tests
func GetTestSuite(t *testing.T) *common.TestSuite {
	// Configure Dart LSP
	repoRoot, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatalf("Failed to get repo root: %v", err)
	}

	config := common.LSPTestConfig{
		Name:             "dart",
		Command:          "dart",
		Args:             []string{"language-server", "--protocol=lsp"},
		WorkspaceDir:     filepath.Join(repoRoot, "integrationtests/workspaces/dart"),
		InitializeTimeMs: 3000, // 3 seconds - Dart LSP can be slower to initialize
	}

	// Create a test suite
	suite := common.NewTestSuite(t, config)

	// Set up the suite
	err = suite.Setup()
	if err != nil {
		t.Fatalf("Failed to set up test suite: %v", err)
	}

	// Run dart pub get in the workspace to ensure dependencies are available
	// This ensures the Dart LSP can index pub packages for workspace symbols
	if err := runDartPubGet(suite.WorkspaceDir); err != nil {
		t.Logf("Warning: Failed to run dart pub get: %v", err)
		// Don't fail the test, as it might still work without dependencies
	}

	// Register cleanup
	t.Cleanup(func() {
		suite.Cleanup()
	})

	return suite
}

// runDartPubGet runs 'dart pub get' in the specified directory
func runDartPubGet(dir string) error {
	cmd := exec.Command("dart", "pub", "get")
	cmd.Dir = dir
	return cmd.Run()
}
