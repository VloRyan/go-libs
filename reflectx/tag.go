package reflectx

import "strings"

type FieldTag struct {
	Value string
	Opts  []string
}

func (t FieldTag) Is(value string) bool {
	return t.Value == value
}

// Has returns true if the given option is available in tagOptions
func (t FieldTag) Has(opt string) bool {
	for _, tagOpt := range t.Opts {
		if tagOpt == opt {
			return true
		}
	}
	return false
}

// ParseTag splits a struct field's tag into its name and a list of options
// which comes after a name. A tag is in the form of: "name,option1,option2".
func ParseTag(tag string) FieldTag {
	p := strings.Split(tag, ",")
	return FieldTag{
		Value: p[0],
		Opts:  p[1:],
	}
}
