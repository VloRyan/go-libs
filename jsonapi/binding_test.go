package jsonapi

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

type testObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Foo  string `json:"foo"`
}

func (t *testObject) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: t.ID, Type: t.Type}
}

func (t *testObject) SetIdentifier(id *ResourceIdentifierObject) {
	t.ID = id.ID
	t.Type = id.Type
}
func Test_Binding_Bind(t *testing.T) {

	tests := []struct {
		name     string
		bodyFunc func() io.Reader
		obj      any
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "bind document body",
			bodyFunc: func() io.Reader {
				body := `{"data":{"id": "1", "type":"test", "attributes": {"foo": "bar"}}}`
				return bytes.NewReader([]byte(body))
			},
			obj:     &testObject{ID: "1", Type: "test", Foo: "bar"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com/resource", tt.bodyFunc())
			if err != nil {
				t.Fatal(err)
			}
			tt.wantErr(t, Binding.Bind(req, tt.obj), fmt.Sprintf("Bind(%v, %v)", req, tt.obj))
		})
	}
}
