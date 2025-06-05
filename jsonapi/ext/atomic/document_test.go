package atomic

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapData(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		lidCache []string
	}{{
		name:     "operation",
		fileName: "./test/operation.json",
	}, {
		name:     "operations",
		fileName: "./test/operations.json",
	}, {
		name:     "operation_lid",
		fileName: "./test/operation_lid.json",
		lidCache: []string{"thePost", "theTag", "catA", "catB"},
	}, {
		name:     "result",
		fileName: "./test/result.json",
	}, {
		name:     "results",
		fileName: "./test/results.json",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{}
			want := readJson(t, tt.fileName)
			err := json.Unmarshal([]byte(want), doc)
			if err != nil {
				t.Fatal(err)
			}

			got, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			assert.JSONEq(t, want, string(got), "marshalling different")
			cache := doc.LIDCache()
			lids := make([]string, 0, len(cache))
			for k := range cache {
				lids = append(lids, k)
			}
			assert.ElementsMatch(t, tt.lidCache, lids)
		})
	}
}

func readJson(t *testing.T, file string) string {
	jsonFile, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	return string(byteValue)
}
