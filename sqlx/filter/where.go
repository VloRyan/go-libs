package filter

type Where struct {
	Statement string
	Parameter map[string]any
}

func (c *Where) Clause() string {
	if c.Statement == "" {
		return ""
	}
	return "WHERE " + c.Statement
}

func (c *Where) Empty() bool {
	return c.Statement == ""
}
