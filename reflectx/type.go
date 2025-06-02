package reflectx

import "reflect"

// ElemTypeOf returns the same as reflect.TypeOf. If the type's Kind is Array, Chan, Map, Pointer, or Slice it returns the type's element type.
func ElemTypeOf(v any, resolvePointedType bool) reflect.Type {
	t := TypeOf(v, resolvePointedType)
	switch t.Kind() {
	case reflect.Ptr, reflect.Array, reflect.Map,
		reflect.Slice, reflect.Chan:
		return DeRef(t.Elem())
	default:
		return t
	}
}

func DeRef(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	for t.Kind() == reflect.Ptr {
		return DeRef(t.Elem())
	}
	return t
}

func DeRefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		return DeRefValue(v.Elem())
	}
	return v
}

func KindElem(v reflect.Value) reflect.Kind {
	for v.Kind() == reflect.Ptr {
		return KindElem(v.Elem())
	}
	return v.Kind()
}

// TypeOf returns the same as reflect.TypeOf. If the type's Kind is Pointer it returns the type's element type.
func TypeOf(v any, resolvePointedType bool) reflect.Type {
	t := reflect.TypeOf(v)
	if resolvePointedType {
		return DeRef(t)
	}
	return t
}

// ElemValueOf returns the same as reflect.ValueOf. If the type's Kind is Pointer or Interface it returns value that the interface v contains or that the pointer v points to.
func ElemValueOf(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return rv.Elem()
	default:
		return rv
	}
}

// ValueOf returns the same as reflect.ValueOf. If the type's Kind is Pointer it returns the value that the pointer v points to.
func ValueOf(v any, resolvePointedType bool) reflect.Value {
	rv := reflect.ValueOf(v)
	if resolvePointedType && rv.Kind() == reflect.Ptr {
		for rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
	}
	return rv
}
