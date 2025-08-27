package filter

import "strings"

type Criteria interface {
	Not() Criteria
	And(c Criteria) Criteria
	Or(c Criteria) Criteria
	ToWhere() Where
}

type EmptyCriteria struct{}

func (c *EmptyCriteria) Not() Criteria {
	return not(c)
}

func (c *EmptyCriteria) And(Second Criteria) Criteria {
	return Second
}

func (c *EmptyCriteria) Or(Second Criteria) Criteria {
	return Second
}

func (c *EmptyCriteria) ToWhere() Where {
	return Where{}
}

func New() Criteria {
	return &EmptyCriteria{}
}

type ConnOp int

const (
	ConnOpAnd ConnOp = iota
	ConnOpOr
)

type OpFuncType int

const (
	ExistsOp OpFuncType = iota
	EqOp
	GtOp
	GtEqOp
	LtOp
	LtEqOp
	LikeOp
	InOp
	BetweenOp
)

type BinaryCriteria struct {
	First  Criteria
	Second Criteria
	Conn   ConnOp
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

func (b *BinaryCriteria) ToWhere() Where {
	where := Where{
		Parameter: make(map[string]any),
	}
	w1 := b.First.ToWhere()
	where.Clause = "(" + w1.Clause + ")"
	if b.Conn == ConnOpAnd {
		where.Clause += " AND "
	} else {
		where.Clause += " OR "
	}
	w2 := b.Second.ToWhere()
	where.Clause += "(" + w2.Clause + ")"
	for k, v := range w1.Parameter {
		where.Parameter[k] = v
	}
	for k, v := range w2.Parameter {
		where.Parameter[k] = v
	}
	if len(where.Parameter) == 0 {
		where.Parameter = nil
	}
	return where
}

type (
	ValueFunctionType func() (string, map[string]any)
	ValueFunction     int
	UnaryCriteria     struct {
		OpType        OpFuncType
		ColumnExpr    string
		ValueExpr     string
		Parameter     map[string]any
		ValueFunction ValueFunctionType
	}
)

func (f *UnaryCriteria) Not() Criteria {
	return not(f)
}

func (f *UnaryCriteria) And(c Criteria) Criteria {
	return and(f, c)
}

func (f *UnaryCriteria) Or(c Criteria) Criteria {
	return or(f, c)
}

func (f *UnaryCriteria) ToWhere() Where {
	var op string
	valueExpr := f.ValueExpr
	switch f.OpType {
	case ExistsOp:
		op = "IS NOT NULL"
		valueExpr = ""
	case EqOp:
		op = "= "
	case LikeOp:
		op = "LIKE "
	case InOp:
		op = "IN ("
		params := make([]string, len(f.Parameter))
		i := 0
		for k := range f.Parameter {
			params[i] = ":" + k
			i++
		}
		valueExpr = strings.Join(params, ", ") + ")"
	case GtOp:
		op = "> "
	case GtEqOp:
		op = ">= "
	case LtOp:
		op = "< "
	case LtEqOp:
		op = "<= "
	case BetweenOp:
		op = "BETWEEN "
		valueExpr = valueExpr + "_0 AND " + valueExpr + "_1"
	default:
	}
	where := Where{
		Clause:    f.ColumnExpr + " " + op + valueExpr,
		Parameter: f.Parameter,
	}
	return where
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

func (n *NotCriteria) ToWhere() Where {
	where := Where{}
	w := n.C.ToWhere()
	where.Clause = "NOT (" + w.Clause + ")"
	if w.Parameter != nil {
		where.Parameter = w.Parameter
	}
	return where
}

func and(c1, Second Criteria) Criteria {
	return &BinaryCriteria{
		Conn:   ConnOpAnd,
		First:  c1,
		Second: Second,
	}
}

func or(c1, Second Criteria) Criteria {
	return &BinaryCriteria{
		Conn:   ConnOpOr,
		First:  c1,
		Second: Second,
	}
}

func not(c Criteria) Criteria {
	return &NotCriteria{c}
}
