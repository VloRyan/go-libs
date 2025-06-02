package router

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type DummyRouter struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

func (d *DummyRouter) handle(method string, path string, handler http.HandlerFunc) {
	d.Method = method
	d.Path = path
	d.Handler = handler
}

func TestRouting(t *testing.T) {
	type testCase struct {
		name string
		path []string
		want string
	}
	tests := []testCase{{
		name: "one level",
		path: []string{"test"},
		want: "/v1/test",
	}, {
		name: "multiple level",
		path: []string{"test", "this", "out"},
		want: "/v1/test/this/out",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noopFunc := func(http.ResponseWriter, *http.Request) {
			}
			router := new(DummyRouter)

			route := NewRoute("/v1", router.handle)
			for _, elem := range tt.path {
				route = route.SubRoute(elem)
			}
			if diff := cmp.Diff(tt.want, route.(*Route).path); diff != "" {
				t.Errorf("SubRoute() mismatch (-want +got):\n%s", diff)
			}

			route.GET("GET", noopFunc)
			if diff := cmp.Diff(tt.want+"/GET", router.Path); diff != "" {
				t.Errorf("GET() mismatch (-want +got):\n%s", diff)
			}

			route.POST("POST", noopFunc)
			if diff := cmp.Diff(tt.want+"/POST", router.Path); diff != "" {
				t.Errorf("POST() mismatch (-want +got):\n%s", diff)
			}

			route.PATCH("PATCH", noopFunc)
			if diff := cmp.Diff(tt.want+"/PATCH", router.Path); diff != "" {
				t.Errorf("PATCH() mismatch (-want +got):\n%s", diff)
			}

			route.DELETE("DELETE", noopFunc)
			if diff := cmp.Diff(tt.want+"/DELETE", router.Path); diff != "" {
				t.Errorf("DELETE() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
