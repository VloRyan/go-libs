package httpx

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vloryan/go-libs/testhelper"
)

func TestStreamAndReplaceToken(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		wantStatusCode int
	}{
		{
			name:           "replace CONTEXT_ROOT",
			fileName:       "lorem.txt",
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantContent := testhelper.ReadFile(t, "./test/"+baseWithoutExt(tt.fileName)+"_want"+filepath.Ext(tt.fileName))
			w := NewInMemResponseWriter()

			StreamFileAndReplaceToken(w, os.DirFS("./test/"), tt.fileName, "::PLACEHOLDER::", "::REPLACED::")

			if w.StatusCode != tt.wantStatusCode {
				t.Errorf("StreamFileAndReplaceToken(): got statusCode %d, want %d", w.StatusCode, tt.wantStatusCode)
			}
			if diff := cmp.Diff(wantContent, string(w.Body)); diff != "" {
				t.Errorf("StreamFileAndReplaceToken(): got response body (-want +got):\n%s", diff)
			}
		})
	}
}

func baseWithoutExt(name string) string {
	ext := filepath.Ext(name)
	n, _ := strings.CutSuffix(filepath.Base(name), ext)
	return n
}
