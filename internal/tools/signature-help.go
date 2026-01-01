package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

// GetSignatureHelp returns function signature information at the given position
func GetSignatureHelp(ctx context.Context, client *lsp.Client, filePath string, line, column int) (string, error) {
	// Open the file if not already open
	err := client.OpenFile(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}

	params := protocol.SignatureHelpParams{}

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

	// Execute the signature help request
	signatureResult, err := client.SignatureHelp(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to get signature help: %v", err)
	}

	var result strings.Builder

	// Check if we have any signatures
	if len(signatureResult.Signatures) == 0 {
		// Extract the line where the signature help was requested
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
		result.WriteString(fmt.Sprintf("No signature help available for this position on the following line:\n%s", lineText))
		return result.String(), nil
	}

	// Determine active signature index
	activeSignatureIdx := int(signatureResult.ActiveSignature)
	if activeSignatureIdx >= len(signatureResult.Signatures) {
		activeSignatureIdx = 0
	}

	// Determine active parameter index
	activeParameterIdx := int(signatureResult.ActiveParameter)

	result.WriteString("Signature Help:\n\n")

	// Display all signatures
	for i, sig := range signatureResult.Signatures {
		prefix := "  "
		if i == activeSignatureIdx {
			prefix = "▶ "
		}
		result.WriteString(fmt.Sprintf("%s%s\n", prefix, sig.Label))

		// For the active signature, show detailed parameter information
		if i == activeSignatureIdx {
			if len(sig.Parameters) > 0 {
				result.WriteString("\nParameters:\n")
				for j, param := range sig.Parameters {
					// Get parameter label
					var paramLabel string
					switch v := param.Label.Value.(type) {
					case string:
						paramLabel = v
					case protocol.Tuple_ParameterInformation_label_Item1:
						// Extract label from signature using offsets
						start := v.Fld0
						end := v.Fld1
						if int(start) < len(sig.Label) && int(end) <= len(sig.Label) {
							paramLabel = sig.Label[start:end]
						}
					}

					// Mark active parameter
					activeMarker := " "
					if j == activeParameterIdx {
						activeMarker = "▶"
					}

					// Get parameter documentation if available
					paramDoc := ""
					if param.Documentation != nil {
						switch v := param.Documentation.Value.(type) {
						case string:
							paramDoc = fmt.Sprintf(" - %s", v)
						case protocol.MarkupContent:
							paramDoc = fmt.Sprintf(" - %s", v.Value)
						}
					}

					result.WriteString(fmt.Sprintf("  %s %s%s\n", activeMarker, paramLabel, paramDoc))
				}
			}

			// Show signature documentation if available
			if sig.Documentation != nil {
				result.WriteString("\nDocumentation:\n")
				switch v := sig.Documentation.Value.(type) {
				case string:
					result.WriteString(fmt.Sprintf("%s\n", v))
				case protocol.MarkupContent:
					result.WriteString(fmt.Sprintf("%s\n", v.Value))
				}
			}
		}
	}

	if len(signatureResult.Signatures) > 1 {
		result.WriteString(fmt.Sprintf("\nShowing %d of %d signatures (▶ marks active signature/parameter)\n",
			activeSignatureIdx+1, len(signatureResult.Signatures)))
	}

	return result.String(), nil
}
