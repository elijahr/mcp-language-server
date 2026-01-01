package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/isaacphi/mcp-language-server/internal/lsp"
	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

func ReadDefinition(ctx context.Context, client *lsp.Client, symbolName string) (string, error) {
	// First, use workspace/symbol to find where the symbol is referenced
	// This gives us a starting position to query for the definition
	symbolResult, err := client.Symbol(ctx, protocol.WorkspaceSymbolParams{
		Query: symbolName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch symbol: %v", err)
	}

	results, err := symbolResult.Results()
	if err != nil {
		return "", fmt.Errorf("failed to parse results: %v", err)
	}

	var definitions []string
	seenLocations := make(map[string]bool) // Track unique locations to avoid duplicates

	for _, symbol := range results {
		kind := ""
		container := ""

		// Skip symbols that we are not looking for. workspace/symbol may return
		// a large number of fuzzy matches.
		switch v := symbol.(type) {
		case *protocol.SymbolInformation:
			// SymbolInformation results have richer data.
			kind = fmt.Sprintf("Kind: %s\n", protocol.TableKindMap[v.Kind])
			if v.ContainerName != "" {
				container = fmt.Sprintf("Container Name: %s\n", v.ContainerName)
			}

			// Check if this symbol matches what we're looking for
			if !symbolMatches(symbolName, symbol.GetName(), v.Kind, v.ContainerName) {
				continue
			}
		default:
			// For generic symbols without type information, use basic matching
			if !symbolMatches(symbolName, symbol.GetName(), 0, "") {
				continue
			}
		}

		toolsLogger.Debug("Found symbol: %s", symbol.GetName())
		loc := symbol.GetLocation()

		// Open the file containing the symbol
		err := client.OpenFile(ctx, loc.URI.Path())
		if err != nil {
			toolsLogger.Error("Error opening file: %v", err)
			continue
		}

		// Use textDocument/definition to get the actual definition location
		defParams := protocol.DefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: loc.URI,
				},
				Position: loc.Range.Start,
			},
		}

		defResult, err := client.Definition(ctx, defParams)
		if err != nil {
			toolsLogger.Error("Error getting definition: %v", err)
			continue
		}

		// Extract locations from the definition result
		defLocations, err := extractDefinitionLocations(defResult)
		if err != nil {
			toolsLogger.Error("Error extracting definition locations: %v", err)
			continue
		}

		// Process each definition location
		for _, defLoc := range defLocations {
			// Create unique key for this location to avoid duplicates
			locationKey := fmt.Sprintf("%s:%d:%d", defLoc.URI, defLoc.Range.Start.Line, defLoc.Range.Start.Character)
			if seenLocations[locationKey] {
				continue
			}
			seenLocations[locationKey] = true

			// Open the file containing the definition
			err := client.OpenFile(ctx, defLoc.URI.Path())
			if err != nil {
				toolsLogger.Error("Error opening file for definition: %v", err)
				continue
			}

			banner := "---\n\n"
			definition, finalLoc, err := GetFullDefinition(ctx, client, defLoc)
			locationInfo := fmt.Sprintf(
				"Symbol: %s\n"+
					"File: %s\n"+
					kind+
					container+
					"Range: L%d:C%d - L%d:C%d\n\n",
				symbol.GetName(),
				strings.TrimPrefix(string(finalLoc.URI), "file://"),
				finalLoc.Range.Start.Line+1,
				finalLoc.Range.Start.Character+1,
				finalLoc.Range.End.Line+1,
				finalLoc.Range.End.Character+1,
			)

			if err != nil {
				toolsLogger.Error("Error getting full definition: %v", err)
				continue
			}

			definition = addLineNumbers(definition, int(finalLoc.Range.Start.Line)+1)
			definitions = append(definitions, banner+locationInfo+definition+"\n")
		}
	}

	if len(definitions) == 0 {
		return fmt.Sprintf("%s not found", symbolName), nil
	}

	return strings.Join(definitions, ""), nil
}

// extractDefinitionLocations extracts Location objects from a Definition result
// which can be: Location, []Location, Definition, or []DefinitionLink
func extractDefinitionLocations(defResult protocol.Or_Result_textDocument_definition) ([]protocol.Location, error) {
	if defResult.Value == nil {
		return nil, fmt.Errorf("no definition found")
	}

	var locations []protocol.Location

	switch v := defResult.Value.(type) {
	case protocol.Definition:
		// Definition is an Or_Definition which can be Location or []Location
		switch def := v.Value.(type) {
		case protocol.Location:
			locations = append(locations, def)
		case []protocol.Location:
			locations = append(locations, def...)
		default:
			return nil, fmt.Errorf("unexpected Definition type: %T", def)
		}
	case []protocol.DefinitionLink:
		// Convert DefinitionLink to Location (use TargetUri and TargetRange)
		for _, link := range v {
			locations = append(locations, protocol.Location{
				URI:   link.TargetURI,
				Range: link.TargetRange,
			})
		}
	default:
		return nil, fmt.Errorf("unexpected definition result type: %T", v)
	}

	return locations, nil
}

// symbolMatches determines if a symbol matches the search query.
// It handles various language conventions including C++ (::), Go/TypeScript (.), and plain names.
func symbolMatches(query, symbolName string, symbolKind protocol.SymbolKind, containerName string) bool {
	// Exact match is always accepted
	if symbolName == query {
		return true
	}

	// Handle qualified names (e.g., "Type.method" or "Type::method")
	if strings.ContainsAny(query, ".:") {
		// Query has explicit qualification - require exact match
		return symbolName == query
	}

	// For unqualified queries, we need to be more flexible

	// 1. Check if symbol name ends with the query (handles Type::method, Type.method, etc.)
	//    This catches C++ qualified names like "TestClass::method" when searching for "method"
	if strings.HasSuffix(symbolName, "::"+query) || strings.HasSuffix(symbolName, "."+query) {
		return true
	}

	// 2. For methods specifically, also check container name matching
	//    clangd may return "method" as the symbol name with "TestClass" as container
	if symbolKind == protocol.Method || symbolKind == protocol.Function {
		// If the symbol name matches and has a container, it's likely what we want
		if symbolName == query && containerName != "" {
			return true
		}
	}

	// 3. Fuzzy matching for classes, structs, types, etc.
	//    Some LSP servers return slightly different names
	if symbolKind == protocol.Class || symbolKind == protocol.Struct ||
	   symbolKind == protocol.Interface || symbolKind == protocol.Enum {
		// Try case-insensitive match
		if strings.EqualFold(symbolName, query) {
			return true
		}

		// Check if symbol name contains the query (substring match)
		// This helps with templates like "std::vector<int>" when searching for "vector"
		if strings.Contains(symbolName, query) {
			return true
		}
	}

	// 4. For variables and constants, be strict but allow for namespace/scope prefixes
	if symbolKind == protocol.Variable || symbolKind == protocol.Constant {
		// Allow namespace::CONSTANT or Scope::variable patterns
		parts := strings.Split(symbolName, "::")
		if len(parts) > 1 && parts[len(parts)-1] == query {
			return true
		}
	}

	return false
}
