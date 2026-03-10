package filter

type Where struct {
	Statement string
	Parameter map[string]any
}

func (w *Where) Clause() string {
	if w.Statement == "" {
		return ""
	}
	return "WHERE " + w.Statement
}

func (w *Where) Empty() bool {
	return w.Statement == ""
}
