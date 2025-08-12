package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/vloryan/go-libs/sqlx/pagination"

	"github.com/vloryan/go-libs/reflectx"
)

const (
	MediaType = "application/vnd.api+json"
	Version   = "1.1"
)

type TypeToPathConverterFunc func(typeStr string) string

type Document struct {
	Data    any             `json:"-"`
	RawData json.RawMessage `json:"data,omitempty"`
	Errors  []*Error        `json:"errors,omitempty"`
	Meta    MetaData        `json:"meta,omitempty"`

	JSONAPI  *Object           `json:"jsonapi,omitempty"`
	Links    map[string]any    `json:"links,omitempty"`
	Included []*ResourceObject `json:"included,omitempty"`
}

func NewDocument() *Document {
	return &Document{
		Meta: make(MetaData),
		JSONAPI: &Object{
			Version: Version,
		},
	}
}

type RelationshipObject struct {
	Links   map[string]any  `json:"links,omitempty"`
	Data    any             `json:"-"`
	RawData json.RawMessage `json:"data,omitempty"`
	Meta    MetaData        `json:"meta,omitempty"`
}
type LinkObject struct {
	HRef        string   `json:"href"`
	Rel         string   `json:"rel,omitempty"`
	Describedby string   `json:"describedby,omitempty"`
	Title       MetaData `json:"title,omitempty"`
	Type        string   `json:"type,omitempty"`
	HRefLang    string   `json:"hreflang,omitempty"`
	Meta        MetaData `json:"meta,omitempty"`
}
type Object struct {
	Version string   `json:"version,omitempty"`
	Ext     []string `json:"ext,omitempty"`
	Profile []string `json:"profile,omitempty"`
	Meta    MetaData `json:"meta,omitempty"`
}
type MetaData map[string]any

type ResourceObjectFieldFilterFunc func(typeName, fieldName string) bool

func (d *Document) AddError(errors ...*Error) *Document {
	d.Errors = append(d.Errors, errors...)
	return d
}

func (d *Document) FindIncluded(id, objType string) *ResourceObject {
	for _, obj := range d.Included {
		if obj.ID == id && obj.Type == objType {
			return obj
		}
	}
	return nil
}

func (d *Document) ObjectsLen() int {
	v := reflect.ValueOf(d.Data)
	if v.Type().Kind() == reflect.Struct {
		return 1
	}
	return v.Len()
}

func (d *Document) Objects(i int) *ResourceObject {
	v := reflect.ValueOf(d.Data)
	if v.IsNil() {
		return nil
	}
	if v.Type().Kind() == reflect.Ptr {
		return v.Interface().(*ResourceObject)
	}
	return v.Index(i).Interface().(*ResourceObject)
}

func (d *Document) SetObjectData(v any, fieldNames ...string) error {
	fieldNameFilter := func(typeName, fieldName string) bool {
		if len(fieldNames) > 0 {
			return slices.ContainsFunc(fieldNames, func(s string) bool {
				return strings.EqualFold(s, fieldName)
			})
		}
		return true
	}
	return d.SetObjectDataFieldFilterFunc(v, fieldNameFilter)
}

func (d *Document) SetObjectDataFieldFilterFunc(v any, fieldNameFunc ResourceObjectFieldFilterFunc) error {
	if v == nil {
		return nil
	}
	objects := make([]*ResourceObject, 0, 10)
	if err := ForEachElem(v, func(e any) error {
		obj, err := MarshalResourceObject(e, fieldNameFunc)
		if err != nil {
			return err
		}
		objects = append(objects, obj)
		return nil
	}); err != nil {
		return err
	}
	vv := reflectx.ValueOf(v, true)
	switch vv.Kind() {
	case reflect.Struct:
		if len(objects) != 1 {
			return fmt.Errorf("mapped %d objects, expected 1", len(objects))
		}
		d.Data = objects[0]
	case reflect.Slice:
		d.Data = objects
	default:
		return errors.New("only struct and slice of structs are allowed")
	}
	return nil
}

func (d *Document) Relationships() map[string][]*ResourceIdentifierObject {
	m := make(map[string][]*ResourceIdentifierObject)
	_ = ForEachElem(d.Data, func(e *ResourceObject) error {
		for k, v := range e.Relationships {
			if v.Data == nil {
				_, ok := m[k]
				if !ok {
					m[k] = nil
				}
				continue
			}
			data, ok := m[k]
			if !ok || data == nil {
				data = make([]*ResourceIdentifierObject, 0)
			}
			if reflect.TypeOf(v.Data).Kind() == reflect.Slice {
				vData := v.Data.([]*ResourceIdentifierObject)
				data = append(data, vData...)
				m[k] = data
			} else {
				vData := v.Data.(*ResourceIdentifierObject)
				m[k] = append(data, vData)
			}
		}
		return nil
	})
	return m
}

