package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

func ReadDefinition(ctx context.Context, client *lsp.Client, symbolName string) (string, error) {
	symbolName, results, err := QuerySymbol(ctx, client, symbolName)
	if err != nil {
		return "", err
	}

	var definitions []string
	for _, symbol := range results {
		kind := ""
		container := ""

		// Skip symbols that we are not looking for. workspace/symbol may return
		// a large number of fuzzy matches. This handles BaseSymbolInformation
		doesSymbolMatch := func(vKind protocol.SymbolKind, vContainerName string) bool {
			thisName := symbol.GetName()

			kind = fmt.Sprintf("Kind: %s\n", protocol.TableKindMap[vKind])
			if vContainerName != "" {
				container = fmt.Sprintf("Container Name: %s\n", vContainerName)
			}

			if thisName == symbolName {
				return true
			}

			// Handle Dart LSP's function/method naming convention
			// Dart returns names like "functionName()" or "functionName(…)"
			if strings.HasPrefix(thisName, symbolName+"(") {
				return true
			}

			// Also handle the reverse - if user searches for "functionName()" but symbol is "functionName"
			if strings.HasPrefix(symbolName, thisName+"(") {
				return true
			}

			// Handle different matching strategies based on the search term
			if strings.Contains(symbolName, ".") {
				// For qualified names like "Type.Method", handle Dart's naming
				parts := strings.Split(symbolName, ".")
				if len(parts) == 2 {
					className := parts[0]
					methodName := parts[1]
					// Check if this is a method in the specified class
					if vContainerName == className &&
						(thisName == methodName || strings.HasPrefix(thisName, methodName+"(")) {
						return true
					}
				}

			} else if vKind == protocol.Method {
				// For methods, only match if the method name matches exactly Type.symbolName or Type::symbolName or symbolName
				if strings.HasSuffix(thisName, "::"+symbolName) || strings.HasSuffix(symbolName, "::"+thisName) {
					return true
				}

				if strings.HasSuffix(thisName, "."+symbolName) || strings.HasSuffix(symbolName, "."+thisName) {
					return true
				}
			}

			return false
		}

		switch v := symbol.(type) {
		case *protocol.SymbolInformation:
			if !doesSymbolMatch(v.Kind, v.ContainerName) {
				continue
			}

		case *protocol.WorkspaceSymbol:
			if !doesSymbolMatch(v.Kind, v.ContainerName) {
				continue
			}
		default:
			if symbol.GetName() != symbolName {
				continue
			}
		}

		toolsLogger.Debug("Found symbol: %s", symbol.GetName())
		loc := symbol.GetLocation()

		// Check if location has a valid URI
		if string(loc.URI) == "" {
			toolsLogger.Error("Symbol %s has empty URI, skipping", symbol.GetName())
			continue
		}

		err := client.OpenFile(ctx, loc.URI.Path())
		if err != nil {
			toolsLogger.Error("Error opening file: %v", err)
			continue
		}

		banner := "---\n\n"

		// Try to get the full definition
		// For Go, workspace/symbol returns just the symbol name range, so GetFullDefinition is needed
		// For Dart, workspace/symbol returns the full definition range
		definition := ""

		// Try GetFullDefinition first for languages like Go that need it
		def, newLoc, _, err := GetFullDefinition(ctx, client, loc)
		if err == nil {
			definition = def
			loc = newLoc
		} else {
			// Fall back to extracting text directly from the location
			// This works for Dart where workspace/symbol returns the full range
			definition, err = ExtractTextFromLocation(loc)
			if err != nil {
				toolsLogger.Error("Error getting definition: %v", err)
				continue
			}
		}

		locationInfo := fmt.Sprintf(
			"Symbol: %s\n"+
				"File: %s\n"+
				kind+
				container+
				"Range: L%d:C%d - L%d:C%d\n\n",
			symbol.GetName(),
			strings.TrimPrefix(string(loc.URI), "file://"),
			loc.Range.Start.Line+1,
			loc.Range.Start.Character+1,
			loc.Range.End.Line+1,
			loc.Range.End.Character+1,
		)

		definition = addLineNumbers(definition, int(loc.Range.Start.Line)+1)

		definitions = append(definitions, banner+locationInfo+definition+"\n")
	}

	if len(definitions) == 0 {
		return fmt.Sprintf("%s not found", symbolName), nil
	}

	return strings.Join(definitions, ""), nil
}
