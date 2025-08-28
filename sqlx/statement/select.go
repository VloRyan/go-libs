package statement

type ColumnSourceType int

const (
	Field ColumnSourceType = iota
	Expression
)

type ColumnExpression struct {
	Name             string
	Alias            string
	SelectExpression string
	ValueExpression  string
	Source           ColumnSourceType
}

func (c *ColumnExpression) QualifiedName() string {
	return maskFieldNameIfNeeded(c.Name)
}

func (c *ColumnExpression) asSelectExp(table ObjectName, useAlias, useQualifiedNames bool) string {
	exp := ""
	switch c.Source {
	case Field:
		fieldName := maskFieldNameIfNeeded(c.Name)
		if useQualifiedNames {
			fieldName = table.FieldString(c.Name)
		}
		exp += fieldName
	case Expression:
		exp += c.SelectExpression
	}
	if useAlias && c.Alias != "" {
		exp += " AS " + maskFieldNameIfNeeded(c.Alias)
	}
	return exp
}

func NewSelect(selectExps ...ColumnExpression) *Statement {
	return &Statement{
		stmtType: SELECT,
		Fields:   selectExps,
	}
}

func (s *Statement) selectSql() string {
	return string(s.stmtType) +
		appendWithBlank(s.fieldListExp(true, true)) +
		appendWithBlank(s.fromExp()) +
		appendWithBlank(s.joinExp()) +
		appendWithBlank(s.whereExp()) +
		appendWithBlank(s.orderExp()) +
		appendWithBlank(s.limitExp()) +
		appendWithBlank(s.offsetExp())
}

func (s *Statement) fieldListExp(useAlias, useQualifiedNames bool) string {
	exp := ""
	tableObject := ObjectName{Name: s.tables[0]}
	for _, field := range s.Fields {
		if exp != "" {
			exp += ", "
		}
		exp += field.asSelectExp(tableObject, useQualifiedNames, useAlias)
	}
	for _, join := range s.tableJoins {
		for _, selectExp := range join.SelectFields {
			if exp != "" {
				exp += ", "
			}
			exp += selectExp.asSelectExp(join.Table, useQualifiedNames, useAlias)
		}
	}
	return exp
}
