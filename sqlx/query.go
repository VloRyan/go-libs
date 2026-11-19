package sqlx

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/vloryan/go-libs/reflectx"
	"github.com/vloryan/go-libs/stringx"
)

type sqlQueryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

func Select(q sqlQueryer, dest any, query string, args ...any) error {
	_query := query
	var paramArgs []any
	if len(args) > 0 {
		var names []string
		var err error
		_query, names, err = compileNamedQuery([]byte(query), '?')
		if err != nil {
			return err
		}
		if len(names) > 0 {
			paramArgs, err = extractParamArgs(args[0], names)
			if err != nil {
				return err
			}
		}
	}

	rows, err := q.Query(_query, paramArgs...)
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	return unmarshalRows(rows, dest)
}

func unmarshalRows(rows *sql.Rows, dest any) error {
	t := reflectx.TypeOf(dest, true)
	if t.Kind() == reflect.Slice {
		if reflect.TypeOf(dest).Kind() != reflect.Pointer {
			return errors.New("dest is no pointer to slice")
		}
		elemType := reflectx.ElemTypeOf(dest, true)
		if elemType.Kind() != reflect.Struct {
			return errors.New("dest is no slice of structs")
		}
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	columns := make([]any, len(cols))
	columnPointers := make([]any, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}
	for rows.Next() {
		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}
		switch t.Kind() {
		case reflect.Struct:
			if err := scanStruct(cols, columnPointers, dest); err != nil {
				return err
			}
			return nil
		case reflect.Slice:
			elemType := reflectx.ElemTypeOf(dest, true)
			elem := reflect.New(elemType)
			if err := scanStruct(cols, columnPointers, elem.Interface()); err != nil {
				return err
			}
			v := reflect.ValueOf(dest).Elem()
			v.Set(reflect.Append(v, elem))
		default:
			if mapper, ok := reflect.ValueOf(dest).Interface().(RowMapper); ok {
				m := make(map[string]any)
				for i, colName := range cols {
					m[colName] = *columnPointers[i].(*any)
				}
				if err := mapper.MapRow(m); err != nil {
					return err
				}
			} else {
				return errors.New("unable to scan rows for type " + reflect.TypeOf(dest).String())
			}
			continue
		}
	}
	return nil
}

func scanStruct(columnNames []string, columnPointers []any, dest any) error {
	for i, colName := range columnNames {
		val := columnPointers[i].(*any)
		if *val == nil {
			continue
		}
		field := walkPath(dest, strings.Split(colName, "."))
		if !field.IsValid() {
			continue
		}
		pField := field
		if pField.Kind() != reflect.Pointer {
			pField = field.Addr()
		}
		if scanner, ok := pField.Interface().(sql.Scanner); ok {
			if err := scanner.Scan(*val); err != nil {
				return err
			}
		} else {
			fieldType := reflectx.ElemTypeOf(field, true)
			if isProbablyJSON(*val) && (fieldType.Kind() == reflect.Struct || fieldType.Kind() == reflect.Map) {
				strValue := (*val).(string)
				newInstance := reflect.New(pField.Type())
				err := json.Unmarshal([]byte(strValue), newInstance.Interface())
				if err == nil {
					if err := reflectx.SetFieldValue(field, newInstance.Elem().Interface()); err != nil {
						return err
					}
					continue
				}
			}
			if err := reflectx.SetFieldValue(field, *val); err != nil {
				return err
			}
		}
	}
	return nil
}

func isProbablyJSON(v any) bool {
	valT := reflectx.ValueOf(v, true)
	if valT.Kind() == reflect.String {
		strValue := strings.TrimSpace(valT.String())
		if strings.HasPrefix(strValue, "{") && strings.HasSuffix(strValue, "}") {
			return true
		}
		if strings.HasPrefix(strValue, "[") && strings.HasSuffix(strValue, "]") {
			return true
		}
	}
	return false
}

