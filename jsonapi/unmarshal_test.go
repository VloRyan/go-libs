package jsonapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/vloryan/go-libs/reflectx"

	"github.com/google/go-cmp/cmp"
	"github.com/vloryan/go-libs/testhelper"
)

type nestedPerson struct {
	Nested []*testPerson
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
		jsonFile: "./test/document_person.json",
		want:     &testPerson{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
	}, {
		name:     "person with children included",
		jsonFile: "./test/document_person_with_children_included.json",
		want: &testPerson{ID: 814, Type: "person", Name: "Gunter Hammer", Children: []*testPerson{
			{ID: 816, Type: "person", Name: "Georg Weber"},
			{ID: 817, Type: "person", Name: "Manfred Hammer"},
		}},
	}, {
		name:     "nested person",
		jsonFile: "./test/document_nestedPerson.json",
		want:     &nestedPerson{Nested: []*testPerson{{ID: 814, Type: "person", Name: "Gunter Hammer"}}},
	}, {
		name:     "indexed relationship",
		jsonFile: "./test/document_indexedRelationship.json",
		want: &nestedPerson{Nested: []*testPerson{
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
