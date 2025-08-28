package statement

import "strconv"

func NewInsert(intoTable string) *Statement {
	return &Statement{
		stmtType:  INSERT,
		tables:    []string{intoTable},
		valuesLen: 1,
	}
}

func (s *Statement) Returning(fields ...ColumnExpression) *Statement {
	s.returningFieldDefs = fields
	return s
}

func (s *Statement) WithValuesLen(len uint) *Statement {
	s.valuesLen = len
	return s
}

func (s *Statement) insertSql() string {
	sql := string(s.stmtType) +
		appendWithBlank(s.tables[0]) + "(" + s.columnNames() + ")" +
		appendWithBlank("VALUES")
	if s.valuesLen > 1 {
		for i := uint(0); i < s.valuesLen; i++ {
			if i > 0 {
				sql += ", "
			}
			sql += "(" + s.valuePlaceholders(false, "["+strconv.Itoa(int(i))+"]") + ")"
		}
	} else {
		sql += "(" + s.valuePlaceholders(false, "") + ")"
	}

	if len(s.returningFieldDefs) > 0 {
		sql += "\nRETURNING "
		returningFields := ""
		for _, fieldDef := range s.returningFieldDefs {
			if returningFields != "" {
				returningFields += ", "
			}
			if fieldDef.Alias != "" {
				returningFields += fieldDef.Alias
			} else {
				returningFields += maskFieldNameIfNeeded(fieldDef.Name)
			}
		}
		sql += returningFields
	}
	return sql
}

func (s *Statement) columnNames() string {
	exp := ""
	for _, field := range s.Fields {
		if exp != "" {
			exp += ", "
		}
		exp += field.QualifiedName()
	}
	for _, join := range s.tableJoins {
		for _, selectExp := range join.SelectFields {
			if exp != "" {
				exp += ", "
			}
			exp += selectExp.QualifiedName()
		}
	}
	return exp
}

func (s *Statement) valuePlaceholders(asUpdate bool, fieldSuffix string) string {
	placeholders := ""
	for _, fieldDef := range s.Fields {
		if placeholders != "" {
			placeholders += ", "
		}
		if asUpdate {
			placeholders += fieldDef.Name + " = "
		}
		switch fieldDef.Source {
		case Field:
			if fieldDef.Alias != "" {
				placeholders += ":" + fieldDef.Alias
			} else {
				placeholders += ":" + fieldDef.Name
			}
			placeholders += fieldSuffix
		case Expression:
			placeholders += fieldDef.ValueExpression
		}
	}
	return placeholders
}
