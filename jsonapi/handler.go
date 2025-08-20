package jsonapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/vloryan/go-libs/httpx"
)

type (
	ResolveObjectWithReqFunc func(req *http.Request, id *ResourceIdentifierObject) (*ResourceObject, *Error)
	resolveObjectFunc        func(id *ResourceIdentifierObject) (*ResourceObject, *Error)
	BeforeWriteFunc[T any]   func(req *http.Request, data *DocumentData[T], doc *Document) error
	GenericHandler[T any]    struct {
		ResolveObjectWithReqFunc ResolveObjectWithReqFunc
		BeforeWriteFunc          BeforeWriteFunc[T]
		FieldFilterFunc          ResourceObjectFieldFilterFunc
		DocumentUpdaters         []DocumentUpdater
	}
	DocumentUpdater interface {
		Update(doc *Document) error
	}
)

var SparseFieldSetFilter = func(sparseFieldset map[string][]string) ResourceObjectFieldFilterFunc {
	return func(typeName string, fieldName string) bool {
		if len(sparseFieldset) == 0 {
			return true
		}
		if fields, ok := sparseFieldset[typeName]; ok {
			return slices.ContainsFunc(fields, func(s string) bool {
				return strings.EqualFold(s, fieldName)
			})
		}
		return true
	}
}

func (h *GenericHandler[T]) Handle(f func(req *http.Request) (*DocumentData[T], *Error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != MediaType {
			w.Header().Set("Content-Type", MediaType+"; charset=utf-8")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, _ = w.Write([]byte(http.StatusText(http.StatusUnsupportedMediaType)))
			return
		}
		data, jErr := f(req)
		if jErr != nil {
			_ = Write(w, h.NewErrorDoc(jErr))
			return
		}
		if data == nil {
			_ = Write(w, nil)
			return
		}
		doc := h.NewDoc(req, data)
		if len(doc.Errors) > 0 {
			_ = Write(w, doc)
			return
		}
		for _, updater := range h.DocumentUpdaters {
			if err := updater.Update(doc); err != nil {
				_ = Write(w, h.NewErrorDoc(err))
				return
			}
		}
		if h.BeforeWriteFunc != nil {
			if err := h.BeforeWriteFunc(req, data, doc); err != nil {
				_ = Write(w, h.NewErrorDoc(err))
				return
			}
		}
		_ = Write(w, doc)
	}
}

func (h *GenericHandler[T]) NewDoc(req *http.Request, data *DocumentData[T]) *Document {
	sparseFieldSets := make(map[string][]string)
	if fieldSets, exist := httpx.QueryFamily(req, "fields"); exist {
		for k, v := range fieldSets {
			fields := strings.Split(v, ",")
			sparseFieldSets[k] = fields
		}
	}

	doc := NewDocument()
	if err := doc.SetObjectDataFieldFilterFunc(data.Raw(), func(typeName string, fieldName string) bool {
		if !SparseFieldSetFilter(sparseFieldSets)(typeName, fieldName) {
			return false
		}
		if h.FieldFilterFunc != nil {
			return h.FieldFilterFunc(typeName, fieldName)
		}
		return true
	}); err != nil {
		return h.NewErrorDoc(err)
	}

	for i, item := range data.Items {
		var obj *ResourceObject
		if data.IsSlice {
			obj = doc.Data.([]*ResourceObject)[i]
		} else {
			obj = doc.Data.(*ResourceObject)
		}
		if obj.Meta == nil {
			obj.Meta = make(MetaData)
		}
		for k, v := range item.Meta {
			obj.Meta[k] = v
		}

		obj.Links = item.Links
	}

	resolverWithReqFunc := h.ResolveObjectWithReqFunc
	if resolverWithReqFunc != nil {
		resolverFunc := func(id *ResourceIdentifierObject) (*ResourceObject, *Error) {
			return resolverWithReqFunc(req, id)
		}
		includesStr := httpx.Query(req, "include")
		if includesStr != "" {
			includes := strings.Split(includesStr, ",")
			if err := h.resolveIncludes(resolverFunc, includes, doc); err != nil {
				return h.NewErrorDoc(err)
			}
		}
	}
	/* clean up metadata */
	if err := ForEachElem(doc.Data, func(e *ResourceObject) error {
		e.SetLocalObjects(nil)
		return nil
	}); err != nil {
		return h.NewErrorDoc(err)
	}

	h.applyMetadata(doc, data)

	return doc
}

func (h *GenericHandler[T]) NewErrorDoc(err error) *Document {
	doc := NewDocument()
	var jErr *Error
	if errors.As(err, &jErr) {
		doc.AddError(jErr)
	} else {
		doc.AddError(NewError(http.StatusInternalServerError, err.Error(), err))
	}
	return doc
}

func (h *GenericHandler[T]) resolveIncludes(resolverFunc resolveObjectFunc, includes []string, doc *Document) error {
	resolvedCache := make(map[string]*ResourceObject)
	var resolvedObjects []*ResourceObject
	relationships := doc.Relationships()
	if err := ForEachElem(doc.Data, func(e *ResourceObject) error {
		for k, v := range e.LocalObjects() {
			resolvedCache[k] = v
		}
		return nil
	}); err != nil {
		return err
	}
	for _, include := range includes {
		objs, err := h.resolveInclude(resolverFunc, relationships, include, resolvedCache)
		if err != nil {
			return err
		}
		if len(objs) > 0 {
			resolvedObjects = append(resolvedObjects, objs...)
		}
	}
	resolvedObjectSet := make([]*ResourceObject, 0, len(resolvedObjects))
	for _, obj := range resolvedObjects {
		if slices.Contains(resolvedObjectSet, obj) {
			continue
		}
		resolvedObjectSet = append(resolvedObjectSet, obj)
	}
	doc.Included = resolvedObjectSet
	return nil
}

