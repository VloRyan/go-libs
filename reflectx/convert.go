package reflectx

import (
	"reflect"
	"strconv"
	"strings"
)

func Convert(v reflect.Value, t reflect.Type) (reflect.Value, bool) {
	if v.Type() == t {
		return v, true
	}
	if v.Type().ConvertibleTo(t) {
		return v.Convert(t), true
	}
	if v.Kind() == reflect.String && isNumeric(t.Kind()) {
		numVal, ok := convertToNumber(v.String(), t.Kind())
		if ok {
			if numVal.Type().ConvertibleTo(t) {
				return numVal.Convert(t), true
			}
		}
	}
	if t.Kind() == reflect.Bool {
		if v.CanInt() {
			value := v.Int()
			return reflect.ValueOf(value != 0), true
		}
		if v.CanUint() {
			value := v.Uint()
			return reflect.ValueOf(value != 0), true
		}
		if v.CanFloat() {
			value := v.Float()
			return reflect.ValueOf(value != 0.0), true
		}
		if v.Kind() == reflect.String {
			value := strings.ToLower(v.String())
			if strings.EqualFold(value, "true") ||
				strings.EqualFold(value, "1") ||
				strings.EqualFold(value, "on") {
				return reflect.ValueOf(true), true
			}
			if strings.EqualFold(value, "false") ||
				strings.EqualFold(value, "0") ||
				strings.EqualFold(value, "off") {
				return reflect.ValueOf(true), true
			}
		}
	}
	return reflect.Value{}, false
}

func convertToNumber(s string, kind reflect.Kind) (reflect.Value, bool) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iVal, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(iVal), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uVal, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(uVal), true
	case reflect.Float32, reflect.Float64:
		fVal, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(fVal), true
	default:
		return reflect.Value{}, false
	}
}

func isNumeric(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
