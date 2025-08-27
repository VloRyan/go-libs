package filter

type Where struct {
	Clause    string
	Parameter map[string]any
}

func (c *Where) SQL() string {
	if c.Clause == "" {
		return ""
	}
	return "WHERE " + c.Clause
}
