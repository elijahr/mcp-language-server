package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

// GetCodeActions returns available code actions for a range in a file
func GetCodeActions(ctx context.Context, client *lsp.Client, filePath string, startLine, startColumn, endLine, endColumn int) (string, error) {
	// Open the file if not already open
	err := client.OpenFile(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}

	// Convert to URI format
	uri := protocol.DocumentUri("file://" + filePath)

	// Get diagnostics for the file to include in CodeActionContext
	diagnostics := client.GetFileDiagnostics(uri)

	// Create the range (convert 1-indexed to 0-indexed)
	actionRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(startLine - 1),
			Character: uint32(startColumn - 1),
		},
		End: protocol.Position{
			Line:      uint32(endLine - 1),
			Character: uint32(endColumn - 1),
		},
	}

	// Create CodeActionParams
	params := protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri,
		},
		Range: actionRange,
		Context: protocol.CodeActionContext{
			Diagnostics: diagnostics,
		},
	}

	// Call the CodeAction method
	actions, err := client.CodeAction(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to get code actions: %v", err)
	}

	if len(actions) == 0 {
		return "No code actions available", nil
	}

	// Format the code actions
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Code Actions (%d available):\n\n", len(actions)))

	for i, actionItem := range actions {
		// The Value field contains either a CodeAction or Command
		if actionItem.Value == nil {
			continue
		}

		// Type assert the value
		switch v := actionItem.Value.(type) {
		case map[string]any:
			// This is a CodeAction
			title, _ := v["title"].(string)
			kindStr, _ := v["kind"].(string)

			kind := "Unknown"
			if kindStr != "" {
				kind = formatCodeActionKind(kindStr)
			}

			result.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, kind, title))

			// Add command if present
			if cmdMap, ok := v["command"].(map[string]any); ok {
				if cmdName, ok := cmdMap["command"].(string); ok {
					result.WriteString(fmt.Sprintf("   Command: %s\n", cmdName))
				}
			}

		default:
			// Unknown type, try to extract what we can
			result.WriteString(fmt.Sprintf("%d. Unknown action type\n", i+1))
		}

		// Add blank line between actions
		if i < len(actions)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// formatCodeActionKind converts a CodeActionKind string into a more readable format
func formatCodeActionKind(kind string) string {
	switch {
	case strings.Contains(kind, "quickfix"):
		return "QuickFix"
	case strings.Contains(kind, "refactor.extract"):
		return "Refactor.Extract"
	case strings.Contains(kind, "refactor.inline"):
		return "Refactor.Inline"
	case strings.Contains(kind, "refactor.rewrite"):
		return "Refactor.Rewrite"
	case strings.Contains(kind, "refactor"):
		return "Refactor"
	case strings.Contains(kind, "source.organizeImports"):
		return "Source.OrganizeImports"
	case strings.Contains(kind, "source.fixAll"):
		return "Source.FixAll"
	case strings.Contains(kind, "source"):
		return "Source"
	default:
		return kind
	}
}
