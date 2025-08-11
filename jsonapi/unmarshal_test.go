package jsonapi

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/vloryan/go-libs/reflectx"

	"github.com/google/go-cmp/cmp"
	"github.com/vloryan/go-libs/testhelper"
)

type testPersonUnmarshall struct {
	ID                uint                    `json:"id,omitempty"`
	Type              string                  `json:"type,omitempty"`
	Name              string                  `json:"name,omitempty"`
	Age               int                     `json:"age,omitempty"`
	Birthday          *time.Time              `json:"birthday,omitempty" time_format:"2006-01-02"`
	Salary            float64                 `json:"salary,omitempty"`
	Spouse            *testPersonUnmarshall   `json:"spouse,omitempty"`
	Children          []*testPersonUnmarshall `json:"children,omitempty"`
	NestedRef         *objectWithRef          `json:"nestedRef,omitempty"`
	ChildrenNestedRef []*objectWithRef        `json:"childrenNestedRef,omitempty"`
}

func (p *testPersonUnmarshall) GetIdentifier() *ResourceIdentifierObject {
	id := ""
	if p.ID != 0 {
		id = strconv.Itoa(int(p.ID))
	}
	return &ResourceIdentifierObject{ID: id, Type: p.Type}
}

func (p *testPersonUnmarshall) SetIdentifier(id *ResourceIdentifierObject) {
	if u, err := strconv.ParseUint(id.ID, 10, 64); err != nil {
		return
	} else {
		p.ID = uint(u)
	}
	p.Type = id.Type
}

type nestedPerson struct {
	Nested []*testPersonUnmarshall
}

func (p *nestedPerson) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: "", Type: ""}
}

func (p *nestedPerson) SetIdentifier(_ *ResourceIdentifierObject) {
}

type mapAttribObject struct {
	Map map[string]int
}

func (p *mapAttribObject) GetIdentifier() *ResourceIdentifierObject {
	return &ResourceIdentifierObject{ID: "", Type: ""}
}

func (p *mapAttribObject) SetIdentifier(_ *ResourceIdentifierObject) {
}

func TestUnmarshalResourceObject(t *testing.T) {
	tests := []struct {
		name     string
		jsonFile string
		want     any
	}{{
		name:     "person",
		jsonFile: "./test/document_person_unmarshal.json",
		want:     &testPersonUnmarshall{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T00:00:00Z"))},
	}, {
		name:     "person with children included",
		jsonFile: "./test/document_person_with_children_included.json",
		want: &testPersonUnmarshall{ID: 814, Type: "person", Name: "Gunter Hammer", Children: []*testPersonUnmarshall{
			{ID: 816, Type: "person", Name: "Georg Weber"},
			{ID: 817, Type: "person", Name: "Manfred Hammer"},
		}},
	}, {
		name:     "nested person",
		jsonFile: "./test/document_nestedPerson.json",
		want:     &nestedPerson{Nested: []*testPersonUnmarshall{{ID: 814, Type: "person", Name: "Gunter Hammer"}}},
	}, {
		name:     "indexed relationship",
		jsonFile: "./test/document_indexedRelationship.json",
		want: &nestedPerson{Nested: []*testPersonUnmarshall{
			{ID: 814, Type: "person", Name: "Gunter Hammer"},
			{ID: 4711, Type: "person", Name: "Erik Janson"},
		}},
	}, {
		name:     "map",
		jsonFile: "./test/document_mapAttrib.json",
		want:     &mapAttribObject{Map: map[string]int{"1": 1, "2": 2}},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument()
			j := testhelper.ReadJson(t, tt.jsonFile)
			err := json.Unmarshal([]byte(j), doc)
			if err != nil {
				t.Fatal(err)
			}
			wantT := reflectx.TypeOf(tt.want, true)
			got := reflect.New(wantT).Interface()

			err = UnmarshalResourceObject(doc.Data.(*ResourceObject), doc.Included, got)
			if err != nil {
				t.Fatalf("UnmarshalResourceObject() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("UnmarshalResourceObject() mismatch (-want +got):\\n%s", diff)
			}
		})
	}
}
