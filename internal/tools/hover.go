package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

// GetHoverInfo retrieves hover information (type, documentation) for a symbol at the specified position
func GetHoverInfo(ctx context.Context, client *lsp.Client, filePath string, line, column int) (string, error) {
	// Open the file if not already open
	err := client.OpenFile(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}

	params := protocol.HoverParams{}

	// Convert 1-indexed line/column to 0-indexed for LSP protocol
	position := protocol.Position{
		Line:      uint32(line - 1),
		Character: uint32(column - 1),
	}
	uri := protocol.DocumentUri("file://" + filePath)
	params.TextDocument = protocol.TextDocumentIdentifier{
		URI: uri,
	}
	params.Position = position

	// Execute the hover request
	// - some LSP (rust) will return "content modified", so retry it
	var hoverResult protocol.Hover
	for i := range 3 {
		hoverResult, err = client.Hover(ctx, params)
		if err == nil {
			break
		} else if errors.Is(err, lsp.ErrContentModified) {
			continue
		} else if i == 2 {
			return "", fmt.Errorf("failed to get hover information: %v", err)
		}
	}

	var result strings.Builder

	// Process the hover contents based on the union type
	if hoverResult.Contents.Value == nil {
		// Extract the line where the hover was requested
		lineText, err := ExtractTextFromLocation(protocol.Location{
			URI: uri,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      position.Line,
					Character: 0,
				},
				End: protocol.Position{
					Line:      position.Line + 1,
					Character: 0,
				},
			},
		})
		if err != nil {
			toolsLogger.Warn("failed to extract line at position: %v", err)
		}
		result.WriteString(fmt.Sprintf("No hover information available for this position on the following line:\n%s", lineText))
	} else {
		// Handle the different types that contents can be
		switch v := hoverResult.Contents.Value.(type) {
		case protocol.MarkupContent:
			result.WriteString(v.Value)
		case protocol.MarkedString:
			// MarkedString is a union type itself (string | {language, value})
			switch ms := v.Value.(type) {
			case string:
				result.WriteString(ms)
			case protocol.MarkedStringWithLanguage:
				// Format as markdown code block
				result.WriteString(fmt.Sprintf("```%s\n%s\n```", ms.Language, ms.Value))
			default:
				result.WriteString(fmt.Sprintf("%v", ms))
			}
		case []protocol.MarkedString:
			// Multiple marked strings - concatenate them
			for i, ms := range v {
				if i > 0 {
					result.WriteString("\n")
				}
				switch msv := ms.Value.(type) {
				case string:
					result.WriteString(msv)
				case protocol.MarkedStringWithLanguage:
					result.WriteString(fmt.Sprintf("```%s\n%s\n```", msv.Language, msv.Value))
				default:
					result.WriteString(fmt.Sprintf("%v", msv))
				}
			}
		default:
			// Fallback for unexpected types
			result.WriteString(fmt.Sprintf("%v", v))
		}
	}

	return result.String(), nil
}
