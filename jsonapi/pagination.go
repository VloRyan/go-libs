package jsonapi

import (
	"github.com/vloryan/go-libs/httpx"
	"github.com/vloryan/go-libs/sqlx/pagination"
	"net/http"
	"strings"
)

var DefaultPageLimit = 25

func ExtractPagination(req *http.Request) *pagination.Page {
	var sorts []string
	queryParamSort := httpx.Query(req, "sort")
	if queryParamSort != "" {
		sorts = strings.Split(queryParamSort, ",")
	}
	offset := httpx.QueryFamilyMemberInt(req, "page", "offset", 0)
	limit := httpx.QueryFamilyMemberInt(req, "page", "limit", DefaultPageLimit)
	return pagination.NewPage(offset, limit, sorts...)
}
