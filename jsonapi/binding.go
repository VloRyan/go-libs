package jsonapi

import (
	"errors"
	"github.com/vloryan/go-libs/httpx/router"
	"io"
	"net/http"
)

var APIBinding = jsonAPIBinding{}

type jsonAPIBinding struct{}

func (j jsonAPIBinding) Name() string {
	return "json:api"
}

func (j jsonAPIBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	d := NewDocument()
	if err := j.BindDocument(req, d); err != nil {
		return err
	}
	return d.MapData(obj)
}

func (j jsonAPIBinding) BindDocument(req *http.Request, doc *Document) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return doc.UnmarshalJSON(b)
}

type ResourceHandler interface {
	RegisterRoutes(route router.RouteElement)
}
