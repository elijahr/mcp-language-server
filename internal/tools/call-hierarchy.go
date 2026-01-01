package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

// GetCallHierarchy returns incoming or outgoing calls for a symbol at the given position
// direction should be "incoming" or "outgoing"
func GetCallHierarchy(ctx context.Context, client *lsp.Client, filePath string, line, column int, direction string) (string, error) {
	// Validate direction parameter
	if direction != "incoming" && direction != "outgoing" {
		return "", fmt.Errorf("direction must be 'incoming' or 'outgoing', got: %s", direction)
	}

	// Open the file first
	err := client.OpenFile(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	// Create URI from file path
	uri := protocol.DocumentUri(fmt.Sprintf("file://%s", filePath))

	// Create CallHierarchyPrepareParams (similar to DefinitionParams)
	params := protocol.CallHierarchyPrepareParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri,
			},
			Position: protocol.Position{
				Line:      uint32(line - 1),      // Convert from 1-indexed to 0-indexed
				Character: uint32(column - 1),    // Convert from 1-indexed to 0-indexed
			},
		},
	}

	// Call PrepareCallHierarchy to get CallHierarchyItem
	items, err := client.PrepareCallHierarchy(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to prepare call hierarchy: %w", err)
	}

	// Check if we found any symbol at the position
	if len(items) == 0 {
		return fmt.Sprintf("No symbol found at %s:%d:%d", filePath, line, column), nil
	}

	// Take the first item (most relevant)
	item := items[0]

	// Get calls based on direction
	var result strings.Builder

	if direction == "incoming" {
		// Get incoming calls (callers)
		incomingParams := protocol.CallHierarchyIncomingCallsParams{
			Item: item,
		}

		incomingCalls, err := client.IncomingCalls(ctx, incomingParams)
		if err != nil {
			return "", fmt.Errorf("failed to get incoming calls: %w", err)
		}

		// Format the output
		result.WriteString(fmt.Sprintf("Incoming calls to: %s", item.Name))
		if item.Detail != "" {
			result.WriteString(fmt.Sprintf(" (%s)", item.Detail))
		}
		result.WriteString(fmt.Sprintf(" at %s:%d\n\n",
			strings.TrimPrefix(string(item.URI), "file://"),
			item.Range.Start.Line+1))

		if len(incomingCalls) == 0 {
			result.WriteString("No incoming calls found\n")
		} else {
			for i, call := range incomingCalls {
				// Format caller information
				callerFile := strings.TrimPrefix(string(call.From.URI), "file://")
				callerLine := call.From.Range.Start.Line + 1

				result.WriteString(fmt.Sprintf("%d. %s", i+1, call.From.Name))
				if call.From.Detail != "" {
					result.WriteString(fmt.Sprintf(" (%s)", call.From.Detail))
				}
				result.WriteString(fmt.Sprintf(" at %s:%d\n",
					callerFile,
					callerLine))

				// Show the ranges where the calls occur
				if len(call.FromRanges) > 0 {
					var ranges []string
					for _, r := range call.FromRanges {
						ranges = append(ranges, fmt.Sprintf("L%d:C%d",
							r.Start.Line+1,
							r.Start.Character+1))
					}
					result.WriteString(fmt.Sprintf("   Call sites: %s\n", strings.Join(ranges, ", ")))
				}
			}
		}
	} else {
		// Get outgoing calls (callees)
		outgoingParams := protocol.CallHierarchyOutgoingCallsParams{
			Item: item,
		}

		outgoingCalls, err := client.OutgoingCalls(ctx, outgoingParams)
		if err != nil {
			return "", fmt.Errorf("failed to get outgoing calls: %w", err)
		}

		// Format the output
		result.WriteString(fmt.Sprintf("Outgoing calls from: %s", item.Name))
		if item.Detail != "" {
			result.WriteString(fmt.Sprintf(" (%s)", item.Detail))
		}
		result.WriteString(fmt.Sprintf(" at %s:%d\n\n",
			strings.TrimPrefix(string(item.URI), "file://"),
			item.Range.Start.Line+1))

		if len(outgoingCalls) == 0 {
			result.WriteString("No outgoing calls found\n")
		} else {
			for i, call := range outgoingCalls {
				// Format callee information
				calleeFile := strings.TrimPrefix(string(call.To.URI), "file://")
				calleeLine := call.To.Range.Start.Line + 1

				result.WriteString(fmt.Sprintf("%d. %s", i+1, call.To.Name))
				if call.To.Detail != "" {
					result.WriteString(fmt.Sprintf(" (%s)", call.To.Detail))
				}
				result.WriteString(fmt.Sprintf(" at %s:%d\n",
					calleeFile,
					calleeLine))

				// Show the ranges where the calls are made from
				if len(call.FromRanges) > 0 {
					var ranges []string
					for _, r := range call.FromRanges {
						ranges = append(ranges, fmt.Sprintf("L%d:C%d",
							r.Start.Line+1,
							r.Start.Character+1))
					}
					result.WriteString(fmt.Sprintf("   Called at: %s\n", strings.Join(ranges, ", ")))
				}
			}
		}
	}

	return result.String(), nil
}