func walkPath(v any, path []string) reflect.Value {
	currentField := reflect.ValueOf(v)
	for i := range path {
		field := reflectx.FindFieldFunc(currentField.Interface(), func(field reflect.StructField) bool {
			dbTag := reflectx.Tag(field, "db")
			if dbTag.Value != "" {
				if dbTag.Value == path[i] {
					return true
				}
				if len(dbTag.Opts) > 0 && dbTag.Opts[0] == path[i] {
					return true
				}
			}
			return strings.EqualFold(field.Name, stringx.ToCamelCase(path[i]))
		})

		if !field.IsValid() {
			return reflect.Value{}
		}
		if i < len(path)-1 {
			if field.Type().Kind() == reflect.Pointer && field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
		}
		currentField = field
	}
	return currentField
}

func Exec(q sqlQueryer, query string, args ...any) (sql.Result, error) {
	_query, names, err := compileNamedQuery([]byte(query), '?')
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return q.Exec(_query)
	}
	paramArgs, err := extractParamArgs(args[0], names)
	if err != nil {
		return nil, err
	}
	return q.Exec(_query, paramArgs...)
}

func extractParamArgs(obj any, names []string) (paramArgs []any, err error) {
	t := reflectx.TypeOf(obj, true)
	switch t.Kind() {
	case reflect.Struct:
		paramArgs, err = extractParamArgsFromStruct(obj, names)
	case reflect.Map:
		paramArgs, err = extractParamArgsFromMap(obj.(map[string]any), names)
	case reflect.Slice:
		paramArgs, err = extractParamArgsFromSlice(obj, names)
	default:
		return nil, errors.New("unsupported type '" + t.String() + "'")
	}
	if err != nil {
		return nil, err
	}
	if len(paramArgs) != len(names) {
		return nil, fmt.Errorf("failed to resolve %d fields", len(names)-len(paramArgs))
	}
	return paramArgs, err
}

func extractParamArgsFromSlice(s any, names []string) ([]any, error) {
	v := reflect.ValueOf(s)
	paramArgs := make([]any, 0, len(names))
	for idx := 0; idx < v.Len(); idx++ {
		value := v.Index(idx).Interface()
		var namesForIdx []string
		suffix := "[" + strconv.Itoa(idx) + "]"
		for _, name := range names {
			if strings.HasSuffix(name, suffix) {
				namesForIdx = append(namesForIdx, name[:len(name)-len(suffix)])
			} else if v.Len() == 1 {
				namesForIdx = append(namesForIdx, name)
			}
		}
		values, err := extractParamArgs(value, namesForIdx)
		if err != nil {
			return nil, err
		}
		paramArgs = append(paramArgs, values...)
	}
	return paramArgs, nil
}

func extractParamArgsFromMap(m map[string]any, names []string) ([]any, error) {
	paramArgs := make([]any, 0, len(names))
	for _, name := range names {
		value, ok := m[name]
		if !ok {
			continue
		}
		paramArgs = append(paramArgs, value)
	}
	return paramArgs, nil
}

func extractParamArgsFromStruct(obj any, names []string) ([]any, error) {
	paramArgs := make([]any, 0, len(names))
	for _, name := range names {
		nameParts := strings.Split(name, ".")
		currObj := obj
		for i, namePart := range nameParts {
			var dbTag reflectx.FieldTag
			field := reflectx.FindFieldFunc(currObj, func(field reflect.StructField) bool {
				dbTag = reflectx.Tag(field, "db")
				if dbTag.Value != "" {
					if dbTag.Value == name {
						return true
					}
				}
				return strings.EqualFold(field.Name, stringx.ToCamelCase(namePart))
			})
			if !field.IsValid() {
				return nil, errors.New("failed to resolve path '" + name + "' on struct " + reflect.TypeOf(obj).String())
			}
			if field.Kind() == reflect.Pointer && field.IsNil() {
				currObj = nil
				break
			}
			currObj = field.Interface()
			if i == len(nameParts)-1 && len(dbTag.Opts) > 0 {
				currObj = convertValue(currObj, dbTag.Opts[0])
			}
		}
		paramArgs = append(paramArgs, currObj)
	}
	return paramArgs, nil
}

func convertValue(value any, typeHint string) any {
	switch strings.ToLower(typeHint) {
	case "unixts":
		fallthrough
	case "unixtimestamp":
		fallthrough
	case "unix_timestamp":
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return value
			}
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct && v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.Unix()
		}
	}
	return value
}
