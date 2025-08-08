package definition_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/dart/internal"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestDartDefinition tests definition lookup functionality with the Dart language server
func TestDartDefinition(t *testing.T) {
	tests := []struct {
		name         string
		symbolName   string
		found        bool
		description  string
		snapshotName string
	}{
		{
			name:         "LocalClass",
			symbolName:   "HelperClass",
			found:        true,
			description:  "Local class defined in helper.dart",
			snapshotName: "local-class",
		},
		{
			name:         "SharedClassName",
			symbolName:   "SharedClass",
			found:        true,
			description:  "Shared class in types.dart",
			snapshotName: "shared-class-name",
		},
		{
			name:         "SharedInterface",
			symbolName:   "SharedInterface",
			found:        true,
			description:  "Shared interface in types.dart",
			snapshotName: "shared-interface",
		},
		{
			name:         "GlobalFunction",
			symbolName:   "createHelper",
			found:        true,
			description:  "Global function in helper.dart",
			snapshotName: "global-function",
		},
		{
			name:         "GlobalVariable",
			symbolName:   "globalHelper",
			found:        true,
			description:  "Global variable in helper.dart",
			snapshotName: "global-variable",
		},
		{
			name:         "Enum",
			symbolName:   "Color",
			found:        true,
			description:  "Enum in types.dart",
			snapshotName: "enum",
		},
		{
			name:         "MainFunction",
			symbolName:   "main",
			found:        true,
			description:  "Main function in main.dart",
			snapshotName: "main-function",
		},
		{
			name:         "MethodQualified",
			symbolName:   "HelperClass.process",
			found:        true,
			description:  "Method with qualified name",
			snapshotName: "method-qualified",
		},
		{
			name:         "MethodSimple",
			symbolName:   "process",
			found:        true,
			description:  "Method with simple name",
			snapshotName: "method-simple",
		},
		{
			name:         "NonexistentSymbol",
			symbolName:   "NonExistentSymbol",
			found:        false,
			description:  "Symbol that doesn't exist",
			snapshotName: "nonexistent-symbol",
		},
		{
			name:         "Typedef",
			symbolName:   "ProcessFunction",
			found:        true,
			description:  "Type alias in types.dart",
			snapshotName: "typedef",
		},
		{
			name:         "Constant",
			symbolName:   "SHARED_CONSTANT",
			found:        true,
			description:  "Constant in types.dart",
			snapshotName: "constant",
		},
	}

	suite := internal.GetTestSuite(t)

	// Wait for initialization
	time.Sleep(time.Duration(suite.Config.InitializeTimeMs) * time.Millisecond)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tools.ReadDefinition(context.Background(), suite.Client, tt.symbolName)
			if err != nil {
				t.Fatalf("Failed to read definition: %v", err)
			}

			// Check if the symbol was found
			notFoundMsg := fmt.Sprintf("%s not found", tt.symbolName)
			if !tt.found {
				// Symbol should not be found
				if result != notFoundMsg {
					t.Errorf("Expected symbol %s to not be found, but got: %s", tt.symbolName, result)
				}
			} else {
				// Symbol should be found
				if result == notFoundMsg {
					t.Errorf("Expected symbol %s to be found, but got: %s", tt.symbolName, result)
				} else if !strings.Contains(result, "Symbol:") {
					t.Errorf("Result should contain symbol information, got: %s", result)
				}
			}

			// Snapshot test for consistent behavior
			common.SnapshotTest(t, suite.LanguageName, "definition", tt.snapshotName, result)
		})
	}
}

// TestDartDefinitionWorkspaceSymbolBehavior tests the workspace/symbol behavior specifically
func TestDartDefinitionWorkspaceSymbolBehavior(t *testing.T) {
	suite := internal.GetTestSuite(t)

	// Wait for initialization
	time.Sleep(time.Duration(suite.Config.InitializeTimeMs) * time.Millisecond)

	// Test different query patterns to understand workspace/symbol behavior
	queryTests := []struct {
		name  string
		query string
	}{
		{"exact-match", "HelperClass"},
		{"partial-match", "Helper"},
		{"lowercase", "helperclass"},
		{"wildcard-prefix", "*Helper"},
		{"regex-pattern", "^Helper.*"},
		{"qualified-name", "helper.HelperClass"},
		{"method-name", "process"},
		{"main", "main"},
		{"shared-class", "SharedClass"},
		{"enum-value", "Color"},
	}

	for _, tt := range queryTests {
		t.Run("query-"+tt.name, func(t *testing.T) {
			result, err := tools.ReadDefinition(context.Background(), suite.Client, tt.query)
			if err != nil {
				t.Fatalf("Failed to read definition: %v", err)
			}

			// Document the current behavior for each query pattern
			t.Logf("Query '%s' result: %s", tt.query, result)
			common.SnapshotTest(t, suite.LanguageName, "definition", "query-"+tt.name, result)
		})
	}
}
