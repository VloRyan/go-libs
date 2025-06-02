package router

import (
	"net/http"
	"testing"

	"github.com/vloryan/go-libs/httpx"
)

func TestEnsureContentTypeHandler(t *testing.T) {
	tests := []struct {
		name            string
		contentType     string
		wantContentType string
	}{
		{name: "Match", contentType: "text/html", wantContentType: "text/html"},
		{name: "No match", contentType: "text/html", wantContentType: "application/json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantUnsupportedMediaType := tt.contentType != tt.wantContentType
			req, err := http.NewRequest(http.MethodGet, "https://test.com", nil)
			req.Header.Set("Content-Type", tt.contentType)
			if err != nil {
				t.Fatal(err)
			}
			nextWasCalled := false
			handler := EnsureContentTypeHandler(tt.wantContentType, func(_ http.ResponseWriter, _ *http.Request) {
				nextWasCalled = true
			})
			writer := httpx.NewInMemResponseWriter()
			handler.ServeHTTP(writer, req)
			if wantUnsupportedMediaType {
				if writer.StatusCode != http.StatusUnsupportedMediaType {
					t.Errorf("EnsureContentTypeHandler() = response status = %v, want %v", writer.StatusCode, http.StatusUnsupportedMediaType)
				}
				if nextWasCalled {
					t.Errorf("EnsureContentTypeHandler() next was called on unsupported media type")
				}
			} else if !nextWasCalled {
				t.Errorf("EnsureContentTypeHandler() next was not called")
			}
		})
	}
}
