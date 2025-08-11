package reflectx

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func FindField(v any, path string) reflect.Value {
	if v == nil {
		return reflect.Value{}
	}
	rv := ValueOf(v, true)

	if rv.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	pathParts := strings.Split(path, ".")
	pathLen := len(pathParts)
	f := rv
	for i, part := range pathParts {
		var elemIdentifier string
		if idx := strings.Index(part, "["); idx != -1 {
			elemIdentifier = part[idx+1 : strings.Index(part, "]")]
			part = part[:idx]
		}
		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				return reflect.Value{}
			}
			f = f.Elem()
		}
		if f.Kind() != reflect.Struct {
			return reflect.Value{}
		}
		f = f.FieldByNameFunc(func(s string) bool {
			return strings.EqualFold(s, part)
		})
		if !f.IsValid() {
			return reflect.Value{}
		}
		if elemIdentifier != "" {
			if f.Kind() == reflect.Slice {
				index, err := strconv.Atoi(elemIdentifier)
				if err != nil {
					return reflect.Value{}
				}
				if index >= f.Len() {
					return reflect.Value{}
				}
				f = f.Index(index)
			}
			if f.Kind() == reflect.Map {
				f = f.MapIndex(reflect.ValueOf(elemIdentifier))
			}
		}
		if i+1 == pathLen {
			return f
		}
	}
	return reflect.Value{}
}

func FindFieldFunc(v any, matchFunc func(field reflect.StructField) bool) reflect.Value {
	rv := ValueOf(v, true)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return rv
	}
	tv := rv.Type()
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	for i := 0; i < tv.NumField(); i++ {
		field := tv.Field(i)
		fieldValue := rv.Field(i)
		if match := matchFunc(field); match {
			return fieldValue
		}
		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			matchField := FindFieldFunc(fieldValue.Addr().Interface(), matchFunc)
			if matchField.IsValid() {
				return matchField
			}
			continue
		}
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && field.Anonymous {
			matchField := FindFieldFunc(fieldValue.Addr().Interface(), matchFunc)
			if matchField.IsValid() {
				return matchField
			}
			continue
		}
	}
	return reflect.Value{}
}

func SetFieldValue(field reflect.Value, value any) error {
	if !field.CanSet() {
		return errors.New("field can not be set")
	}
	if value == nil {
		field.SetZero()
		return nil
	}
	rv := reflect.ValueOf(value)
	if setFieldValue(field, rv) {
		return nil
	}

	if rv.Type().Kind() == reflect.Ptr && !rv.IsNil() && field.Type().Kind() != reflect.Ptr {
		if setFieldValue(field, rv.Elem()) {
			return nil
		}
	}

	if field.Kind() == reflect.Ptr && rv.Type().Kind() != reflect.Ptr {
		field.Set(reflect.New(field.Type().Elem()))
		if setFieldValue(field.Elem(), rv) {
			return nil
		}
	}

	return errors.New("unknown conversion from type '" + rv.Type().String() + "' to set field of type '" + field.Type().String())
}

func setFieldValue(field reflect.Value, rv reflect.Value) bool {
	if rv.Type() == field.Type() {
		field.Set(rv)
		return true
	}
	cv, ok := Convert(rv, field.Type())
	if !ok {
		return false
	}
	return setFieldValue(field, cv)
}

func Tag(s reflect.StructField, name string) FieldTag {
	return ParseTag(s.Tag.Get(name))
}

func SetTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	switch strings.ToLower(timeFormat) {
	case "dateonly":
		timeFormat = time.DateOnly
	case "timeonly":
		timeFormat = time.TimeOnly
	case "datetime":
		timeFormat = time.DateTime
	case "":
		timeFormat = time.RFC3339
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	t, err := time.Parse(timeFormat, val)
	if err != nil {
		return err
	}
	if value.Type().Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value.Elem().Set(reflect.ValueOf(t))
	} else {
		value.Set(reflect.ValueOf(t))
	}
	return nil
}