func (d *Document) MapData(v any) error {
	if v == nil {
		return errors.New("v is nil")
	}
	if d.Data == nil {
		return nil
	}
	dv := reflect.ValueOf(d.Data)
	if dv.Kind() == reflect.Slice {
		vv := reflect.ValueOf(v)
		if vv.Kind() != reflect.Ptr {
			return errors.New("v is no slice ptr")
		}
		vve := vv.Elem()
		if vve.Kind() != reflect.Slice {
			return errors.New("v is no slice ptr")
		}
		for i := 0; i < dv.Len(); i++ {
			obj := dv.Index(i).Interface().(*ResourceObject)
			et := reflectx.ElemTypeOf(v, true)
			if et.Kind() == reflect.Ptr {
				et.Elem()
			}
			newElem := reflect.New(et)
			if err := UnmarshalResourceObject(obj, d.Included, newElem.Interface()); err != nil {
				return err
			}
			vve.Set(reflect.Append(vve, newElem))
		}
	} else {
		if err := UnmarshalResourceObject(d.Data.(*ResourceObject), d.Included, v); err != nil {
			return err
		}
	}
	return nil
}

func (d *Document) MarshalJSON() ([]byte, error) {
	type document Document
	if d.Data != nil {
		var err error
		d.RawData, err = json.Marshal(d.Data)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal((*document)(d))
}

func (d *Document) UnmarshalJSON(b []byte) error {
	type document Document
	if len(b) == 0 {
		return nil
	}
	if err := json.Unmarshal(b, (*document)(d)); err != nil {
		return err
	}
	if d.RawData == nil {
		return nil
	}
	if bytes.HasPrefix(d.RawData, []byte("[")) {
		s := make([]*ResourceObject, 0)
		if err := json.Unmarshal(d.RawData, &s); err != nil {
			return err
		}
		d.Data = s
	} else if bytes.HasPrefix(d.RawData, []byte("{")) {
		o := new(ResourceObject)
		if err := json.Unmarshal(d.RawData, o); err != nil {
			return err
		}
		d.Data = o
	}
	return nil
}

func (r *RelationshipObject) MarshalJSON() ([]byte, error) {
	type rObject RelationshipObject
	if r.Data != nil {
		var err error
		r.RawData, err = json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal((*rObject)(r))
}

func (r *RelationshipObject) UnmarshalJSON(b []byte) error {
	type rObject RelationshipObject
	if err := json.Unmarshal(b, (*rObject)(r)); err != nil {
		return err
	}

	if bytes.HasPrefix(r.RawData, []byte("[")) {
		data := make([]*ResourceIdentifierObject, 0, 10)
		if err := json.Unmarshal(r.RawData, &data); err != nil {
			return err
		}
		r.Data = data
		return nil
	}
	if r.RawData == nil || string(r.RawData) == "null" {
		return nil
	}
	r.Data = new(ResourceIdentifierObject)
	return json.Unmarshal(r.RawData, r.Data)
}

func HasLID(id *ResourceIdentifierObject) bool {
	return id.ID == "" && id.LID != ""
}

func ForEachElem[T any](v any, elemFunc func(e T) error) error {
	if v == nil {
		return nil
	}
	vv := reflectx.ValueOf(v, true)
	if vv.Kind() != reflect.Slice {
		return elemFunc(v.(T))
	}
	for i := 0; i < vv.Len(); i++ {
		e := vv.Index(i).Interface()
		err := elemFunc(e.(T))
		if err != nil {
			return err
		}
	}
	return nil
}

type DocumentDataItem[T any] struct {
	Data  T
	Links map[string]any
	Meta  MetaData
}

type DocumentData[T any] struct {
	Items    []*DocumentDataItem[T]
	Page     *pagination.Page
	IsSlice  bool
	MetaData MetaData
	Includes []*ResourceObject
}

func NewDocumentData[T any](v any, self string) *DocumentData[T] {
	rv := reflectx.ValueOf(v, true)
	if rv.Kind() == reflect.Slice {
		dataItems := make([]*DocumentDataItem[T], rv.Len())
		for i, vi := range v.([]T) {
			dataItems[i] = &DocumentDataItem[T]{
				Data:  vi,
				Links: map[string]any{"self": self},
			}
			if len(self) == 0 {
				continue
			}
			id, ok := identifier(vi)
			if !ok {
				continue
			}
			dataItems[i].Links = map[string]any{"self": joinUrl(self, id.ID)}
		}
		return &DocumentData[T]{
			Items:   dataItems,
			IsSlice: true,
		}
	}
	items := []*DocumentDataItem[T]{
		{Data: v.(T)},
	}

	if len(self) > 0 {
		id, ok := identifier(v.(T))
		if ok {
			items[0].Links = map[string]any{"self": joinUrl(self, id.ID)}
		}
	}
	return &DocumentData[T]{
		Items: items,
	}
}

func joinUrl(base string, p ...string) string {
	r, _ := url.JoinPath(base, p...)
	return r
}

func (d *DocumentData[T]) Single() T {
	var item T
	if len(d.Items) > 0 {
		item = d.Items[0].Data
	}
	return item
}

func (d *DocumentData[T]) Raw() any {
	if d == nil {
		return nil
	}
	datas := make([]any, len(d.Items))
	if !d.IsSlice {
		return d.Single()
	}
	for i, item := range d.Items {
		datas[i] = item.Data
	}
	return datas
}
