package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/vloryan/go-libs/reflectx"
	"github.com/vloryan/go-libs/stringx"
)

type FieldMarshaler interface {
	MarshalField(obj *ResourceObject, f reflect.StructField) error
}
type LIDGeneratorFunc func() string

var reservedAttribNames = []string{"id", "type"}

var LIDGenerator LIDGeneratorFunc = func() string {
	return strconv.Itoa(int(time.Now().UnixNano()))
}

func MarshalResourceObject(data any, fieldFilterFunc ResourceObjectFieldFilterFunc) (*ResourceObject, error) {
	if data == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(data)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("expected data to be a struct, got " + rv.Type().Name())
	}
	return toResourceObject(data, fieldFilterFunc)
}

func toResourceObject(v any, fieldFilterFunc ResourceObjectFieldFilterFunc) (*ResourceObject, error) {
	id, ok := identifier(v)
	if !ok {
		t := reflectx.TypeOf(v, true)
		return nil, errors.New(t.PkgPath() + "." + t.Name() + " does not implement ResourceIdentifierSource interface")
	}
	obj := NewResourceObject(id)
	rv := reflectx.ValueOf(v, true)

	if marshaller, ok := v.(ResourceObjectMarshaller); ok {
		err := marshaller.MarshalResourceObject(obj)
		return obj, err
	}

	if rv.Kind() != reflect.Struct {
		return nil, errors.New("marshal only supports structs")
	}

	err := mapStructFields(obj, rv, "", fieldFilterFunc)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func identifier(v any) (*ResourceIdentifierObject, bool) {
	vv := reflect.ValueOf(v)
	if v == nil || (vv.Type().Kind() == reflect.Ptr && vv.IsNil()) {
		return nil, false
	}
	identifier, ok := v.(ResourceIdentifierSource)
	if !ok {
		return nil, false
	}
	return identifier.GetIdentifier(), true
}

func mapStructFields(obj *ResourceObject, rv reflect.Value, path string, fieldNameFunc ResourceObjectFieldFilterFunc) error {
	t := rv.Type()
	resourceLidCounter := 0
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if path == "" && slices.Contains(reservedAttribNames, strings.ToLower(sf.Name)) ||
			fieldNameFunc != nil && !fieldNameFunc(obj.Type, sf.Name) {
			continue
		}
		jsonTag := reflectx.Tag(sf, "json")
		if jsonTag.Value == "-" {
			continue
		}
		fv := rv.FieldByName(sf.Name)
		if !fv.IsValid() {
			continue
		}
		ft := sf.Type
		if ft.Kind() == reflect.Ptr {
			if fv.IsNil() {
				if !jsonTag.Has("omitempty") {
					obj.Attributes[stringx.ToCamelCase(sf.Name)] = nil
				}
				continue
			}
			ft = ft.Elem()
		}
		if jsonTag.Has("omitempty") && fv.IsZero() {
			continue
		}
		if m, ok := fv.Interface().(FieldMarshaler); ok {
			if err := m.MarshalField(obj, sf); err != nil {
				return err
			}
			continue
		}

		switch ft.Kind() {
		case reflect.Struct:
			if sf.Anonymous {
				if err := mapStructFields(obj, fv, path, fieldNameFunc); err != nil {
					return err
				}
				continue
			} else if _, ok := identifier(fv.Interface()); ok {
				resObj, err := toResourceObject(fv.Interface(), fieldNameFunc)
				if err != nil {
					return err
				}
				if resObj.ID == "" {
					if obj.ID != "" {
						resObj.LID = obj.ID + "_" + strconv.Itoa(resourceLidCounter)
					} else {
						obj.LID = LIDGenerator()
						resObj.LID = obj.LID + "_" + strconv.Itoa(resourceLidCounter)
					}

					resourceLidCounter++
					localObject := obj.LocalObjects()
					if localObject == nil {
						localObject = make(map[string]*ResourceObject)
						obj.SetLocalObjects(localObject)
					}
					localObject[resObj.LID] = resObj
				}
				name := joinPath(path, attribName(sf))
				obj.Relationships[name] = &RelationshipObject{Data: &resObj.ResourceIdentifierObject}
				continue
			}

			name := fieldName(sf)
			isMarshaler := false
			if _, ok := fv.Interface().(json.Marshaler); ok {
				isMarshaler = true
			} else if _, ok := reflectx.DeRefValue(fv).Interface().(json.Marshaler); ok {
				isMarshaler = true
			}
			if isMarshaler {
				if err := setAttribute(obj.Attributes, joinPath(path, name), fv.Interface()); err != nil {
					return err
				}
				continue
			}
			newPath := joinPath(path, name)
			if err := setAttribute(obj.Attributes, newPath, make(map[string]any)); err != nil {
				return err
			}
			if err := mapStructFields(obj, reflectx.DeRefValue(fv), newPath, fieldNameFunc); err != nil {
				return err
			}
			continue

		case reflect.Slice:
			if fv.IsNil() {
				continue
			}
			if reflectx.DeRef(ft.Elem()).Kind() == reflect.Struct {
				name := fieldName(sf)
				newPath := joinPath(path, name)

				if err := setAttribute(obj.Attributes, newPath, make([]map[string]any, fv.Len())); err != nil {
					return err
				}
				for i := 0; i < fv.Len(); i++ {
					elem := fv.Index(i)
					elemPath := newPath + "[" + strconv.Itoa(i) + "]"
					if err := setAttribute(obj.Attributes, elemPath, make(map[string]any)); err != nil {
						return err
					}
					if err := mapStructFields(obj, reflectx.DeRefValue(elem), elemPath, fieldNameFunc); err != nil {
						return err
					}
				}
				continue
			}
		default:
			// do nothing
		}
		if err := setAttribute(obj.Attributes, joinPath(path, fieldName(sf)), fv.Interface()); err != nil {
			return err
		}
	}
	return nil
}

