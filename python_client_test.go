package notes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPythonRunner_Execute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/execute" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req pythonExecRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Code != "print('hello')" {
			t.Errorf("unexpected code: %s", req.Code)
		}
		if req.Variables["x"] != float64(42) {
			t.Errorf("unexpected variable x: %v", req.Variables["x"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pythonExecResponse{Output: "hello"})
	}))
	defer srv.Close()

	runner := NewPythonRunner(srv.URL, "")
	output, err := runner.Execute(context.Background(), "print('hello')", map[string]interface{}{"x": 42}, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
}

func TestPythonRunner_ExecuteError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(pythonExecResponse{Error: "name 'x' is not defined"})
	}))
	defer srv.Close()

	runner := NewPythonRunner(srv.URL, "")
	_, err := runner.Execute(context.Background(), "print(x)", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "name 'x' is not defined" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPythonRunner_Unavailable(t *testing.T) {
	runner := NewPythonRunner("http://localhost:1", "")
	_, err := runner.Execute(context.Background(), "print(1)", nil, nil)
	if err == nil {
		t.Fatal("expected error for unavailable server")
	}
}

func TestPythonRunner_WithRequirements(t *testing.T) {
	var receivedReqs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req pythonExecRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedReqs = req.Requirements
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pythonExecResponse{Output: "ok"})
	}))
	defer srv.Close()

	runner := NewPythonRunner(srv.URL, "")
	runner.Execute(context.Background(), "import requests", nil, []string{"requests", "beautifulsoup4"})

	if len(receivedReqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(receivedReqs))
	}
	if receivedReqs[0] != "requests" {
		t.Errorf("expected 'requests', got %q", receivedReqs[0])
	}
}
