package statement

import (
	"strings"

	"github.com/vloryan/go-libs/sqlx/pagination"
	"github.com/vloryan/go-libs/stringx"
)

type OrderBy struct {
	FieldName string
	Direction OrderDirection
	Coalesce  string
}

func (o *OrderBy) SQL() string {
	s := o.FieldName
	if o.Coalesce != "" {
		s = "COALESCE(" + o.FieldName + ", '" + o.Coalesce + "')"
	}
	if o.Direction == OrderDescending {
		s += " " + string(OrderDescending)
	}
	return s
}

func ToOrderBy(p *pagination.Page, tableName string) []*OrderBy {
	var orders []*OrderBy
	if p == nil {
		return orders
	}
	for _, sort := range p.Sort {
		name := sort
		var dir OrderDirection
		var coalesce string
		if sort[0] == '-' {
			name = sort[1:]
			dir = OrderDescending
		}
		sep := strings.Index(name, ":")
		if sep != -1 {
			coalesce = name[sep+1:]
			name = name[:sep]
		}
		orders = append(orders, &OrderBy{FieldName: FieldName(name, tableName), Direction: dir, Coalesce: coalesce})
	}
	return orders
}

func FieldName(path, tableName string) string {
	parts := strings.Split(path, ".")
	switch len(parts) {
	case 1:
		return stringx.ToSnakeCase(tableName) + "." + stringx.ToSnakeCase(parts[0])
	case 2:
		if parts[0] == "Base" {
			return stringx.ToSnakeCase(tableName) + "_entity." + stringx.ToSnakeCase(parts[1])
		}
		return stringx.ToSnakeCase(parts[0]) + "." + stringx.ToSnakeCase(parts[1])
	case 3:
		if parts[1] == "Base" {
			return stringx.ToSnakeCase(parts[0]) + "_entity." + stringx.ToSnakeCase(parts[2])
		}
	}
	return ""
}
