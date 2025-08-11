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
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	d := NewDocument()
	if err := b.BindDocument(req, d); err != nil {
		return err
	}
	return d.MapData(obj)
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
