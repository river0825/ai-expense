package http

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil) // Restore logger

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	loggingHandler := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "/test-path", nil)
	w := httptest.NewRecorder()

	loggingHandler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "[API]") {
		t.Errorf("Log output should contain [API] prefix")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Errorf("Log output should contain HTTP method")
	}
	if !strings.Contains(logOutput, "/test-path") {
		t.Errorf("Log output should contain URL path")
	}
	if !strings.Contains(logOutput, "200") {
		t.Errorf("Log output should contain status code")
	}
}
