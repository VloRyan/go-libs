package statement

func NewDelete(fromTable string) *Statement {
	return &Statement{
		stmtType: DELETE,
		tables:   []string{fromTable},
	}
}

func (s *Statement) deleteSql() string {
	return string(s.stmtType) +
		appendWithBlank(s.fromExp()) +
		appendWithBlank(s.whereExp())
}
