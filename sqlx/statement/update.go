package statement

func NewUpdate(table string) *Statement {
	return &Statement{
		stmtType: UPDATE,
		tables:   []string{table},
	}
}

func (s *Statement) updateSql() string {
	return string(s.stmtType) +
		appendWithBlank(s.tables[0]) +
		appendWithBlank("SET") +
		appendWithBlank(s.valuePlaceholders(true, "")) +
		appendWithBlank(s.whereExp())
}
