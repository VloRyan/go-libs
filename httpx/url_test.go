package httpx

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/vloryan/go-libs/testhelper"
	"golang.org/x/net/html"
)

func TestGenerateReplacedIndexHTML(t *testing.T) {
	tests := []struct {
		name       string
		assetPath  string
		serverData string
		wantFile   string
		wantErr    assert.ErrorAssertionFunc
	}{{
		name:      "replace asset path",
		assetPath: "/REPLACED",
		wantFile:  "index_want_asset_path.html",
		wantErr:   assert.NoError,
	}, {
		name:       "replace server data",
		serverData: `{"fieldA": "valueA", "fieldB": "valueB"};`,
		assetPath:  "/",
		wantFile:   "index_want_server_data.html",
		wantErr:    assert.NoError,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := renderAsHTML(t, testhelper.ReadFile(t, "./test/"+tt.wantFile))
			fSys := os.DirFS("./test/")
			got, err := GenerateReplacedIndexHTML(fSys, tt.assetPath, tt.serverData)
			if !tt.wantErr(t, err) {
				t.Fatalf("GenerateReplacedIndexHTML() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("GenerateReplacedIndexHTML() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func renderAsHTML(t *testing.T, content string) string {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	w := NewInMemResponseWriter()
	if err := html.Render(w, doc); err != nil {
		t.Fatal(err)
	}
	return string(w.Body)
}
