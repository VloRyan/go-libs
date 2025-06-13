package jsonapi

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vloryan/go-libs/testhelper"
)

type object struct {
	ID    string          `json:"id"`
	Type  string          `json:"type"`
	Field *fieldMarshaler `json:"field"`
}

func (p *object) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: p.ID, Type: p.Type}
}

func (p *object) SetIdentifier(id *ResourceIdentifierObject) {
	p.ID = id.ID
	p.Type = id.Type
}

type fieldMarshaler struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (m *fieldMarshaler) MarshalField(obj *ResourceObject, sf reflect.StructField) error {
	rels := obj.Relationships
	if rels == nil {
		rels = make(map[string]*RelationshipObject)
	}
	rels[sf.Tag.Get("json")] = &RelationshipObject{
		Data: ResourceIdentifierObject{ID: m.ID, Type: m.Type},
	}
	obj.Relationships = rels
	return nil
}

type ref struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (p *ref) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: p.ID, Type: p.Type}
}

func (p *ref) SetIdentifier(id *ResourceIdentifierObject) {
	p.ID = id.ID
	p.Type = id.Type
}

type objectWithRef struct {
	Name  string `json:"name,omitempty"`
	Other *ref   `json:"other"`
}
type testPerson struct {
	ID                uint             `json:"id,omitempty"`
	Type              string           `json:"type,omitempty"`
	Name              string           `json:"name,omitempty"`
	Age               int              `json:"age,omitempty"`
	Birthday          *time.Time       `json:"birthday,omitempty"`
	Salary            float64          `json:"salary,omitempty"`
	Spouse            *testPerson      `json:"spouse,omitempty"`
	Children          []*testPerson    `json:"children,omitempty"`
	NestedRef         *objectWithRef   `json:"nestedRef,omitempty"`
	ChildrenNestedRef []*objectWithRef `json:"childrenNestedRef,omitempty"`
}

func (p *testPerson) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: strconv.Itoa(int(p.ID)), Type: p.Type}
}

func (p *testPerson) SetIdentifier(id *ResourceIdentifierObject) {
	if u, err := strconv.ParseUint(id.ID, 10, 64); err != nil {
		return
	} else {
		p.ID = uint(u)
	}
	p.Type = id.Type
}

func TestMarshalResourceObject(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		args        any
		fieldFilter ResourceObjectFieldFilterFunc
		wantErr     bool
	}{{
		name:     "person",
		fileName: "./test/person.json",
		args:     &testPerson{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
	}, {
		name:     "person with spouse",
		fileName: "./test/person_with_spouse.json",
		args:     &testPerson{ID: 815, Type: "person", Name: "Hanna Weber", Spouse: &testPerson{ID: 816, Type: "person", Name: "Georg Weber"}},
	}, {
		name:     "person only age",
		fileName: "./test/person_only_age.json",
		args:     &testPerson{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
		fieldFilter: func(t, f string) bool {
			return strings.EqualFold(f, "age")
		},
	}, {
		name:     "nested ref",
		fileName: "./test/nested_ref.json",
		args: &testPerson{
			ID: 4711, Type: "person",
			NestedRef: &objectWithRef{Name: "I am the object", Other: &ref{ID: "4711", Type: "Type1"}},
			ChildrenNestedRef: []*objectWithRef{
				{Name: "First", Other: &ref{ID: "4712", Type: "Type2"}},
				{Name: "Second", Other: &ref{ID: "4713", Type: "Type3"}},
			},
		},
	}, {
		name:     "fieldMarshaler",
		fileName: "./test/field_marshaler.json",
		args:     &object{ID: "1", Type: "object", Field: &fieldMarshaler{ID: "4711", Type: "fieldMarshaler"}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := MarshalResourceObject(tt.args, tt.fieldFilter)

			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalResourceObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := testhelper.ReadJson(t, tt.fileName)
			gotJson, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			assert.JSONEq(t, want, string(gotJson), "marshalling different")
		})
	}
}
