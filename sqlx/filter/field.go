package filter

const (
	ToLowerFunc     = "to_lower"
	JSONExtractFunc = "json_extract"
)

type FieldFilter struct {
	FieldPath    string
	Function     string
	FunctionArgs []string
}

type ObjectFilter struct {
	ObjectPath string
}
type JSONFilter struct {
	FieldPath   string
	FieldFilter *FieldFilter
}

func (f *FieldFilter) Exists() Criteria {
	return newCriteria(ExistsOp, f.FieldPath, f.Function, f.FunctionArgs, nil)
}

func (f *FieldFilter) NotExists() Criteria {
	return newCriteria(ExistsOp, f.FieldPath, f.Function, f.FunctionArgs, nil).Not()
}

func (f *FieldFilter) IsNil() Criteria {
	return f.Eq(nil)
}

func (f *FieldFilter) IsNotNil() Criteria {
	return f.Eq(nil).Not()
}

func (f *FieldFilter) IsTrue() Criteria {
	return f.Eq(true)
}

func (f *FieldFilter) IsFalse() Criteria {
	return f.Eq(false)
}

func (f *FieldFilter) IsNilOrNotExists() Criteria {
	return f.IsNil().Or(f.NotExists())
}

func (f *FieldFilter) Eq(value any) Criteria {
	return newCriteria(EqOp, f.FieldPath, f.Function, f.FunctionArgs, value)
}

func (f *FieldFilter) Gt(value any) Criteria {
	return newCriteria(GtOp, f.FieldPath, f.Function, f.FunctionArgs, value)
}

func (f *FieldFilter) GtEq(value any) Criteria {
	return newCriteria(GtEqOp, f.FieldPath, f.Function, f.FunctionArgs, value)
}

func (f *FieldFilter) Lt(value any) Criteria {
	return newCriteria(LtOp, f.FieldPath, f.Function, f.FunctionArgs, value)
}

func (f *FieldFilter) LtEq(value any) Criteria {
	return newCriteria(LtEqOp, f.FieldPath, f.Function, f.FunctionArgs, value)
}

func (f *FieldFilter) Neq(value any) Criteria {
	return f.Eq(value).Not()
}

func (f *FieldFilter) In(values ...any) Criteria {
	return newCriteria(InOp, f.FieldPath, f.Function, f.FunctionArgs, values)
}

func (f *FieldFilter) Like(pattern string) Criteria {
	return newCriteria(LikeOp, f.FieldPath, f.Function, f.FunctionArgs, pattern)
}

func (f *FieldFilter) Contains(elems ...any) Criteria {
	return newCriteria(ContainsOp, f.FieldPath, f.Function, f.FunctionArgs, elems)
}

// Between filters for values less than or equal to high and greater than or equal to the value of low.
// (field >= low AND field <= high)
func (f *FieldFilter) Between(low any, high any) Criteria {
	return newCriteria(BetweenOp, f.FieldPath, f.Function, f.FunctionArgs, []any{low, high})
}

func (f *FieldFilter) ToLowerCase() *FieldFilter {
	f.Function = ToLowerFunc
	return f
}

func (f *FieldFilter) JSONExtract(path string) *FieldFilter {
	f.Function = JSONExtractFunc
	f.FunctionArgs = []string{path}
	return f
}

func (f *ObjectFilter) Field(name string) *FieldFilter {
	if f.ObjectPath == "" {
		return &FieldFilter{
			FieldPath: name,
		}
	}
	return &FieldFilter{
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

func Field(path string) *FieldFilter {
	return &FieldFilter{
		FieldPath: path,
	}
}
