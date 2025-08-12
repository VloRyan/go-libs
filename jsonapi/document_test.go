package jsonapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/vloryan/go-libs/testhelper"
)

func TestSetObjectData(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		args     any
		fieldSet []string
		wantErr  bool
	}{{
		name:     "person",
		fileName: "./test/document_person.json",
		args:     &testPerson{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
	}, {
		name:     "people",
		fileName: "./test/document_people.json",
		args: []*testPerson{
			{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
			{ID: 4712, Type: "person", Name: "Gudrun Müller", Age: 46, Salary: 777.77, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1986-06-15T12:13:59Z"))},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument()

			err := doc.SetObjectData(tt.args, tt.fieldSet...)

			if (err != nil) != tt.wantErr {
				t.Errorf("TestApplyData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := testhelper.ReadJson(t, tt.fileName)
			gotJson, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			assert.JSONEq(t, want, string(gotJson), "marshalling different")
		})
	}
}

func TestMapData(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     any
		wantErr  bool
	}{{
		name:     "person",
		fileName: "./test/document_person.json",
		want:     &testPerson{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
	}, {
		name:     "person with children included",
		fileName: "./test/document_person_with_children_included.json",
		want: &testPerson{ID: 814, Type: "person", Name: "Gunter Hammer", Children: []*testPerson{
			{ID: 816, Type: "person", Name: "Georg Weber"},
			{ID: 817, Type: "person", Name: "Manfred Hammer"},
		}},
	}, {
		name:     "people",
		fileName: "./test/document_people.json",
		want: []*testPerson{
			{ID: 4711, Type: "person", Name: "Hans Müller", Age: 47, Salary: 666.66, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1985-06-15T12:13:59Z"))},
			{ID: 4712, Type: "person", Name: "Gudrun Müller", Age: 46, Salary: 777.77, Birthday: testhelper.Ptr(testhelper.ParseTime(t, "1986-06-15T12:13:59Z"))},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument()
			err := json.Unmarshal([]byte(testhelper.ReadJson(t, tt.fileName)), doc)
			if err != nil {
				t.Fatal(err)
			}
			var got any
			if reflect.TypeOf(doc.Data).Kind() == reflect.Slice {
				s := make([]*testPerson, 0, 10)
				err = doc.MapData(&s)
				got = s
			} else {
				got = &testPerson{}
				err = doc.MapData(got)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ToDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToDocument() mismatch (-want +got):\\n%s", diff)
			}
		})
	}
}

func TestRelationships(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     map[string][]*ResourceIdentifierObject
	}{{
		name: "without_relationships",
		want: map[string][]*ResourceIdentifierObject{},
	}, {
		name: "one_to_one",
		want: map[string][]*ResourceIdentifierObject{"one": {{ID: "2", Type: "object"}}},
	}, {
		name: "one_to_many",
		want: map[string][]*ResourceIdentifierObject{"many": {{ID: "2", Type: "object"}, {ID: "3", Type: "other"}}},
	}, {
		name: "null",
		want: map[string][]*ResourceIdentifierObject{"null": nil},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument()
			fileName := tt.fileName
			if fileName == "" {
				fileName = "document_" + tt.name + ".json"
			}
			err := json.Unmarshal([]byte(testhelper.ReadJson(t, "./test/relationships/"+fileName)), doc)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(doc.Relationships(), tt.want); diff != "" {
				t.Errorf("Relationships() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}
