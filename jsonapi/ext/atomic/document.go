package atomic

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/vloryan/go-libs/jsonapi"
)

const MediaType = "application/vnd.api+json;ext=\"https://jsonapi.org/ext/atomic\""

type Document struct {
	Operations []*Operation     `json:"atomic:operations,omitempty"`
	Results    []*Result        `json:"atomic:results,omitempty"`
	Errors     []*jsonapi.Error `json:"errors,omitempty"`
	Meta       jsonapi.MetaData `json:"meta,omitempty"`

	JSONAPI *jsonapi.Object `json:"jsonapi,omitempty"`
	Links   map[string]any  `json:"links,omitempty"`

	baseURL  string
	lidCache map[string][]*jsonapi.ResourceIdentifierObject
}

func (d *Document) UnmarshalJSON(b []byte) error {
	type docType Document
	if err := json.Unmarshal(b, (*docType)(d)); err != nil {
		return err
	}
	return d.buildLIDCache()
}

type Operation struct {
	Op      string                            `json:"op"`
	Ref     *jsonapi.ResourceIdentifierObject `json:"ref,omitempty"`
	HRef    string                            `json:"href,omitempty"`
	Data    any                               `json:"-"`
	RawData json.RawMessage                   `json:"data,omitempty"`
	Meta    jsonapi.MetaData                  `json:"meta,omitempty"`
}

type Result struct {
	Data    any              `json:"-"`
	RawData json.RawMessage  `json:"data,omitempty"`
	Meta    jsonapi.MetaData `json:"meta,omitempty"`
}

func (o *Operation) MarshalJSON() ([]byte, error) {
	type elem Operation
	if o.Data != nil {
		var err error
		o.RawData, err = json.Marshal(o.Data)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal((*elem)(o))
}

func (o *Operation) UnmarshalJSON(b []byte) error {
	type elem Operation
	if err := json.Unmarshal(b, (*elem)(o)); err != nil {
		return err
	}
	if bytes.HasPrefix(o.RawData, []byte("[")) {
		o.Data = make([]*jsonapi.ResourceObject, 0, 10)
	}
	if bytes.HasPrefix(o.RawData, []byte("{")) {
		o.Data = new(jsonapi.ResourceObject)
	}
	if o.RawData == nil {
		return nil
	}
	return json.Unmarshal(o.RawData, o.Data)
}

func (r *Result) MarshalJSON() ([]byte, error) {
	type elem Result
	if r.Data != nil {
		var err error
		r.RawData, err = json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal((*elem)(r))
}

func (r *Result) UnmarshalJSON(b []byte) error {
	type elem Result
	if err := json.Unmarshal(b, (*elem)(r)); err != nil {
		return err
	}
	if bytes.HasPrefix(r.RawData, []byte("[")) {
		r.Data = make([]*jsonapi.ResourceObject, 0, 10)
	}
	if bytes.HasPrefix(r.RawData, []byte("{")) {
		r.Data = new(jsonapi.ResourceObject)
	}
	if r.RawData == nil {
		return nil
	}
	return json.Unmarshal(r.RawData, r.Data)
}

func NewDocument(baseURL string) *Document {
	return &Document{
		Meta: make(jsonapi.MetaData),
		JSONAPI: &jsonapi.Object{
			Version: jsonapi.Version,
		},
		baseURL: baseURL,
	}
}

func (d *Document) AddError(errors ...*jsonapi.Error) *Document {
	d.Errors = append(d.Errors, errors...)
	return d
}

func (d *Document) buildLIDCache() error {
	lidCache := make(map[string][]*jsonapi.ResourceIdentifierObject)
	for _, operation := range d.Operations {
		v := reflect.ValueOf(operation.Data)
		if !v.IsValid() || v.IsNil() {
			continue
		}
		if v.Type().Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				if err := d.populateCacheForObject(lidCache, v.Index(i).Interface().(*jsonapi.ResourceObject)); err != nil {
					return err
				}
			}
		} else {
			if err := d.populateCacheForObject(lidCache, operation.Data.(*jsonapi.ResourceObject)); err != nil {
				return err
			}
		}
	}
	d.lidCache = lidCache
	return nil
}

func (d *Document) populateCacheForObject(cache map[string][]*jsonapi.ResourceIdentifierObject, object *jsonapi.ResourceObject) error {
	if jsonapi.HasLID(&object.ResourceIdentifierObject) {
		objList := cache[object.LID]
		objList = append(objList, &object.ResourceIdentifierObject)
		cache[object.LID] = objList
	}
	for _, rel := range object.Relationships {
		v := reflect.ValueOf(rel.Data)
		if !v.IsValid() || v.IsNil() {
			continue
		}
		if v.Type().Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				objId := v.Index(i).Interface().(*jsonapi.ResourceIdentifierObject)
				if jsonapi.HasLID(objId) {
					objList := cache[objId.LID]
					objList = append(objList, objId)
					cache[objId.LID] = objList
				}
			}
		} else {
			objId := rel.Data.(*jsonapi.ResourceIdentifierObject)
			if jsonapi.HasLID(objId) {
				objList := cache[objId.LID]
				objList = append(objList, objId)
				cache[objId.LID] = objList
			}
		}
	}

	return nil
}

func (d *Document) UpdateLID(lid string, id string) bool {
	ids, ok := d.lidCache[lid]
	if !ok {
		return false
	}
	for _, _id := range ids {
		_id.ID = id
	}
	return ok
}

func (d *Document) LIDCache() map[string]string {
	cache := make(map[string]string)
	for k, v := range d.lidCache {
		cache[k] = v[0].ID
	}
	return cache
}

func (o *Operation) LIDs() []string {
	lids := make([]string, 0, 10)

	v := reflect.ValueOf(o.Data)
	if !v.IsValid() || v.IsNil() {
		return lids
	}
	if v.Type().Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			object := v.Index(i).Interface().(*jsonapi.ResourceObject)
			if jsonapi.HasLID(&object.ResourceIdentifierObject) {
				lids = append(lids, object.LID)
			}
		}
	} else {
		object := o.Data.(*jsonapi.ResourceObject)
		if jsonapi.HasLID(&object.ResourceIdentifierObject) {
			lids = append(lids, object.LID)
		}
	}
	return lids
}
