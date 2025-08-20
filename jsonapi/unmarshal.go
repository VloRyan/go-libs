package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vloryan/go-libs/reflectx"
)

func UnmarshalResourceObject(obj *ResourceObject, includes []*ResourceObject, v any) error {
	dest, ok := v.(ResourceIdentifierDestination)
	if !ok {
		return fmt.Errorf("%s does not implement ResourceIdentifierDestination", reflect.TypeOf(v).String())
	}
	dest.SetIdentifier(&ResourceIdentifierObject{ID: obj.ID, Type: obj.Type})
	if err := mapAttribs(obj.Attributes, v); err != nil {
		return err
	}
	if err := mapRelationships(obj.Relationships, includes, v); err != nil {
		return err
	}
	return nil
}

func mapAttribs(attribs map[string]any, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v is no pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("mapAttribs only supports Structs, got: " + rv.String())
	}

	for name, attrValue := range attribs {

		field := reflectx.FindField(v, name)
		if !field.IsValid() {
			continue
		}
		ft := reflectx.DeRef(field.Type())
		at := reflect.TypeOf(attrValue)
		value := attrValue

		var unmarshaler json.Unmarshaler
		if field.Type().Kind() == reflect.Ptr {
			if attrValue == nil {
				field.SetZero()
				continue
			}
			field.Set(reflect.New(ft))
			unmarshaler, _ = field.Interface().(json.Unmarshaler)
		} else {
			if attrValue == nil {
				return errors.New("can not set nil value to non-pointer field " + name)
			}
			unmarshaler, _ = field.Addr().Interface().(json.Unmarshaler)
		}
		if unmarshaler != nil {
			if ft == reflect.TypeOf(time.Time{}) && at.Kind() == reflect.String {
				if sf, found := reflectx.TypeOf(v, true).FieldByNameFunc(func(s string) bool {
					return strings.EqualFold(s, name)
				}); found {
					if err := reflectx.SetTimeField(value.(string), sf, field); err != nil {
						return err
					}
				} else {
					b, err := json.Marshal(attrValue)
					if err != nil {
						return err
					}
					if err := unmarshaler.UnmarshalJSON(b); err != nil {
						return err
					}
				}
			} else {
				b, err := json.Marshal(attrValue)
				if err != nil {
					return err
				}
				if err := unmarshaler.UnmarshalJSON(b); err != nil {
					return err
				}
			}
			continue
		}

		if at.Kind() == reflect.Map {
			if ft.Kind() == reflect.Map {
				if ft.Key().Kind() != at.Key().Kind() {
					return errors.New("mapAttribs only supports maps with same key type, got: " + ft.Key().String() + " expected: " + at.Key().String() + ")")
				}
				mv := reflect.MakeMap(ft)
				av := reflect.ValueOf(attrValue)
				for _, key := range av.MapKeys() {
					ev := reflect.ValueOf(av.MapIndex(key).Interface()) // for some reason valueOf(interface()) is needed to get the correct type
					var nv reflect.Value
					if ev.Type().Kind() == reflect.Map {
						if reflectx.DeRef(ft.Elem()).Kind() != reflect.Struct {
							return fmt.Errorf("attributes map value of type %s can not be converted to %s", ev.Type().String(), ft.Elem().String())
						}
						value = reflect.New(reflectx.DeRef(ft.Elem())).Interface()
						if err := mapAttribs(ev.Interface().(map[string]any), value); err != nil {
							return err
						}
						nv = reflect.ValueOf(value)

					} else {
						var ok bool
						if nv, ok = reflectx.Convert(ev, ft.Elem()); !ok {
							return fmt.Errorf("attributes map value of type %s can not be converted to %s", ev.Type().String(), ft.Elem().String())
						}
					}
					mv.SetMapIndex(key, nv)
				}
				if err := reflectx.SetFieldValue(field, mv.Interface()); err != nil {
					return err
				}

				continue
			}
			if ft.Kind() != reflect.Struct {
				return errors.New("mapAttribs only supports Structs: " + ft.String())
			}
			value = reflect.New(ft).Interface()
			if err := mapAttribs(attrValue.(map[string]any), value); err != nil {
				return err
			}
		}
		if at.Kind() == reflect.Slice {
			sv := reflect.ValueOf(value)
			var firstElem any
			if sv.Len() > 0 {
				firstElem = sv.Index(0).Interface()
			}
			if firstElem != nil && reflect.TypeOf(firstElem).Kind() == reflect.Map {
				slice := reflect.MakeSlice(ft, 0, sv.Len())
				for i := 0; i < sv.Len(); i++ {
					elem := sv.Index(i).Interface()
					elemValue := reflect.New(reflectx.DeRef(ft.Elem()))
					e := elemValue.Interface()
					if unmarshaler, ok := e.(json.Unmarshaler); ok {
						b, err := json.Marshal(elem)
						if err != nil {
							return err
						}
						if err := unmarshaler.UnmarshalJSON(b); err != nil {
							return err
						}
					} else {
						if elem != nil && reflect.TypeOf(elem).Kind() == reflect.Map {
							if err := mapAttribs(elem.(map[string]any), e); err != nil {
								return err
							}
						} else {
							return fmt.Errorf("unsupported element type %s", reflect.TypeOf(elem).String())
						}
					}

					slice = reflect.Append(slice, elemValue)
				}
				value = slice.Interface()
			}
		}
		if err := reflectx.SetFieldValue(field, value); err != nil {
			return err
		}
	}
	return nil
}