func (h *GenericHandler[T]) resolveInclude(resolverFunc resolveObjectFunc, relationships map[string][]*ResourceIdentifierObject, include string, cache map[string]*ResourceObject) ([]*ResourceObject, *Error) {
	var resolvedObjects []*ResourceObject
	relations, matchedPath := findRelationships(relationships, include)
	if matchedPath == "" {
		// TODO: empty relationships are missing. Eg. person without address - relationships is empty
		// err := errors.New("relationship with name '" + include + "' not found")
		// return nil, jsonapi.NewError(400, "include failed", err)
		return resolvedObjects, nil
	}
	nextPart := ""
	if len(include) > len(matchedPath)+1 {
		nextPart = include[len(matchedPath)+1:]
	}
	if err := ForEachElem(relations, func(e *ResourceIdentifierObject) error {
		var cacheKey string
		if e.ID != "" {
			cacheKey = e.ID + e.Type
		} else {
			if e.LID == "" {
				return NewError(http.StatusInternalServerError, "relation has no id", nil)
			}
			cacheKey = e.LID
		}
		var resolvedObj *ResourceObject
		if cachedObj, found := cache[cacheKey]; found {
			resolvedObj = cachedObj
		} else {
			var err *Error
			resolvedObj, err = resolverFunc(e)
			if err != nil {
				return err
			}
			if resolvedObj != nil {
				cache[cacheKey] = resolvedObj
			}
		}
		if resolvedObj == nil {
			return nil
		}
		resolvedObjects = append(resolvedObjects, resolvedObj)

		if nextPart != "" {
			flatResIds := make(map[string][]*ResourceIdentifierObject)
			for k, v := range resolvedObj.Relationships {
				if reflect.TypeOf(v.Data).Kind() == reflect.Slice {
					flatResIds[k] = v.Data.([]*ResourceIdentifierObject)
				} else {
					flatResIds[k] = []*ResourceIdentifierObject{v.Data.(*ResourceIdentifierObject)}
				}
			}

			objs, err := h.resolveInclude(resolverFunc, flatResIds, nextPart, cache)
			if err != nil {
				return err
			}
			if len(objs) > 0 {
				resolvedObjects = append(resolvedObjects, objs...)
			}
		}
		return nil
	}); err != nil {
		return nil, NewError(500, "include failed", err)
	}
	return resolvedObjects, nil
}
func (h *GenericHandler[T]) applyMetadata(doc *Document, data *DocumentData[T]) {
	for k, v := range data.MetaData {
		doc.Meta[k] = v
	}
	if data.Page != nil {
		doc.Meta["page[limit]"] = data.Page.Limit
		doc.Meta["page[offset]"] = data.Page.Offset
		doc.Meta["page[sort]"] = strings.Join(data.Page.Sort, ",")
		doc.Meta["page[total]"] = data.Page.TotalCount
	}
}
func findRelationships(relationships map[string][]*ResourceIdentifierObject, name string) ([]*ResourceIdentifierObject, string) {
	currentRelationships := relationships
	var nextRelationships map[string][]*ResourceIdentifierObject
	matchedPart := ""
	for _, part := range strings.Split(name, ".") {
		nextRelationships = findNextRelationships(currentRelationships, part)
		if len(nextRelationships) == 0 {
			var result []*ResourceIdentifierObject
			for _, v := range currentRelationships {
				result = append(result, v...)
			}
			return result, matchedPart
		}
		if matchedPart != "" {
			matchedPart += "."
		}
		matchedPart += part
		currentRelationships = nextRelationships
	}
	if len(nextRelationships) > 0 {
		var result []*ResourceIdentifierObject
		for _, v := range currentRelationships {
			result = append(result, v...)
		}
		return result, matchedPart
	}
	return nil, ""
}

func findNextRelationships(relationships map[string][]*ResourceIdentifierObject, name string) map[string][]*ResourceIdentifierObject {
	matches := make(map[string][]*ResourceIdentifierObject)
	part := name
	if idx := strings.Index(part, "."); idx != -1 {
		part = part[:idx]
	}
	for k, v := range relationships {
		kPart := k
		kNext := ""
		if idx := strings.Index(kPart, "."); idx != -1 {
			if len(kPart) > idx {
				kNext = kPart[idx+1:]
			}
			kPart = kPart[:idx]
		}
		if kPart == part {
			matches[kNext] = v
			continue
		}
		elemIdx := strings.Index(kPart, "[")
		if elemIdx == -1 {
			continue
		}
		kWithoutIndex := kPart[:elemIdx] + kPart[strings.Index(k, "]")+1:]
		if kWithoutIndex == part {
			matches[kNext] = v
			continue
		}
	}
	return matches
}

func Write(writer http.ResponseWriter, doc *Document) error {
	writer.Header().Set("Content-Type", MediaType+"; charset=utf-8")
	if doc == nil {
		writer.WriteHeader(http.StatusNoContent)
		return nil
	}
	status := http.StatusOK
	for _, apiErr := range doc.Errors {
		apiStatus, _ := strconv.ParseInt(apiErr.Status, 10, 64)
		if int(apiStatus) > status {
			status = int(apiStatus)
		}
		log.Err(apiErr)
	}
	writer.WriteHeader(status)
	return writeJSON(writer, doc)
}

func writeJSON(rw http.ResponseWriter, data any) error {
	if data == nil {
		return nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = rw.Write(b)
	return err
}
