package main

import (
	"context"
	"fmt"

	"github.com/isaacphi/mcp-language-server/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *mcpServer) registerTools() error {
	coreLogger.Debug("Registering MCP tools")

	applyTextEditTool := mcp.NewTool("edit_file",
		mcp.WithDescription("Apply multiple text edits to a file."),
		mcp.WithArray("edits",
			mcp.Required(),
			mcp.Description("List of edits to apply"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"startLine": map[string]any{
						"type":        "number",
						"description": "Start line to replace, inclusive, one-indexed",
					},
					"endLine": map[string]any{
						"type":        "number",
						"description": "End line to replace, inclusive, one-indexed",
					},
					"newText": map[string]any{
						"type":        "string",
						"description": "Replacement text. Replace with the new text. Leave blank to remove lines.",
					},
				},
				"required": []string{"startLine", "endLine"},
			}),
		),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("Path to the file to edit"),
		),
	)

	s.mcpServer.AddTool(applyTextEditTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		// Extract edits array
		editsArg, ok := request.Params.Arguments["edits"]
		if !ok {
			return mcp.NewToolResultError("edits is required"), nil
		}

		// Type assert and convert the edits
		editsArray, ok := editsArg.([]any)
		if !ok {
			return mcp.NewToolResultError("edits must be an array"), nil
		}

		var edits []tools.TextEdit
		for _, editItem := range editsArray {
			editMap, ok := editItem.(map[string]any)
			if !ok {
				return mcp.NewToolResultError("each edit must be an object"), nil
			}

			startLine, ok := editMap["startLine"].(float64)
			if !ok {
				return mcp.NewToolResultError("startLine must be a number"), nil
			}

			endLine, ok := editMap["endLine"].(float64)
			if !ok {
				return mcp.NewToolResultError("endLine must be a number"), nil
			}

			newText, _ := editMap["newText"].(string) // newText can be empty

			edits = append(edits, tools.TextEdit{
				StartLine: int(startLine),
				EndLine:   int(endLine),
				NewText:   newText,
			})
		}

		coreLogger.Debug("Executing edit_file for file: %s", filePath)
		response, err := tools.ApplyTextEdits(s.ctx, s.lspClient, filePath, edits)
		if err != nil {
			coreLogger.Error("Failed to apply edits: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to apply edits: %v", err)), nil
		}
		return mcp.NewToolResultText(response), nil
	})

	readDefinitionTool := mcp.NewTool("definition",
		mcp.WithDescription("Read the source code definition of a symbol (function, type, constant, etc.) from the codebase. Returns the complete implementation code where the symbol is defined."),
		mcp.WithString("symbolName",
			mcp.Required(),
			mcp.Description("The name of the symbol whose definition you want to find (e.g. 'mypackage.MyFunction', 'MyType.MyMethod')"),
		),
	)

	s.mcpServer.AddTool(readDefinitionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		symbolName, ok := request.Params.Arguments["symbolName"].(string)
		if !ok {
			return mcp.NewToolResultError("symbolName must be a string"), nil
		}

		coreLogger.Debug("Executing definition for symbol: %s", symbolName)
		text, err := tools.ReadDefinition(s.ctx, s.lspClient, symbolName)
		if err != nil {
			coreLogger.Error("Failed to get definition: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get definition: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	findReferencesTool := mcp.NewTool("references",
		mcp.WithDescription("Find all usages and references of a symbol throughout the codebase. Returns a list of all files and locations where the symbol appears."),
		mcp.WithString("symbolName",
			mcp.Required(),
			mcp.Description("The name of the symbol to search for (e.g. 'mypackage.MyFunction', 'MyType')"),
		),
	)

	s.mcpServer.AddTool(findReferencesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		symbolName, ok := request.Params.Arguments["symbolName"].(string)
		if !ok {
			return mcp.NewToolResultError("symbolName must be a string"), nil
		}

		coreLogger.Debug("Executing references for symbol: %s", symbolName)
		text, err := tools.FindReferences(s.ctx, s.lspClient, symbolName)
		if err != nil {
			coreLogger.Error("Failed to find references: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to find references: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	getDiagnosticsTool := mcp.NewTool("diagnostics",
		mcp.WithDescription("Get diagnostic information for a specific file from the language server."),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("The path to the file to get diagnostics for"),
		),
		mcp.WithBoolean("contextLines",
			mcp.Description("Lines to include around each diagnostic."),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("showLineNumbers",
			mcp.Description("If true, adds line numbers to the output"),
			mcp.DefaultBool(true),
		),
	)

	s.mcpServer.AddTool(getDiagnosticsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		contextLines := 5 // default value
		if contextLinesArg, ok := request.Params.Arguments["contextLines"].(int); ok {
			contextLines = contextLinesArg
		}

		showLineNumbers := true // default value
		if showLineNumbersArg, ok := request.Params.Arguments["showLineNumbers"].(bool); ok {
			showLineNumbers = showLineNumbersArg
		}

		coreLogger.Debug("Executing diagnostics for file: %s", filePath)
		text, err := tools.GetDiagnosticsForFile(s.ctx, s.lspClient, filePath, contextLines, showLineNumbers)
		if err != nil {
			coreLogger.Error("Failed to get diagnostics: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get diagnostics: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	getCodeLensTool := mcp.NewTool("get_codelens",
		mcp.WithDescription("Get code lens hints for a given file from the language server."),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("The path to the file to get code lens information for"),
		),
	)

	s.mcpServer.AddTool(getCodeLensTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		coreLogger.Debug("Executing get_codelens for file: %s", filePath)
		text, err := tools.GetCodeLens(s.ctx, s.lspClient, filePath)
		if err != nil {
			coreLogger.Error("Failed to get code lens: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get code lens: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	executeCodeLensTool := mcp.NewTool("execute_codelens",
		mcp.WithDescription("Execute a code lens command for a given file and lens index."),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("The path to the file containing the code lens to execute"),
		),
		mcp.WithNumber("index",
			mcp.Required(),
			mcp.Description("The index of the code lens to execute (from get_codelens output), 1 indexed"),
		),
	)

	s.mcpServer.AddTool(executeCodeLensTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		// Handle both float64 and int for index due to JSON parsing
		var index int
		switch v := request.Params.Arguments["index"].(type) {
		case float64:
			index = int(v)
		case int:
			index = v
		default:
			return mcp.NewToolResultError("index must be a number"), nil
		}

		coreLogger.Debug("Executing execute_codelens for file: %s index: %d", filePath, index)
		text, err := tools.ExecuteCodeLens(s.ctx, s.lspClient, filePath, index)
		if err != nil {
			coreLogger.Error("Failed to execute code lens: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to execute code lens: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	hoverTool := mcp.NewTool("hover",
		mcp.WithDescription("Get hover information (type, documentation) for a symbol at the specified position."),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("The path to the file to get hover information for"),
		),
		mcp.WithNumber("line",
			mcp.Required(),
			mcp.Description("The line number where the hover is requested (1-indexed)"),
		),
		mcp.WithNumber("column",
			mcp.Required(),
			mcp.Description("The column number where the hover is requested (1-indexed)"),
		),
	)

	s.mcpServer.AddTool(hoverTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		// Handle both float64 and int for line and column due to JSON parsing
		var line, column int
		switch v := request.Params.Arguments["line"].(type) {
		case float64:
			line = int(v)
		case int:
			line = v
		default:
			return mcp.NewToolResultError("line must be a number"), nil
		}

		switch v := request.Params.Arguments["column"].(type) {
		case float64:
			column = int(v)
		case int:
			column = v
		default:
			return mcp.NewToolResultError("column must be a number"), nil
		}

		coreLogger.Debug("Executing hover for file: %s line: %d column: %d", filePath, line, column)
		text, err := tools.GetHoverInfo(s.ctx, s.lspClient, filePath, line, column)
		if err != nil {
			coreLogger.Error("Failed to get hover information: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get hover information: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	renameSymbolTool := mcp.NewTool("rename_symbol",
		mcp.WithDescription("Rename a symbol (variable, function, class, etc.) at the specified position and update all references throughout the codebase."),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("The path to the file containing the symbol to rename"),
		),
		mcp.WithNumber("line",
			mcp.Required(),
			mcp.Description("The line number where the symbol is located (1-indexed)"),
		),
		mcp.WithNumber("column",
			mcp.Required(),
			mcp.Description("The column number where the symbol is located (1-indexed)"),
		),
		mcp.WithString("newName",
			mcp.Required(),
			mcp.Description("The new name for the symbol"),
		),
	)

	s.mcpServer.AddTool(renameSymbolTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		newName, ok := request.Params.Arguments["newName"].(string)
		if !ok {
			return mcp.NewToolResultError("newName must be a string"), nil
		}

		// Handle both float64 and int for line and column due to JSON parsing
		var line, column int
		switch v := request.Params.Arguments["line"].(type) {
		case float64:
			line = int(v)
		case int:
			line = v
		default:
			return mcp.NewToolResultError("line must be a number"), nil
		}

		switch v := request.Params.Arguments["column"].(type) {
		case float64:
			column = int(v)
		case int:
			column = v
		default:
			return mcp.NewToolResultError("column must be a number"), nil
		}

		coreLogger.Debug("Executing rename_symbol for file: %s line: %d column: %d newName: %s", filePath, line, column, newName)
		text, err := tools.RenameSymbol(s.ctx, s.lspClient, filePath, line, column, newName)
		if err != nil {
			coreLogger.Error("Failed to rename symbol: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to rename symbol: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	codeActionsTool := mcp.NewTool("code_actions",
		mcp.WithDescription("Get available code actions (quick fixes, refactorings) for a range"),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("startLine",
			mcp.Required(),
			mcp.Description("Start line (1-indexed)"),
		),
		mcp.WithNumber("startColumn",
			mcp.Required(),
			mcp.Description("Start column (1-indexed)"),
		),
		mcp.WithNumber("endLine",
			mcp.Required(),
			mcp.Description("End line (1-indexed)"),
		),
		mcp.WithNumber("endColumn",
			mcp.Required(),
			mcp.Description("End column (1-indexed)"),
		),
	)

	s.mcpServer.AddTool(codeActionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		// Handle both float64 and int for all numeric parameters due to JSON parsing
		var startLine, startColumn, endLine, endColumn int

		switch v := request.Params.Arguments["startLine"].(type) {
		case float64:
			startLine = int(v)
		case int:
			startLine = v
		default:
			return mcp.NewToolResultError("startLine must be a number"), nil
		}

		switch v := request.Params.Arguments["startColumn"].(type) {
		case float64:
			startColumn = int(v)
		case int:
			startColumn = v
		default:
			return mcp.NewToolResultError("startColumn must be a number"), nil
		}

		switch v := request.Params.Arguments["endLine"].(type) {
		case float64:
			endLine = int(v)
		case int:
			endLine = v
		default:
			return mcp.NewToolResultError("endLine must be a number"), nil
		}

		switch v := request.Params.Arguments["endColumn"].(type) {
		case float64:
			endColumn = int(v)
		case int:
			endColumn = v
		default:
			return mcp.NewToolResultError("endColumn must be a number"), nil
		}

		coreLogger.Debug("Executing code_actions for file: %s range: (%d,%d) to (%d,%d)", filePath, startLine, startColumn, endLine, endColumn)
		text, err := tools.GetCodeActions(s.ctx, s.lspClient, filePath, startLine, startColumn, endLine, endColumn)
		if err != nil {
			coreLogger.Error("Failed to get code actions: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get code actions: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	signatureHelpTool := mcp.NewTool("signature_help",
		mcp.WithDescription("Get function/method signature information at cursor position"),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("Path to the file"),
		),
		mcp.WithNumber("line",
			mcp.Required(),
			mcp.Description("Line number (1-indexed)"),
		),
		mcp.WithNumber("column",
			mcp.Required(),
			mcp.Description("Column number (1-indexed)"),
		),
	)

	s.mcpServer.AddTool(signatureHelpTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		// Handle both float64 and int for line and column due to JSON parsing
		var line, column int
		switch v := request.Params.Arguments["line"].(type) {
		case float64:
			line = int(v)
		case int:
			line = v
		default:
			return mcp.NewToolResultError("line must be a number"), nil
		}

		switch v := request.Params.Arguments["column"].(type) {
		case float64:
			column = int(v)
		case int:
			column = v
		default:
			return mcp.NewToolResultError("column must be a number"), nil
		}

		coreLogger.Debug("Executing signature_help for file: %s line: %d column: %d", filePath, line, column)
		text, err := tools.GetSignatureHelp(s.ctx, s.lspClient, filePath, line, column)
		if err != nil {
			coreLogger.Error("Failed to get signature help: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get signature help: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	documentSymbolsTool := mcp.NewTool("document_symbols",
		mcp.WithDescription("Get the hierarchical symbol outline of a file (classes, functions, methods, etc.)"),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("Path to the file to get symbols for"),
		),
	)

	s.mcpServer.AddTool(documentSymbolsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		coreLogger.Debug("Executing document_symbols for file: %s", filePath)
		text, err := tools.GetDocumentSymbols(s.ctx, s.lspClient, filePath)
		if err != nil {
			coreLogger.Error("Failed to get document symbols: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get document symbols: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})


	callHierarchyTool := mcp.NewTool("call_hierarchy",
		mcp.WithDescription("Find incoming callers or outgoing callees for a symbol at the specified position."),
		mcp.WithString("filePath",
			mcp.Required(),
			mcp.Description("Path to the file containing the symbol"),
		),
		mcp.WithNumber("line",
			mcp.Required(),
			mcp.Description("Line number (1-indexed)"),
		),
		mcp.WithNumber("column",
			mcp.Required(),
			mcp.Description("Column number (1-indexed)"),
		),
		mcp.WithString("direction",
			mcp.Required(),
			mcp.Description("'incoming' for callers or 'outgoing' for callees"),
		),
	)

	s.mcpServer.AddTool(callHierarchyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		filePath, ok := request.Params.Arguments["filePath"].(string)
		if !ok {
			return mcp.NewToolResultError("filePath must be a string"), nil
		}

		direction, ok := request.Params.Arguments["direction"].(string)
		if !ok {
			return mcp.NewToolResultError("direction must be a string"), nil
		}

		// Validate direction
		if direction != "incoming" && direction != "outgoing" {
			return mcp.NewToolResultError("direction must be 'incoming' or 'outgoing'"), nil
		}

		// Handle both float64 and int for line and column due to JSON parsing
		var line, column int
		switch v := request.Params.Arguments["line"].(type) {
		case float64:
			line = int(v)
		case int:
			line = v
		default:
			return mcp.NewToolResultError("line must be a number"), nil
		}

		switch v := request.Params.Arguments["column"].(type) {
		case float64:
			column = int(v)
		case int:
			column = v
		default:
			return mcp.NewToolResultError("column must be a number"), nil
		}

		coreLogger.Debug("Executing call_hierarchy for file: %s line: %d column: %d direction: %s", filePath, line, column, direction)
		text, err := tools.GetCallHierarchy(s.ctx, s.lspClient, filePath, line, column, direction)
		if err != nil {
			coreLogger.Error("Failed to get call hierarchy: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to get call hierarchy: %v", err)), nil
		}
		return mcp.NewToolResultText(text), nil
	})
	coreLogger.Info("Successfully registered all MCP tools")
	return nil
}