func fieldName(sf reflect.StructField) string {
	jsonTag := reflectx.Tag(sf, "json")
	if jsonTag.Value != "" {
		return jsonTag.Value
	}
	return stringx.ToCamelCase(sf.Name)
}

func joinPath(path, newElem string) string {
	if path == "" {
		return newElem
	}
	return path + "." + newElem
}

func setAttribute(m map[string]any, path string, v any) (err error) {
	elemIdx := -1
	dotIdx := strings.Index(strings.ToLower(path), ".")
	if dotIdx == -1 {
		if idx := strings.Index(path, "["); idx != -1 {
			s := path[idx+1 : strings.Index(path, "]")]
			elemIdx, err = strconv.Atoi(s)
			if err != nil {
				return err
			}
			mv, ok := m[path[:idx]]
			if !ok {
				return errors.New("unable to resolve " + path)
			}
			mvv := reflect.ValueOf(mv)
			if mvv.Kind() != reflect.Slice {
				return fmt.Errorf("elem of path %s is no slice: %s", path, reflect.TypeOf(v).String())
			}
			mvv.Index(elemIdx).Set(reflect.ValueOf(v))
			return nil
		}
		m[path] = v
		return nil
	}
	part := path[:dotIdx]

	if idx := strings.Index(part, "["); idx != -1 {
		s := part[idx+1 : strings.Index(part, "]")]
		elemIdx, err = strconv.Atoi(s)
		if err != nil {
			return err
		}
		part = part[:idx]
	}

	mv, ok := m[part]
	if !ok {
		return errors.New("unable to resolve " + path)
	}
	if elemIdx != -1 {
		if reflect.TypeOf(mv).Kind() != reflect.Slice {
			return fmt.Errorf("elem of path %s is no slice: %s", path, reflect.TypeOf(v).String())
		}
		return setAttribute(mv.([]map[string]any)[elemIdx], path[dotIdx+1:], v)
	}
	if reflect.TypeOf(mv).Kind() != reflect.Map {
		return fmt.Errorf("elem of path %s is no map: %s", path, reflect.TypeOf(v).String())
	}
	return setAttribute(mv.(map[string]any), path[dotIdx+1:], v)
}

func attribName(f reflect.StructField) string {
	jsonTag := strings.Split(f.Tag.Get("json"), ",")
	if jsonTag[0] != "" {
		return jsonTag[0]
	}
	return stringx.ToCamelCase(f.Name)
}

func NewResourceObject(id *ResourceIdentifierObject) *ResourceObject {
	return &ResourceObject{
		ResourceIdentifierObject: *id,
		Links:                    make(map[string]any),
		Attributes:               make(map[string]any),
		Relationships:            make(map[string]*RelationshipObject),
		Meta:                     make(MetaData),
	}
}
