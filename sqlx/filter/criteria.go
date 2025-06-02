package filter

type Criteria interface {
	Not() Criteria
	And(c Criteria) Criteria
	Or(c Criteria) Criteria
}

type EmptyCriteria struct{}

func (c *EmptyCriteria) Not() Criteria {
	return not(c)
}

func (c *EmptyCriteria) And(c2 Criteria) Criteria {
	return c2
}

func (c *EmptyCriteria) Or(c2 Criteria) Criteria {
	return c2
}

func New() Criteria {
	return &EmptyCriteria{}
}

const (
	OpTypeAnd = iota
	OpTypeOr
)

const (
	ExistsOp = iota
	EqOp
	GtOp
	GtEqOp
	LtOp
	LtEqOp
	LikeOp
	InOp
	ContainsOp
	BetweenOp
	FunctionOp
)

type BinaryCriteria struct {
	C1     Criteria
	C2     Criteria
	OpType int
}

func (b *BinaryCriteria) Not() Criteria {
	return not(b)
}

func (b *BinaryCriteria) And(c Criteria) Criteria {
	return and(b, c)
}

func (b *BinaryCriteria) Or(c Criteria) Criteria {
	return or(b, c)
}

type UnaryCriteria struct {
	OpType       int
	FieldPath    string
	FieldModType int
	Function     string
	Args         []string
	Value        any
}
type NotCriteria struct {
	C Criteria
}

func (n *NotCriteria) Not() Criteria {
	return not(n)
}

func (n *NotCriteria) And(c Criteria) Criteria {
	return and(n, c)
}

func (n *NotCriteria) Or(c Criteria) Criteria {
	return or(n, c)
}

func (f *UnaryCriteria) Not() Criteria {
	return not(f)
}

func (f *UnaryCriteria) And(c Criteria) Criteria {
	return and(f, c)
}

func (f *UnaryCriteria) Or(c Criteria) Criteria {
	return or(f, c)
}

func newCriteria(opType int, path, function string, functionArgs []string, value any) Criteria {
	return &UnaryCriteria{
		OpType:    opType,
		FieldPath: path,
		Function:  function,
		Args:      functionArgs,
		Value:     value,
	}
}

func and(c1, c2 Criteria) Criteria {
	return &BinaryCriteria{
		OpType: OpTypeAnd,
		C1:     c1,
		C2:     c2,
	}
}

func or(c1, c2 Criteria) Criteria {
	return &BinaryCriteria{
		OpType: OpTypeOr,
		C1:     c1,
		C2:     c2,
	}
}

func not(c Criteria) Criteria {
	return &NotCriteria{c}
}
