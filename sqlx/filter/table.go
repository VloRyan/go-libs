package filter

import (
	"reflect"
	"strconv"
)

type ColumnFunctionType int

var (
	lowerFunc ColumnFunctionFunc = func(columnExpr string, args ...any) string {
		return "LOWER(" + columnExpr + ")"
	}
	upperFunc ColumnFunctionFunc = func(columnExpr string, args ...any) string {
		return "UPPER(" + columnExpr + ")"
	}
	jsonbExtractFunc = func(columnExpr string, args ...any) string {
		return "jsonb_extract(" + columnExpr + ", '" + args[0].(string) + "')"
	}
	dateFunc = func(columnExpr string, args ...any) string {
		return "DATE(" + columnExpr + ")"
	}
)

type (
	ColumnFunctionFunc func(columnExpr string, args ...any) string
	ValueDecorator     func(valueExpr string) string
	TableFilter        struct {
		Name string
	}
	ColumnFilter struct {
		/*a.k.a ColumnName*/
		Name               string
		TableName          string
		ColumnFunction     ColumnFunctionFunc
		ColumnFunctionArgs []any
	}
)

var AsDate ValueDecorator = func(valueExpr string) string {
	return dateFunc(valueExpr)
}

func NewTable(name string) *TableFilter {
	return &TableFilter{
		Name: name,
	}
}

func (t *TableFilter) Column(name string) *ColumnFilter {
	return &ColumnFilter{
		Name:      name,
		TableName: t.Name,
	}
}

func (f *ColumnFilter) Exists() Criteria {
	return f.asCriteria(ExistsOp, nil, nil)
}

func (f *ColumnFilter) IsNil() Criteria {
	return f.Eq(nil)
}

func (f *ColumnFilter) IsNotNil() Criteria {
	return f.Eq(nil).Not()
}

func (f *ColumnFilter) IsTrue() Criteria {
	return f.Eq(true)
}

func (f *ColumnFilter) IsFalse() Criteria {
	return f.Eq(false)
}

func (f *ColumnFilter) Eq(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(EqOp, value, decorator)
}

func (f *ColumnFilter) Gt(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(GtOp, value, decorator)
}

func (f *ColumnFilter) GtEq(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(GtEqOp, value, decorator)
}

func (f *ColumnFilter) Lt(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(LtOp, value, decorator)
}

func (f *ColumnFilter) LtEq(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(LtEqOp, value, decorator)
}

func (f *ColumnFilter) Neq(value any, decorator ...ValueDecorator) Criteria {
	return f.Eq(value, decorator...).Not()
}

func (f *ColumnFilter) In(values []any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(InOp, values, decorator)
}

func (f *ColumnFilter) Like(pattern string, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(LikeOp, pattern, decorator)
}

// Between filters for values less than or equal to high and greater than or equal to the value of low.
// (field >= low AND field <= high)
func (f *ColumnFilter) Between(low any, high any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(BetweenOp, []any{low, high}, decorator)
}

func (f *ColumnFilter) ToLower() *ColumnFilter {
	f.ColumnFunction = lowerFunc
	return f
}

func (f *ColumnFilter) ToUpper() *ColumnFilter {
	f.ColumnFunction = upperFunc
	return f
}

func (f *ColumnFilter) JSONBExtract(path string) *ColumnFilter {
	f.ColumnFunction = jsonbExtractFunc
	f.ColumnFunctionArgs = []any{path}
	return f
}

func (f *ColumnFilter) AsDate() *ColumnFilter {
	f.ColumnFunction = dateFunc
	return f
}

func (f *ColumnFilter) WithColumnFunc(fun ColumnFunctionFunc, args ...any) *ColumnFilter {
	f.ColumnFunction = fun
	f.ColumnFunctionArgs = args
	return f
}

func (f *ColumnFilter) asCriteria(opType OpFuncType, value any, decorator []ValueDecorator) *UnaryCriteria {
	columnExpr := f.TableName + "." + f.Name
	if f.ColumnFunction != nil {
		columnExpr = f.ColumnFunction(columnExpr, f.ColumnFunctionArgs...)
	}
	paramName := f.TableName + "_" + f.Name
	valueExpr := ":" + paramName
	for _, decorate := range decorator {
		valueExpr = decorate(valueExpr)
	}
	var parameter map[string]any
	switch opType {

	case BetweenOp:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			parameter = make(map[string]any, v.Len())
			valueExpr = ":" + paramName + "_0 AND " + ":" + paramName + "_1"
			parameter[paramName+"_0"] = v.Index(0).Interface()
			parameter[paramName+"_1"] = v.Index(1).Interface()
		}
	case InOp:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			parameter = make(map[string]any, v.Len())
			valueExpr = "("
			for i := 0; i < v.Len(); i++ {
				elemName := paramName + "_" + strconv.Itoa(i)
				parameter[elemName] = v.Index(i).Interface()
				if i > 0 {
					valueExpr += ", "
				}
				valueExpr += ":" + elemName
			}
			valueExpr += ")"
		}
	default:
		parameter = map[string]any{paramName: value}
	}
	return &UnaryCriteria{
		OpType:     opType,
		ColumnExpr: columnExpr,
		ValueExpr:  valueExpr,
		Parameter:  parameter,
	}
}
