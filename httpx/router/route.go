package router

import (
	"net/http"
	"net/url"
	"strings"
)

type HandleRouteFunc func(method, path string, handler http.HandlerFunc)

type RouteElement interface {
	SubRoute(path string) RouteElement
	Path() string
	GET(path string, handler http.HandlerFunc)
	POST(path string, handler http.HandlerFunc)
	DELETE(path string, handler http.HandlerFunc)
	PATCH(path string, handler http.HandlerFunc)
}

func NewRoute(path string, handler HandleRouteFunc) RouteElement {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	root := &RootRoute{
		handleRouteFunc: handler,
	}
	return root.SubRoute(path)
}

type RootRoute struct {
	handleRouteFunc HandleRouteFunc
}

func (r *RootRoute) SubRoute(path string) RouteElement {
	return &Route{
		root: r,
		path: path,
	}
}

func (r *RootRoute) Path() string {
	return "/"
}

type Route struct {
	root *RootRoute
	path string
}

func (e *Route) SubRoute(path string) RouteElement {
	return &Route{
		root: e.root,
		path: joinPath(e.path, path),
	}
}

func (e *Route) Path() string {
	return e.path
}

func (e *Route) GET(path string, handler http.HandlerFunc) {
	e.root.handleRouteFunc(http.MethodGet, joinPath(e.path, path), handler)
}

func (e *Route) POST(path string, handler http.HandlerFunc) {
	e.root.handleRouteFunc(http.MethodPost, joinPath(e.path, path), handler)
}

func (e *Route) DELETE(path string, handler http.HandlerFunc) {
	e.root.handleRouteFunc(http.MethodDelete, joinPath(e.path, path), handler)
}

func (e *Route) PATCH(path string, handler http.HandlerFunc) {
	e.root.handleRouteFunc(http.MethodPatch, joinPath(e.path, path), handler)
}

func joinPath(base string, elem ...string) string {
	path, err := url.JoinPath(base, elem...)
	if err != nil {
		return ""
	}
	return path
}
