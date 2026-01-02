package lsp

import (
	"testing"

	"github.com/isaacphi/mcp-language-server/internal/protocol"
)

func TestHasDefinitionSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "both capabilities present with non-nil Values",
			caps: &protocol.ServerCapabilities{
				DefinitionProvider: &protocol.Or_ServerCapabilities_definitionProvider{
					Value: true,
				},
				WorkspaceSymbolProvider: &protocol.Or_ServerCapabilities_workspaceSymbolProvider{
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "definition present, workspace symbol missing",
			caps: &protocol.ServerCapabilities{
				DefinitionProvider: &protocol.Or_ServerCapabilities_definitionProvider{
					Value: true,
				},
				WorkspaceSymbolProvider: nil,
			},
			expected: false,
		},
		{
			name: "definition pointer non-nil but Value is nil (unsupported)",
			caps: &protocol.ServerCapabilities{
				DefinitionProvider: &protocol.Or_ServerCapabilities_definitionProvider{
					Value: nil,
				},
				WorkspaceSymbolProvider: &protocol.Or_ServerCapabilities_workspaceSymbolProvider{
					Value: true,
				},
			},
			expected: false,
		},
		{
			name: "both present but values are nil",
			caps: &protocol.ServerCapabilities{
				DefinitionProvider: &protocol.Or_ServerCapabilities_definitionProvider{
					Value: nil,
				},
				WorkspaceSymbolProvider: &protocol.Or_ServerCapabilities_workspaceSymbolProvider{
					Value: nil,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDefinitionSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasDefinitionSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasReferencesSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "references supported",
			caps: &protocol.ServerCapabilities{
				ReferencesProvider: &protocol.Or_ServerCapabilities_referencesProvider{
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "references pointer non-nil but Value nil",
			caps: &protocol.ServerCapabilities{
				ReferencesProvider: &protocol.Or_ServerCapabilities_referencesProvider{
					Value: nil,
				},
			},
			expected: false,
		},
		{
			name:     "references nil",
			caps:     &protocol.ServerCapabilities{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasReferencesSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasReferencesSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasHoverSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "hover supported",
			caps: &protocol.ServerCapabilities{
				HoverProvider: &protocol.Or_ServerCapabilities_hoverProvider{
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "hover Value nil",
			caps: &protocol.ServerCapabilities{
				HoverProvider: &protocol.Or_ServerCapabilities_hoverProvider{
					Value: nil,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasHoverSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasHoverSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasDocumentSymbolSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "document symbols supported",
			caps: &protocol.ServerCapabilities{
				DocumentSymbolProvider: &protocol.Or_ServerCapabilities_documentSymbolProvider{
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "document symbols Value nil",
			caps: &protocol.ServerCapabilities{
				DocumentSymbolProvider: &protocol.Or_ServerCapabilities_documentSymbolProvider{
					Value: nil,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDocumentSymbolSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasDocumentSymbolSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasCallHierarchySupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "call hierarchy supported",
			caps: &protocol.ServerCapabilities{
				CallHierarchyProvider: &protocol.Or_ServerCapabilities_callHierarchyProvider{
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "call hierarchy Value nil",
			caps: &protocol.ServerCapabilities{
				CallHierarchyProvider: &protocol.Or_ServerCapabilities_callHierarchyProvider{
					Value: nil,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasCallHierarchySupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasCallHierarchySupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasWorkspaceSymbolSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "workspace symbol supported",
			caps: &protocol.ServerCapabilities{
				WorkspaceSymbolProvider: &protocol.Or_ServerCapabilities_workspaceSymbolProvider{
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "workspace symbol Value nil",
			caps: &protocol.ServerCapabilities{
				WorkspaceSymbolProvider: &protocol.Or_ServerCapabilities_workspaceSymbolProvider{
					Value: nil,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasWorkspaceSymbolSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasWorkspaceSymbolSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasRenameSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "rename supported as bool true",
			caps: &protocol.ServerCapabilities{
				RenameProvider: true,
			},
			expected: true,
		},
		{
			name: "rename supported as options struct",
			caps: &protocol.ServerCapabilities{
				RenameProvider: map[string]interface{}{"prepareProvider": true},
			},
			expected: true,
		},
		{
			name: "rename not supported (nil)",
			caps: &protocol.ServerCapabilities{
				RenameProvider: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasRenameSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasRenameSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasCodeActionSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "code action supported as bool",
			caps: &protocol.ServerCapabilities{
				CodeActionProvider: true,
			},
			expected: true,
		},
		{
			name: "code action supported as options",
			caps: &protocol.ServerCapabilities{
				CodeActionProvider: map[string]interface{}{"codeActionKinds": []string{"quickfix"}},
			},
			expected: true,
		},
		{
			name: "code action not supported",
			caps: &protocol.ServerCapabilities{
				CodeActionProvider: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasCodeActionSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasCodeActionSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasSignatureHelpSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "signature help supported",
			caps: &protocol.ServerCapabilities{
				SignatureHelpProvider: &protocol.SignatureHelpOptions{},
			},
			expected: true,
		},
		{
			name: "signature help not supported",
			caps: &protocol.ServerCapabilities{
				SignatureHelpProvider: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasSignatureHelpSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasSignatureHelpSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasCodeLensSupport(t *testing.T) {
	tests := []struct {
		name     string
		caps     *protocol.ServerCapabilities
		expected bool
	}{
		{
			name: "code lens supported",
			caps: &protocol.ServerCapabilities{
				CodeLensProvider: &protocol.CodeLensOptions{},
			},
			expected: true,
		},
		{
			name: "code lens not supported",
			caps: &protocol.ServerCapabilities{
				CodeLensProvider: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasCodeLensSupport(tt.caps)
			if result != tt.expected {
				t.Errorf("HasCodeLensSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
