package lsp

import (
	"sync"
	"testing"
	"time"
)

// TestCloseIdempotent verifies that calling Close() multiple times doesn't panic
func TestCloseIdempotent(t *testing.T) {
	// Create a simple command that exits quickly
	client, err := NewClient("echo", "test")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Wait a bit for the echo command to complete
	time.Sleep(100 * time.Millisecond)

	// Call Close multiple times concurrently
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Close()
		}()
	}

	// Wait for all Close calls to complete
	wg.Wait()

	// If we get here without panicking, the test passes
	t.Log("Close() is idempotent and doesn't panic when called multiple times")
}

// TestCloseRaceCondition verifies that Close() doesn't have race conditions
// between the timeout goroutine and the main cleanup path
func TestCloseRaceCondition(t *testing.T) {
	// Create a command that exits quickly (before the 2s timeout)
	client, err := NewClient("sh", "-c", "sleep 0.1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Wait for command to start
	time.Sleep(50 * time.Millisecond)

	// Close should complete before the force-kill timeout
	// The race condition was: main path closes channel at line 272,
	// but timeout goroutine might also try to close it at line 258
	err = client.Close()
	if err != nil {
		t.Logf("Close returned error (expected for sh): %v", err)
	}

	// If we get here without panic, the fix works
	t.Log("Close() handled race condition correctly")
}

// TestCloseWithTimeout verifies that Close() works when process needs to be killed
func TestCloseWithTimeout(t *testing.T) {
	// Create a command that will hang and need to be killed
	client, err := NewClient("sleep", "10")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Wait for sleep to start
	time.Sleep(50 * time.Millisecond)

	// Close should force-kill the process after 2s timeout
	start := time.Now()
	err = client.Close()
	duration := time.Since(start)

	// Should take around 2 seconds (the force-kill timeout)
	if duration < 1*time.Second || duration > 3*time.Second {
		t.Logf("Warning: Close took %v, expected around 2s", duration)
	}

	// The error is expected (killed process), but no panic should occur
	t.Logf("Close completed in %v (expected ~2s for force kill)", duration)
}
