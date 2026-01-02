package lsp

import "github.com/isaacphi/mcp-language-server/internal/protocol"

// HasDefinitionSupport checks if the server supports textDocument/definition
// AND workspace/symbol (both required by our definition tool implementation).
//
// The definition tool uses workspace/symbol to locate symbols by name (step 1),
// then textDocument/definition to retrieve the actual source code (step 2).
// Verified in internal/tools/definition.go:13-17.
//
// CRITICAL: Uses two-part check for Or_* types (pointer != nil && .Value != nil).
// See design doc Section 2.3 for Or_* type behavior.
func HasDefinitionSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.DefinitionProvider != nil &&
		caps.DefinitionProvider.Value != nil &&
		caps.WorkspaceSymbolProvider != nil &&
		caps.WorkspaceSymbolProvider.Value != nil
}

// HasReferencesSupport checks if the server supports textDocument/references.
//
// CRITICAL: Uses two-part check for Or_* type (pointer != nil && .Value != nil).
func HasReferencesSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.ReferencesProvider != nil &&
		caps.ReferencesProvider.Value != nil
}

// HasHoverSupport checks if the server supports textDocument/hover.
//
// CRITICAL: Uses two-part check for Or_* type (pointer != nil && .Value != nil).
func HasHoverSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.HoverProvider != nil &&
		caps.HoverProvider.Value != nil
}

// HasDocumentSymbolSupport checks if the server supports textDocument/documentSymbol.
//
// CRITICAL: Uses two-part check for Or_* type (pointer != nil && .Value != nil).
func HasDocumentSymbolSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.DocumentSymbolProvider != nil &&
		caps.DocumentSymbolProvider.Value != nil
}

// HasCallHierarchySupport checks if the server supports call hierarchy
// (textDocument/prepareCallHierarchy, callHierarchy/incomingCalls, callHierarchy/outgoingCalls).
//
// Call Hierarchy was added in LSP 3.16.0.
//
// CRITICAL: Uses two-part check for Or_* type (pointer != nil && .Value != nil).
func HasCallHierarchySupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.CallHierarchyProvider != nil &&
		caps.CallHierarchyProvider.Value != nil
}

// HasWorkspaceSymbolSupport checks if the server supports workspace/symbol.
//
// Used by definition tool as a dependency check.
//
// CRITICAL: Uses two-part check for Or_* type (pointer != nil && .Value != nil).
func HasWorkspaceSymbolSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.WorkspaceSymbolProvider != nil &&
		caps.WorkspaceSymbolProvider.Value != nil
}

// HasRenameSupport checks if the server supports textDocument/rename.
//
// RenameProvider is interface{} type - can be bool or RenameOptions.
// Simple nil check is sufficient (no .Value field to check).
func HasRenameSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.RenameProvider != nil
}

// HasCodeActionSupport checks if the server supports textDocument/codeAction.
//
// CodeActionProvider is interface{} type - can be bool or CodeActionOptions.
// Simple nil check is sufficient (no .Value field to check).
func HasCodeActionSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.CodeActionProvider != nil
}

// HasSignatureHelpSupport checks if the server supports textDocument/signatureHelp.
//
// SignatureHelpProvider is *SignatureHelpOptions type.
// Simple nil check is sufficient (pointer type, not Or_* type).
func HasSignatureHelpSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.SignatureHelpProvider != nil
}

// HasCodeLensSupport checks if the server supports textDocument/codeLens.
//
// CodeLensProvider is *CodeLensOptions type.
// Simple nil check is sufficient (pointer type, not Or_* type).
func HasCodeLensSupport(caps *protocol.ServerCapabilities) bool {
	if caps == nil {
		return false
	}
	return caps.CodeLensProvider != nil
}

// AlwaysSupported returns true for core tools that don't require capability checks.
//
// Core tools:
// - edit_file: Requires TextDocumentSync, which every LSP server must provide
// - diagnostics: Uses push notifications (textDocument/publishDiagnostics), not capability-based
//
// This function exists for documentation and consistency with the capability check pattern.
func AlwaysSupported() bool {
	return true
}
