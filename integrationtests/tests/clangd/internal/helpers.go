// Package internal contains shared helpers for Clangd tests
package internal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
)

// GetTestSuite returns a test suite for Clangd language server tests
func GetTestSuite(t *testing.T) *common.TestSuite {
	// Configure Clangd LSP
	repoRoot, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatalf("Failed to get repo root: %v", err)
	}

	config := common.LSPTestConfig{
		Name:             "clangd",
		Command:          "clangd",
		Args:             []string{}, // Will be set up after workspace is copied
		WorkspaceDir:     filepath.Join(repoRoot, "integrationtests/workspaces/clangd"),
		InitializeTimeMs: 2000,
	}

	// Create a test suite
	suite := common.NewTestSuite(t, config)

	// Set up the suite
	if err := suite.Setup(); err != nil {
		t.Fatalf("Failed to set up test suite: %v", err)
	}

	// After setup, the workspace has been copied to suite.WorkspaceDir
	// Now we need to:
	// 1. Update compile_commands.json with the actual workspace directory
	// 2. Configure clangd to use the copied workspace's compile_commands.json

	compileCommandsPath := filepath.Join(suite.WorkspaceDir, "compile_commands.json")
	if err := processCompileCommands(compileCommandsPath, suite.WorkspaceDir); err != nil {
		t.Fatalf("Failed to process compile_commands.json: %v", err)
	}

	// Shutdown the initial LSP client (started with wrong compile-commands-dir)
	// We need to properly close it to avoid resource leaks
	shutdownCtx, cancel := context.WithTimeout(suite.Context, 5*time.Second)
	defer cancel()

	if err := suite.Client.Shutdown(shutdownCtx); err != nil {
		t.Logf("Warning: Failed to shutdown LSP: %v", err)
	}
	if err := suite.Client.Exit(shutdownCtx); err != nil {
		t.Logf("Warning: Failed to exit LSP: %v", err)
	}
	if err := suite.Client.Close(); err != nil {
		t.Logf("Warning: Failed to close LSP: %v", err)
	}

	// Create new client with correct args
	suite.Config.Args = []string{"--compile-commands-dir=" + suite.WorkspaceDir}
	newClient, err := suite.RestartLSP()
	if err != nil {
		t.Fatalf("Failed to restart LSP with correct compile-commands-dir: %v", err)
	}
	suite.Client = newClient

	// Register cleanup - but only once
	t.Cleanup(func() {
		suite.Cleanup()
	})

	return suite
}

// processCompileCommands reads compile_commands.json and replaces ${WORKSPACE_DIR} with actual path
func processCompileCommands(path string, workspaceDir string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Replace the placeholder with the actual workspace directory
	updated := strings.ReplaceAll(string(content), "${WORKSPACE_DIR}", workspaceDir)

	return os.WriteFile(path, []byte(updated), 0644)
}
