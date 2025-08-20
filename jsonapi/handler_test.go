package jsonapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vloryan/go-libs/sqlx/pagination"

	"github.com/vloryan/go-libs/httpx"
)

type Item struct {
	ID            uint   `json:"ID,omitempty"`
	Type          string `json:"type,omitempty"`
	AttributeA    string `json:"attributeA,omitempty"`
	AttributeB    string `json:"attributeB,omitempty"`
	RelationshipA *Item  `json:"relationshipA,omitempty"`
}

var defaultItem = &Item{ID: 4711, Type: "default.item", AttributeA: "A", AttributeB: "B"}

func (d *Item) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: fmt.Sprintf("%v", d.ID), Type: d.Type}
}

func (d *Item) SetIdentifier(id *ResourceIdentifierObject) {
	uid, err := strconv.ParseUint(id.ID, 10, 64)
	if err != nil {
		return
	}
	d.ID = uint(uid)
	d.Type = id.Type
}

func NewRequest(t *testing.T, method, url string, body io.Reader, header http.Header) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header = header
	return req
}

func TestGenericHandler_Handle(t *testing.T) {
	type testCase struct {
		name        string
		reqFunc     func(t *testing.T) *http.Request
		f           func(req *http.Request) (*DocumentData[*Item], *Error)
		fieldFilter ResourceObjectFieldFilterFunc
		resolveMap  map[string]*ResourceObject
		wantStatus  int
		wantBody    string
	}
	tests := []testCase{{
		name: "unsupported media type",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080", nil, map[string][]string{"Content-Type": {"application/json"}})
		},
		f:          func(req *http.Request) (*DocumentData[*Item], *Error) { return nil, nil },
		wantStatus: http.StatusUnsupportedMediaType, wantBody: "Unsupported Media Type",
	}, {
		name: "ok",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080/default/item/4711", nil, map[string][]string{"Content-Type": {MediaType}})
		},
		f: func(req *http.Request) (*DocumentData[*Item], *Error) {
			return NewDocumentData[*Item](defaultItem, "/default/item"), nil
		},
		wantStatus: http.StatusOK, wantBody: `{"data":{"id":"4711","type":"default.item","attributes":{"attributeA":"A","attributeB":"B"},"links":{"self":"/default/item/4711"}},"jsonapi":{"version":"1.1"}}`,
	}, {
		name: "no content",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080/default/item/4711", nil, map[string][]string{"Content-Type": {MediaType}})
		},
		f: func(req *http.Request) (*DocumentData[*Item], *Error) {
			return nil, nil
		},
		wantStatus: http.StatusNoContent, wantBody: "",
	}, {
		name: "sparse fieldset",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080/default/item/4711?fields[default.item]=attributeB", nil, map[string][]string{"Content-Type": {MediaType}})
		},
		f: func(req *http.Request) (*DocumentData[*Item], *Error) {
			return NewDocumentData[*Item](defaultItem, "/default/item"), nil
		},
		wantStatus: http.StatusOK, wantBody: `{"data":{"id":"4711","type":"default.item","attributes":{"attributeB":"B"},"links":{"self":"/default/item/4711"}},"jsonapi":{"version":"1.1"}}`,
	}, {
		name: "additional field filter",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080/default/item/4711?fields[default.item]=attributeB", nil, map[string][]string{"Content-Type": {MediaType}})
		},
		f: func(req *http.Request) (*DocumentData[*Item], *Error) {
			return NewDocumentData[*Item](defaultItem, "/default/item"), nil
		},
		fieldFilter: func(typeName, fieldName string) bool {
			return !strings.EqualFold(fieldName, "attributeB")
		},
		wantStatus: http.StatusOK, wantBody: `{"data":{"id":"4711","type":"default.item","links":{"self":"/default/item/4711"}},"jsonapi":{"version":"1.1"}}`,
	}, {
		name: "error",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080/default/item/4711?fields[default.item]=attributeB", nil, map[string][]string{"Content-Type": {MediaType}})
		},
		f: func(req *http.Request) (*DocumentData[*Item], *Error) {
			return nil, NewError(http.StatusBadRequest, "invalid id", errors.New("failed to parse id"))
		},
		wantStatus: http.StatusBadRequest, wantBody: `{"errors":[{"status":"400","title":"invalid id","detail":"failed to parse id"}],"jsonapi":{"version":"1.1"}}`,
	}, {
		name: "include",
		reqFunc: func(t *testing.T) *http.Request {
			return NewRequest(t, http.MethodGet, "http://localhost:8080/default/item/4711?include=relationshipA", nil, map[string][]string{"Content-Type": {MediaType}})
		},
		f: func(req *http.Request) (*DocumentData[*Item], *Error) {
			return NewDocumentData[*Item](&Item{ID: 4712, Type: "default.item", RelationshipA: defaultItem}, "/default/item"), nil
		},
		resolveMap: map[string]*ResourceObject{"4711": {ResourceIdentifierObject: ResourceIdentifierObject{ID: "4711", Type: "default.item"}, Links: map[string]any{"self": "/default/item/4711"}}},
		wantStatus: http.StatusOK, wantBody: `{"data":{"id":"4712","type":"default.item","relationships":{"relationshipA":{"data":{"id":"4711","type":"default.item"}}},"links":{"self":"/default/item/4712"}},"jsonapi":{"version":"1.1"},"included":[{"id":"4711","type":"default.item","links":{"self":"/default/item/4711"}}]}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := httpx.NewInMemResponseWriter()
			h := &GenericHandler[*Item]{
				ResolveObjectWithReqFunc: func(req *http.Request, id *ResourceIdentifierObject) (*ResourceObject, *Error) {
					if item, ok := tt.resolveMap[id.ID]; ok {
						return item, nil
					}
					return nil, nil
				},
				FieldFilterFunc: tt.fieldFilter,
			}
			req := tt.reqFunc(t)
			h.Handle(tt.f)(writer, req)
			if writer.Header().Get("Content-Type") != MediaType+"; charset=utf-8" {
				t.Fatalf("Content-Type mismatch, want: %s, got: %s", MediaType, writer.Header().Get("Content-Type"))
			}
			if writer.StatusCode != tt.wantStatus {
				t.Fatalf("StatusCode mismatch, want: %d, got: %d\nreponse:\n%s", tt.wantStatus, writer.StatusCode, string(writer.Body))
			}
			if diff := cmp.Diff(tt.wantBody, string(writer.Body)); diff != "" {
				t.Fatalf("Body mismatch (-want +got):\\n%s", diff)
			}
		})
	}
}

func TestGenericHandler_resolveIncludes(t *testing.T) {
	type testCase struct {
		name          string
		includes      []string
		relationships map[string]*RelationshipObject
		resolveMap    map[string]*ResourceObject
		locals        map[string]*ResourceObject
		want          []*ResourceObject
	}
	tests := []testCase{{
		name:     "comments",
		includes: []string{"comments"},
		relationships: map[string]*RelationshipObject{
			"comments": {Data: &ResourceIdentifierObject{ID: "5", Type: "comments"}},
		},
		resolveMap: map[string]*ResourceObject{
			"5": {ResourceIdentifierObject: ResourceIdentifierObject{ID: "5", Type: "comments"}},
		},
		want: []*ResourceObject{{
			ResourceIdentifierObject: ResourceIdentifierObject{ID: "5", Type: "comments"},
		}},
	}, {
		name:     "comments.author",
		includes: []string{"comments.author"},
		relationships: map[string]*RelationshipObject{
			"comments": {
				Data: []*ResourceIdentifierObject{
					{ID: "5", Type: "comments"},
					{ID: "12", Type: "comments"},
				},
			},
		},
		resolveMap: map[string]*ResourceObject{
			"9": {ResourceIdentifierObject: ResourceIdentifierObject{ID: "9", Type: "author"}},
			"5": {
				ResourceIdentifierObject: ResourceIdentifierObject{ID: "5", Type: "comments"},
				Relationships: map[string]*RelationshipObject{"author": {
					Data: &ResourceIdentifierObject{ID: "9", Type: "author"},
				}},
			},
			"12": {
				ResourceIdentifierObject: ResourceIdentifierObject{ID: "12", Type: "comments"},
				Relationships: map[string]*RelationshipObject{"author": {
					Data: &ResourceIdentifierObject{ID: "9", Type: "author"},
				}},
			},
		},
		want: []*ResourceObject{
			{
				ResourceIdentifierObject: ResourceIdentifierObject{ID: "5", Type: "comments"},
				Relationships: map[string]*RelationshipObject{"author": {
					Data: &ResourceIdentifierObject{ID: "9", Type: "author"},
				}},
			},
			{ResourceIdentifierObject: ResourceIdentifierObject{ID: "9", Type: "author"}},
			{
				ResourceIdentifierObject: ResourceIdentifierObject{ID: "12", Type: "comments"},
				Relationships: map[string]*RelationshipObject{"author": {
					Data: &ResourceIdentifierObject{ID: "9", Type: "author"},
				}},
			},
		},
	}, {
		name:     "resolve from lid",
		includes: []string{"comments.author"},
		relationships: map[string]*RelationshipObject{
			"comments": {
				Data: []*ResourceIdentifierObject{{LID: "default.type_0_1", Type: "comments"}},
			},
		},
		resolveMap: map[string]*ResourceObject{
			"9": {ResourceIdentifierObject: ResourceIdentifierObject{ID: "9", Type: "author"}},
		},
		locals: map[string]*ResourceObject{
			"default.type_0_1": {
				ResourceIdentifierObject: ResourceIdentifierObject{LID: "default.type_0_1", Type: "comments"},
				Attributes:               map[string]any{"text": "Success if you can read me"},
				Relationships: map[string]*RelationshipObject{"author": {
					Data: &ResourceIdentifierObject{ID: "9", Type: "author"},
				}},
			},
		},
		want: []*ResourceObject{
			{
				ResourceIdentifierObject: ResourceIdentifierObject{LID: "default.type_0_1", Type: "comments"},
				Attributes:               map[string]any{"text": "Success if you can read me"},
				Relationships: map[string]*RelationshipObject{"author": {
					Data: &ResourceIdentifierObject{ID: "9", Type: "author"},
				}},
			}, {ResourceIdentifierObject: ResourceIdentifierObject{ID: "9", Type: "author"}},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := GenericHandler[*Item]{}
			obj := &ResourceObject{
				ResourceIdentifierObject: ResourceIdentifierObject{ID: "0", Type: "default.type"},
				Relationships:            tt.relationships,
				Meta:                     MetaData{"local-objects": tt.locals},
			}

			doc := &Document{
				Data: obj,
				Meta: make(MetaData),
			}
			err := handler.resolveIncludes(func(id *ResourceIdentifierObject) (*ResourceObject, *Error) {
				if item, ok := tt.resolveMap[id.ID]; ok {
					return item, nil
				}
				return nil, nil
			}, tt.includes, doc)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, doc.Included); diff != "" {
				t.Errorf("resolveIncludes() error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGenericHandler_applyMetadata(t *testing.T) {
	type testCase struct {
		name string
		page *pagination.Page
		meta MetaData
		want MetaData
	}
	tests := []testCase{{
		name: "with meta data",
		meta: map[string]any{"a": "va", "b": "vb"},
		want: map[string]any{"a": "va", "b": "vb"},
	}, {
		name: "with page",
		page: &pagination.Page{
			Offset:     1,
			Limit:      2,
			Sort:       []string{"a", "b"},
			TotalCount: 7,
		},
		want: map[string]any{"page[limit]": 2, "page[offset]": 1, "page[sort]": "a,b", "page[total]": 7},
	}, {
		name: "with page and meta data",
		meta: map[string]any{"b": "vb"},
		page: &pagination.Page{
			Offset:     2,
			Limit:      3,
			Sort:       []string{"a", "b"},
			TotalCount: 8,
		},
		want: map[string]any{"b": "vb", "page[limit]": 3, "page[offset]": 2, "page[sort]": "a,b", "page[total]": 8},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := GenericHandler[*Item]{}
			doc := NewDocument()
			data := &DocumentData[*Item]{
				Page:     tt.page,
				MetaData: tt.meta,
			}
			h.applyMetadata(doc, data)

			if diff := cmp.Diff(tt.want, doc.Meta); diff != "" {
				t.Errorf("applyMetadata() error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
