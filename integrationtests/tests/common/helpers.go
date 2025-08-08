package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Logger is an interface for logging in tests
type Logger interface {
	Printf(format string, v ...any)
}

// Helper to copy directories recursively
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err = CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper to copy a single file
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close source file: %v\n", err)
		}
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close destination file: %v\n", err)
		}
	}()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// CleanupTestSuites is a helper to clean up all test suites in a test
func CleanupTestSuites(suites ...*TestSuite) {
	for _, suite := range suites {
		if suite != nil {
			suite.Cleanup()
		}
	}
}

// use instead of runtime.GOROOT which is deprecated
func getGoRoot() string {
	cmd := exec.Command("go", "env", "GOROOT")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(out.String())
}

// normalizePaths replaces absolute paths in the result with placeholder paths for consistent snapshots
func normalizePaths(_ *testing.T, input string) string {
	// No need to get the repo root - we're just looking for patterns

	// But this is useful
	goroot := getGoRoot()

	// Simple approach: just replace any path segments that contain workspace/
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		// Any line containing a path to a workspace file needs normalization
		if strings.Contains(line, "/workspace/") {
			// Extract everything after /workspace/
			parts := strings.Split(line, "/workspace/")
			if len(parts) > 1 {
				// Replace with a simple placeholder path
				lines[i] = "/TEST_OUTPUT/workspace/" + parts[1]
			}
		}
		// Some tests, e.g. clangd, may include fully qualified paths to the base /workspaces/ directory
		if strings.Contains(line, "/workspaces/") {
			// Extract everything after /workspace/
			parts := strings.Split(line, "/workspaces/")
			if len(parts) > 1 {
				// Replace with a simple placeholder path
				lines[i] = "/TEST_OUTPUT/workspace/" + parts[1]
			}
		}
		if strings.Contains(line, goroot) {
			parts := strings.Split(line, goroot)
			if len(parts) > 1 {
				// Replace with a simple placeholder path
				lines[i] = "/GOROOT" + parts[1]
			}
		}

		// Normalize Dart SDK paths (various possible locations)
		// Handle paths like /Users/*/fvm/versions/*/bin/cache/dart-sdk/
		// or /opt/hostedtoolcache/dart/*/x64/
		// or /home/*/.cache/dart-sdk/
		if strings.Contains(line, "/dart-sdk/") || strings.Contains(line, "/lib/") {
			// Look for common Dart SDK path patterns
			patterns := []struct {
				contains string
				prefix   string
			}{
				{"/fvm/versions/", "/DART_SDK/"},
				{"/hostedtoolcache/dart/", "/DART_SDK/"},
				{"/.cache/dart-sdk/", "/DART_SDK/"},
				{"/bin/cache/dart-sdk/", "/DART_SDK/"},
			}

			for _, pattern := range patterns {
				if strings.Contains(line, pattern.contains) {
					// Find where the SDK lib directory starts
					if idx := strings.Index(line, "/lib/"); idx != -1 {
						before := line[:idx]
						after := line[idx+1:] // Skip the leading slash to avoid double slash
						// Replace everything before /lib/ with /DART_SDK
						if strings.Contains(before, "File: ") {
							lines[i] = "File: " + pattern.prefix + after
						} else {
							lines[i] = pattern.prefix + after
						}
						break
					}
				}
			}
		}

		// Normalize pub-cache paths
		// Handle paths like /Users/*/.pub-cache/ or /home/*/.pub-cache/
		if strings.Contains(line, "/.pub-cache/") {
			parts := strings.Split(line, "/.pub-cache/")
			if len(parts) > 1 {
				// Replace with a simple placeholder path
				if strings.Contains(line, "File: ") {
					lines[i] = "File: /PUB_CACHE/" + parts[1]
				} else {
					lines[i] = "/PUB_CACHE/" + parts[1]
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// sortDefinitions sorts multi-definition results for deterministic ordering
func sortDefinitions(input string) string {
	// Split by definition separators (---\n\n)
	definitions := strings.Split(input, "---\n\n")
	if len(definitions) <= 1 {
		return input // Single definition or no definitions
	}

	// Sort definitions as raw strings
	sort.Strings(definitions)

	return strings.Join(definitions, "---\n\n")
}

// FindRepoRoot locates the repository root by looking for specific indicators
// Exported so it can be used by other packages
func FindRepoRoot() (string, error) {
	// Start from the current directory and walk up until we find the main.go file
	// which is at the repository root
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if this is the repo root (has a go.mod file)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Found the repo root
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the filesystem root without finding repo root
			return "", fmt.Errorf("repository root not found")
		}
		dir = parent
	}
}

// SnapshotTest compares the actual result against an expected result file
// If the file doesn't exist or UPDATE_SNAPSHOTS=true env var is set, it will update the snapshot
func SnapshotTest(t *testing.T, languageName, toolName, testName, actualResult string) {
	// Normalize paths in the result to avoid system-specific paths in snapshots
	actualResult = normalizePaths(t, actualResult)

	// Sort definitions for Dart tests to ensure deterministic ordering
	if languageName == "dart" {
		actualResult = sortDefinitions(actualResult)
	}

	// Get the absolute path to the snapshots directory
	repoRoot, err := FindRepoRoot()
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}

	// Build path based on language/tool/testName hierarchy
	snapshotDir := filepath.Join(repoRoot, "integrationtests", "snapshots", languageName, toolName)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		t.Fatalf("Failed to create snapshots directory: %v", err)
	}

	snapshotFile := filepath.Join(snapshotDir, testName+".snap")

	// Use a package-level flag to control snapshot updates
	updateFlag := os.Getenv("UPDATE_SNAPSHOTS") == "true"

	// If snapshot doesn't exist or update flag is set, write the snapshot
	_, err = os.Stat(snapshotFile)
	if os.IsNotExist(err) || updateFlag {
		if err := os.WriteFile(snapshotFile, []byte(actualResult), 0644); err != nil {
			t.Fatalf("Failed to write snapshot: %v", err)
		}
		if os.IsNotExist(err) {
			t.Logf("Created new snapshot: %s", snapshotFile)
		} else {
			t.Logf("Updated snapshot: %s", snapshotFile)
		}
		return
	}

	// Read the expected result
	expectedBytes, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("Failed to read snapshot: %v", err)
	}
	expected := string(expectedBytes)

	// Compare the results
	if expected != actualResult {
		t.Errorf("Result doesn't match snapshot.\nExpected:\n%s\n\nActual:\n%s", expected, actualResult)

		// Create a diff file for debugging
		diffFile := snapshotFile + ".diff"
		diffContent := fmt.Sprintf("=== Expected ===\n%s\n\n=== Actual ===\n%s", expected, actualResult)
		if err := os.WriteFile(diffFile, []byte(diffContent), 0644); err != nil {
			t.Logf("Failed to write diff file: %v", err)
		} else {
			t.Logf("Wrote diff to: %s", diffFile)
		}
	}
}
