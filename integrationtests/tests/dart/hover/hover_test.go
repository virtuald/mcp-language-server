package hover_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/dart/internal"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestHover tests hover functionality with the Dart language server
func TestHover(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		line           int
		column         int
		expectedText   string // Text that should be in the hover result
		unexpectedText string // Text that should NOT be in the hover result (optional)
		snapshotName   string
	}{
		// Tests using types.dart file
		{
			name:         "SharedClass",
			file:         "types.dart",
			line:         5,
			column:       7,
			expectedText: "SharedClass",
			snapshotName: "shared-class",
		},
		{
			name:         "ClassMethod",
			file:         "types.dart",
			line:         15,
			column:       10,
			expectedText: "process",
			snapshotName: "class-method",
		},
		{
			name:         "Interface",
			file:         "types.dart",
			line:         21,
			column:       16,
			expectedText: "SharedInterface",
			snapshotName: "interface",
		},
		{
			name:         "TypeAlias",
			file:         "types.dart",
			line:         27,
			column:       9,
			expectedText: "ProcessFunction",
			snapshotName: "type-alias",
		},
		{
			name:         "Constant",
			file:         "types.dart",
			line:         30,
			column:       14,
			expectedText: "SHARED_CONSTANT",
			snapshotName: "constant",
		},
		{
			name:         "Enum",
			file:         "types.dart",
			line:         33,
			column:       6,
			expectedText: "Color",
			snapshotName: "enum",
		},
		// Tests using helper.dart file
		{
			name:         "HelperClass",
			file:         "helper.dart",
			line:         4,
			column:       7,
			expectedText: "HelperClass",
			snapshotName: "helper-class",
		},
		{
			name:         "OverrideMethod",
			file:         "helper.dart",
			line:         10,
			column:       8,
			expectedText: "doSomething",
			snapshotName: "override-method",
		},
		{
			name:         "Function",
			file:         "helper.dart",
			line:         25,
			column:       14,
			expectedText: "createHelper",
			snapshotName: "function",
		},
		{
			name:         "GlobalVariable",
			file:         "helper.dart",
			line:         30,
			column:       7,
			expectedText: "globalHelper",
			snapshotName: "global-variable",
		},
		// Tests using main.dart file
		{
			name:         "MainFunction",
			file:         "main.dart",
			line:         5,
			column:       6,
			expectedText: "main",
			snapshotName: "main-function",
		},
		{
			name:         "Variable",
			file:         "main.dart",
			line:         8,
			column:       7,
			expectedText: "helper",
			snapshotName: "variable",
		},
	}

	suite := internal.GetTestSuite(t)

	// Wait for initialization
	time.Sleep(time.Duration(suite.Config.InitializeTimeMs) * time.Millisecond)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(suite.WorkspaceDir, tt.file)

			// Get hover information
			result, err := tools.GetHoverInfo(context.Background(), suite.Client, filePath, tt.line, tt.column)
			if err != nil {
				t.Fatalf("Failed to get hover info: %v", err)
			}

			// Check if expected text is present
			if tt.expectedText != "" && !strings.Contains(result, tt.expectedText) {
				t.Errorf("Expected text '%s' not found in hover result: %s", tt.expectedText, result)
			}

			// Check if unexpected text is absent
			if tt.unexpectedText != "" && strings.Contains(result, tt.unexpectedText) {
				t.Errorf("Unexpected text '%s' found in hover result: %s", tt.unexpectedText, result)
			}

			// Perform snapshot test
			common.SnapshotTest(t, suite.LanguageName, "hover", tt.snapshotName, result)
		})
	}
}

// TestHoverNoInfo tests hover on positions without hover information
func TestHoverNoInfo(t *testing.T) {
	suite := internal.GetTestSuite(t)

	// Wait for initialization
	time.Sleep(time.Duration(suite.Config.InitializeTimeMs) * time.Millisecond)

	// Test hover on an empty line or comment
	filePath := filepath.Join(suite.WorkspaceDir, "main.dart")
	result, err := tools.GetHoverInfo(context.Background(), suite.Client, filePath, 1, 1)
	if err != nil {
		t.Fatalf("Failed to get hover info: %v", err)
	}

	// Should indicate no hover information
	if !strings.Contains(result, "No hover information") {
		t.Logf("Result when no hover info expected: %s", result)
	}

	// Snapshot test for no hover info case
	common.SnapshotTest(t, suite.LanguageName, "hover", "no-hover-info", result)
}
