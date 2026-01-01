package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

// GetCompletions returns context-aware code completion suggestions
// limit caps the number of results (default 20 if 0)
func GetCompletions(ctx context.Context, client *lsp.Client, filePath string, line, column, limit int) (string, error) {
	// Default limit
	if limit <= 0 {
		limit = 20
	}

	// Open the file if not already open
	err := client.OpenFile(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}

	// Create completion parameters
	params := protocol.CompletionParams{}

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

	// Execute the completion request
	completionResult, err := client.Completion(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to get completions: %v", err)
	}

	// Handle the response which can be CompletionList or []CompletionItem
	var items []protocol.CompletionItem
	var totalCount int

	// Extract items from completionResult
	if completionResult.Value == nil {
		return "No completions available", nil
	}

	switch v := completionResult.Value.(type) {
	case map[string]any:
		// This is a CompletionList
		if itemsRaw, ok := v["items"].([]any); ok {
			totalCount = len(itemsRaw)
			for _, itemRaw := range itemsRaw {
				if itemMap, ok := itemRaw.(map[string]any); ok {
					item := parseCompletionItem(itemMap)
					items = append(items, item)
				}
			}
		}
	case []any:
		// This is []CompletionItem
		totalCount = len(v)
		for _, itemRaw := range v {
			if itemMap, ok := itemRaw.(map[string]any); ok {
				item := parseCompletionItem(itemMap)
				items = append(items, item)
			}
		}
	default:
		return "", fmt.Errorf("unexpected completion result type: %T", v)
	}

	if len(items) == 0 {
		return "No completions available", nil
	}

	// Sort by SortText or Label
	sort.Slice(items, func(i, j int) bool {
		// Use SortText if available, otherwise use Label
		iSort := items[i].SortText
		if iSort == "" {
			iSort = items[i].Label
		}
		jSort := items[j].SortText
		if jSort == "" {
			jSort = items[j].Label
		}
		return iSort < jSort
	})

	// Limit results
	if len(items) > limit {
		items = items[:limit]
	}

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Completions (%d of %d):\n\n", len(items), totalCount))

	for i, item := range items {
		// Get kind string
		kindStr := getCompletionKindString(item.Kind)

		// Format: index. [Kind] Label
		output.WriteString(fmt.Sprintf("%d. [%s] %s", i+1, kindStr, item.Label))

		// Add detail if available (type information)
		if item.Detail != "" {
			output.WriteString(fmt.Sprintf("\n   Type: %s", item.Detail))
		}

		// Add truncated documentation if available
		if item.Documentation != nil {
			docStr := extractDocumentation(item.Documentation)
			if docStr != "" {
				// Truncate to first line or 100 chars
				lines := strings.Split(docStr, "\n")
				if len(lines) > 0 && lines[0] != "" {
					doc := lines[0]
					if len(doc) > 100 {
						doc = doc[:97] + "..."
					}
					output.WriteString(fmt.Sprintf("\n   Doc: %s", doc))
				}
			}
		}

		output.WriteString("\n\n")
	}

	return output.String(), nil
}

// parseCompletionItem converts a map[string]any to CompletionItem
func parseCompletionItem(itemMap map[string]any) protocol.CompletionItem {
	item := protocol.CompletionItem{}

	if label, ok := itemMap["label"].(string); ok {
		item.Label = label
	}

	if kind, ok := itemMap["kind"].(float64); ok {
		item.Kind = protocol.CompletionItemKind(kind)
	}

	if detail, ok := itemMap["detail"].(string); ok {
		item.Detail = detail
	}

	if sortText, ok := itemMap["sortText"].(string); ok {
		item.SortText = sortText
	}

	if filterText, ok := itemMap["filterText"].(string); ok {
		item.FilterText = filterText
	}

	// Documentation can be string or MarkupContent
	if doc, ok := itemMap["documentation"]; ok {
		item.Documentation = &protocol.Or_CompletionItem_documentation{Value: doc}
	}

	return item
}

// extractDocumentation extracts documentation string from Or_CompletionItem_documentation
func extractDocumentation(doc *protocol.Or_CompletionItem_documentation) string {
	if doc == nil || doc.Value == nil {
		return ""
	}

	switch v := doc.Value.(type) {
	case string:
		return v
	case map[string]any:
		// MarkupContent
		if value, ok := v["value"].(string); ok {
			return value
		}
	}

	return ""
}

// getCompletionKindString returns a human-readable string for CompletionItemKind
func getCompletionKindString(kind protocol.CompletionItemKind) string {
	switch kind {
	case protocol.TextCompletion:
		return "Text"
	case protocol.MethodCompletion:
		return "Method"
	case protocol.FunctionCompletion:
		return "Function"
	case protocol.ConstructorCompletion:
		return "Constructor"
	case protocol.FieldCompletion:
		return "Field"
	case protocol.VariableCompletion:
		return "Variable"
	case protocol.ClassCompletion:
		return "Class"
	case protocol.InterfaceCompletion:
		return "Interface"
	case protocol.ModuleCompletion:
		return "Module"
	case protocol.PropertyCompletion:
		return "Property"
	case protocol.UnitCompletion:
		return "Unit"
	case protocol.ValueCompletion:
		return "Value"
	case protocol.EnumCompletion:
		return "Enum"
	case protocol.KeywordCompletion:
		return "Keyword"
	case protocol.SnippetCompletion:
		return "Snippet"
	case protocol.ColorCompletion:
		return "Color"
	case protocol.FileCompletion:
		return "File"
	case protocol.ReferenceCompletion:
		return "Reference"
	default:
		return "Unknown"
	}
}
