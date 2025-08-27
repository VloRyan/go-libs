package filter

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
	ColumnFilter       struct {
		/*a.k.a ColumnName*/
		Name               string
		ColumnFunction     ColumnFunctionFunc
		ColumnFunctionArgs []any
	}
)

var AsDate ValueDecorator = func(valueExpr string) string {
	return dateFunc(valueExpr)
}

type ObjectFilter struct {
	ObjectPath string
}
type JSONFilter struct {
	FieldPath   string
	FieldFilter *ColumnFilter
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
	return f.asCriteria(EqOp, map[string]any{f.Name: value}, decorator)
}

func (f *ColumnFilter) Gt(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(GtOp, map[string]any{f.Name: value}, decorator)
}

func (f *ColumnFilter) GtEq(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(GtEqOp, map[string]any{f.Name: value}, decorator)
}

func (f *ColumnFilter) Lt(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(LtOp, map[string]any{f.Name: value}, decorator)
}

func (f *ColumnFilter) LtEq(value any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(LtEqOp, map[string]any{f.Name: value}, decorator)
}

func (f *ColumnFilter) Neq(value any, decorator ...ValueDecorator) Criteria {
	return f.Eq(value, decorator...).Not()
}

func (f *ColumnFilter) In(values []any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(InOp, map[string]any{f.Name: values}, decorator)
}

func (f *ColumnFilter) Like(pattern string, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(LikeOp, map[string]any{f.Name: pattern}, decorator)
}

// Between filters for values less than or equal to high and greater than or equal to the value of low.
// (field >= low AND field <= high)
func (f *ColumnFilter) Between(low any, high any, decorator ...ValueDecorator) Criteria {
	return f.asCriteria(BetweenOp, map[string]any{f.Name: []any{low, high}}, decorator)
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

func (f *ColumnFilter) asCriteria(opType OpFuncType, params map[string]any, decorator []ValueDecorator) *UnaryCriteria {
	columnExpr := f.Name
	if f.ColumnFunction != nil {
		columnExpr = f.ColumnFunction(columnExpr, f.ColumnFunctionArgs...)
	}
	valueExpr := ":" + f.Name
	for _, decorate := range decorator {
		valueExpr = decorate(valueExpr)
	}
	return &UnaryCriteria{
		OpType:     opType,
		ColumnExpr: columnExpr,
		ValueExpr:  valueExpr,
		Parameter:  params,
	}
}

func (f *ObjectFilter) Field(name string) *ColumnFilter {
	if f.ObjectPath == "" {
		return &ColumnFilter{
			FieldPath: name,
		}
	}
	return &ColumnFilter{
		FieldPath: f.ObjectPath + "." + name,
	}
}

func (f *ObjectFilter) Object(path string) *ObjectFilter {
	if f.ObjectPath == "" {
		return &ObjectFilter{
			ObjectPath: path,
		}
	}
	return &ObjectFilter{
		ObjectPath: f.ObjectPath + "." + path,
	}
}

func Object(path string) *ObjectFilter {
	return &ObjectFilter{
		ObjectPath: path,
	}
}

func Field(path string) *ColumnFilter {
	return &ColumnFilter{
		Name: path,
	}
}
