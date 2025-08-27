package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
		if reflect.TypeOf(dest).Kind() != reflect.Ptr {
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
		currentField := reflect.ValueOf(dest)
		fieldPath := strings.Split(colName, ".")
		for i := 0; i < len(fieldPath); i++ {
			field := reflectx.FindFieldFunc(currentField.Interface(), func(field reflect.StructField) bool {
				dbTag := reflectx.Tag(field, "db")
				if dbTag.Value != "" {
					if dbTag.Value == fieldPath[i] {
						return true
					}
					if len(dbTag.Opts) > 0 && dbTag.Opts[0] == fieldPath[i] {
						return true
					}
				}
				return strings.EqualFold(field.Name, stringx.ToCamelCase(fieldPath[i]))
			})

			if !field.IsValid() {
				continue
			}
			if i == len(fieldPath)-1 {
				pField := field
				if pField.Kind() != reflect.Ptr {
					pField = field.Addr()
				}
				if scanner, ok := pField.Interface().(sql.Scanner); ok {
					if err := scanner.Scan(*val); err != nil {
						return err
					}
				} else {
					if err := reflectx.SetFieldValue(field, *val); err != nil {
						return err
					}
				}
			} else {
				if field.Type().Kind() == reflect.Ptr && field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
			}
			currentField = field

		}
	}
	return nil
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
		for _, namePart := range nameParts {
			field := reflectx.FindFieldFunc(currObj, func(field reflect.StructField) bool {
				dbTag := reflectx.Tag(field, "db")
				if dbTag.Value != "" {
					if dbTag.Value == name {
						return true
					}
					if len(dbTag.Opts) > 0 && dbTag.Opts[0] == namePart {
						return true
					}
				}
				return strings.EqualFold(field.Name, stringx.ToCamelCase(namePart))
			})
			if !field.IsValid() {
				continue
			}
			if field.Kind() == reflect.Ptr && field.IsNil() {
				currObj = nil
				break
			}
			currObj = field.Interface()
		}
		paramArgs = append(paramArgs, currObj)
	}
	return paramArgs, nil
}
