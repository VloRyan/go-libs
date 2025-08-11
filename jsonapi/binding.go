package jsonapi

import (
	"errors"
	"io"
	"net/http"
)

var Binding = binding{}

type binding struct{}

func (b binding) Name() string {
	return "json:api"
}

func (b binding) Bind(req *http.Request, obj any) error {
	return b.BindFunc(req, func(doc *Document) error {
		return doc.MapData(obj)
	})
}
func (b binding) BindFunc(req *http.Request, bindFunc func(doc *Document) error) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	d := NewDocument()
	if err := b.BindDocument(req, d); err != nil {
		return err
	}
	return bindFunc(d)
}

func (b binding) BindDocument(req *http.Request, doc *Document) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return doc.UnmarshalJSON(bytes)
}