func mapRelationships(relations map[string]*RelationshipObject, includes []*ResourceObject, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v is no pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("unmarshal only supports structs")
	}

	for name, rel := range relations {
		if rel == nil {
			return errors.New("invalid relationship: " + name)
		}
		field := reflectx.FindField(v, name)
		if !field.IsValid() {
			continue
		}
		if rel.Data == nil {
			if err := reflectx.SetFieldValue(field, nil); err != nil {
				return err
			}
			continue
		}

		if reflect.TypeOf(rel.Data).Kind() == reflect.Slice {
			for _, elem := range rel.Data.([]*ResourceIdentifierObject) {
				ft := field.Type().Elem()
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				newInstance := reflect.New(ft)
				newInstanceDest, ok := newInstance.Interface().(ResourceIdentifierDestination)
				if !ok {
					return fmt.Errorf("%s does not implement ResourceIdentifierDestination", reflect.TypeOf(v).String())
				}
				newInstanceDest.SetIdentifier(&ResourceIdentifierObject{ID: elem.ID, Type: elem.Type})
				inc := FindIncluded(includes, elem.ID, elem.Type)
				if inc != nil {
					if err := mapAttribs(inc.Attributes, newInstance.Interface()); err != nil {
						return err
					}
				}
				if err := reflectx.SetFieldValue(field, reflect.Append(field, newInstance).Interface()); err != nil {
					return err
				}
			}
		} else {
			ft := field.Type()
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			newValue := reflect.New(ft).Interface()
			newInstanceDest, ok := newValue.(ResourceIdentifierDestination)
			if !ok {
				return fmt.Errorf("%s does not implement ResourceIdentifierDestination", reflect.TypeOf(v).String())
			}
			newInstanceDest.SetIdentifier(rel.Data.(*ResourceIdentifierObject))
			inc := FindIncluded(includes, rel.Data.(*ResourceIdentifierObject).ID, rel.Data.(*ResourceIdentifierObject).Type)
			if inc != nil {
				if err := mapAttribs(inc.Attributes, newValue); err != nil {
					return err
				}
			}
			if err := reflectx.SetFieldValue(field, newValue); err != nil {
				return err
			}
		}

	}
	return nil
}

func FindIncluded(includes []*ResourceObject, id, objType string) *ResourceObject {
	for _, obj := range includes {
		if obj.ID == id && obj.Type == objType {
			return obj
		}
	}
	return nil
}
