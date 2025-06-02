package httpx

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/vloryan/go-libs/httpx/binding"
)

func Query(req *http.Request, name string) string {
	return req.URL.Query().Get(name)
}

// QueryMap returns a map for a given query key.
func QueryMap(req *http.Request, key string) map[string]string {
	qm, _ := QueryFamily(req, key)
	return qm
}

func QueryFamily(req *http.Request, familyName string) (map[string]string, bool) {
	return queryFamily(req.URL.Query(), familyName)
}

func QueryFamilyMember(req *http.Request, familyName, memberName string) (string, bool) {
	if qm, ok := QueryFamily(req, familyName); ok {
		return qm[memberName], true
	}
	return "", false
}

func QueryFamilyMemberInt(req *http.Request, objectName, memberName string, def int) int {
	value, exits := QueryFamilyMember(req, objectName, memberName)
	if !exits {
		return def
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return def
	}
	return int(i)
}

func QueryFamilyMemberUint(req *http.Request, objectName, memberName string, def uint) uint {
	value, exits := QueryFamilyMember(req, objectName, memberName)
	if !exits {
		return def
	}
	i, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return def
	}
	return uint(i)
}

// queryFamily is an internal method and returns a map which satisfy conditions.
func queryFamily(m map[string][]string, key string) (map[string]string, bool) {
	dict := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dict[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dict, exist
}

// BindQuery is a shortcut for c.MustBindWith(obj, binding.Query).
func BindQuery(req *http.Request, obj any) error {
	return ShouldBindWith(req, obj, binding.Query)
}

// MustBindWith binds the passed struct pointer using the specified binding engine.
// It will abort the request with HTTP 400 if any error occurs.
// See the binding package.
func MustBindWith(req *http.Request, obj any, b binding.Binding) {
	if err := ShouldBindWith(req, obj, b); err != nil {
		panic(err)
	}
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func ShouldBindWith(req *http.Request, obj any, b binding.Binding) (err error) {
	return b.Bind(req, obj)
}
