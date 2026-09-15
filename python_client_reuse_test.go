package notes

import (
	"testing"
	"time"
)

// PythonRunner is long-lived (built once at startup, stored on the Store), but
// Execute used to allocate a fresh http.Client on every code run. The runner now
// holds one pooled client created in NewPythonRunner. Assert it exists, is
// bounded, and is reused across calls.
func TestPythonRunner_SharedHTTPClient(t *testing.T) {
	r := NewPythonRunner("http://localhost/apps/shell/api", "tok")
	if r.httpClient == nil {
		t.Fatal("PythonRunner must hold a pooled http client created in NewPythonRunner")
	}
	if r.httpClient.Timeout <= 0 || r.httpClient.Timeout > 5*time.Minute {
		t.Fatalf("runner client timeout = %v, want a sane positive bound", r.httpClient.Timeout)
	}
	// A second runner is independent, but each runner reuses its own single client.
	if NewPythonRunner("http://x", "").httpClient == r.httpClient {
		t.Error("distinct runners should not share the same client instance")
	}
}
