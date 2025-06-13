package jsonapi

import (
	"fmt"
	"strconv"
)

type Error struct {
	ID     string         `json:"id,omitempty"`
	Links  map[string]any `json:"links,omitempty"`
	Status string         `json:"status,omitempty"`
	Code   string         `json:"code,omitempty"`
	Title  string         `json:"title,omitempty"`
	Detail string         `json:"detail,omitempty"`
	Source string         `json:"source,omitempty"`
	Meta   MetaData       `json:"meta,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("jsonapi(status: %s): %s\n%s", e.Status, e.Title, e.Detail)
}

func NewError(status int, title string, detail error) *Error {
	e := &Error{
		Status: strconv.Itoa(status),
		Title:  title,
	}
	if detail != nil {
		e.Detail = detail.Error()
	}
	return e
}
