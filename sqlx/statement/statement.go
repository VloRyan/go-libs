package statement

import (
	"strconv"
	"strings"
)

type ObjectName struct {
	Name  string
	Alias string
}

func (o *ObjectName) String() string {
	if o.Alias != "" {
		return o.Name + " AS " + o.Alias
	}
	return o.Name
}

func (o *ObjectName) FieldString(fieldName string) string {
	if o.Alias != "" {
		return o.Alias + "." + maskFieldNameIfNeeded(fieldName)
	}
	return o.Name + "." + maskFieldNameIfNeeded(fieldName)
}

func (o *ObjectName) Identifier() string {
	if o.Alias != "" {
		return o.Alias
	}
	return o.Name
}

type Statement struct {
	stmtType           Type
	Fields             []ColumnExpression
	returningFieldDefs []ColumnExpression
	tables             []string
	tableJoins         []TableJoinDefinition
	where              string
	limit              int
	offset             int
	order              []*OrderBy
	valuesLen          uint
}

type (
	JoinType   string
	Type       string
	AppendType string
)

const (
	InnerJoin JoinType = "JOIN"
	LeftJoin  JoinType = "LEFT OUTER JOIN"
)

type OrderDirection string

const (
	// OrderDefault leaved ordering to DBMS, SQLites default is ascending
	OrderDefault    OrderDirection = ""
	OrderDescending OrderDirection = "DESC"
)

const (
	SELECT Type = "SELECT"
	INSERT Type = "INSERT INTO"
	UPDATE Type = "UPDATE"
	DELETE Type = "DELETE"
)

type TableJoinDefinition struct {
	Table        ObjectName
	OnConditions []string
	Type         JoinType
	SelectFields []ColumnExpression
}

func (s *Statement) Copy() *Statement {
	return &Statement{
		stmtType:   s.stmtType,
		Fields:     s.Fields,
		tables:     s.tables,
		tableJoins: s.tableJoins,
		where:      s.where,
		limit:      s.limit,
		offset:     s.offset,
		order:      s.order,
	}
}

func (s *Statement) As(stmtType Type) *Statement {
	s.stmtType = stmtType
	return s
}

func (s *Statement) WithFields(fields ...ColumnExpression) *Statement {
	s.Fields = fields
	return s
}

func (s *Statement) From(names ...string) *Statement {
	s.tables = append(s.tables, names...)
	return s
}

func (s *Statement) Joins(joinDefs ...TableJoinDefinition) *Statement {
	s.tableJoins = append(s.tableJoins, joinDefs...)
	return s
}

func (s *Statement) Join(table string, onConditions ...string) *Statement {
	s.tableJoins = append(s.tableJoins, TableJoinDefinition{
		Type: InnerJoin,
		Table: ObjectName{
			Name: table,
		},
		OnConditions: onConditions,
	})
	return s
}

func (s *Statement) JoinAs(table string, alias string, onConditions ...string) *Statement {
	s.tableJoins = append(s.tableJoins, TableJoinDefinition{
		Type: InnerJoin,
		Table: ObjectName{
			Name:  table,
			Alias: alias,
		},
		OnConditions: onConditions,
	})
	return s
}

func (s *Statement) LeftJoin(table string, onConditions ...string) *Statement {
	s.tableJoins = append(s.tableJoins, TableJoinDefinition{
		Type: LeftJoin,
		Table: ObjectName{
			Name: table,
		},
		OnConditions: onConditions,
	})
	return s
}

func (s *Statement) LeftJoinAs(table string, alias string, onConditions ...string) *Statement {
	s.tableJoins = append(s.tableJoins, TableJoinDefinition{
		Type: LeftJoin,
		Table: ObjectName{
			Name:  table,
			Alias: alias,
		},
		OnConditions: onConditions,
	})
	return s
}

func (s *Statement) Where(where string) *Statement {
	s.where = where
	return s
}

func (s *Statement) Limit(limit int) *Statement {
	s.limit = limit
	return s
}

func (s *Statement) Order(order []*OrderBy) *Statement {
	s.order = order
	return s
}

func (s *Statement) Offset(offset int) *Statement {
	s.offset = offset
	return s
}

func (s *Statement) SQL() string {
	var sql string
	switch s.stmtType {
	case SELECT:
		sql = s.selectSql()
	case INSERT:
		sql = s.insertSql()
	case UPDATE:
		sql = s.updateSql()
	case DELETE:
		sql = s.deleteSql()
	}
	return sql
}

func appendWithBlank(text string) string {
	if text == "" {
		return ""
	}
	return " " + text
}

func (s *Statement) fromExp() string {
	return "FROM " + strings.Join(s.tables, ", ")
}

func (s *Statement) joinExp() string {
	joins := ""
	for _, join := range s.tableJoins {
		if joins != "" {
			joins += " "
		}
		joins += join.String()
	}
	return joins
}

func (s *Statement) whereExp() string {
	if s.where == "" {
		return ""
	}
	return "WHERE " + s.where
}

func (s *Statement) limitExp() string {
	if s.limit > 0 {
		return "LIMIT " + strconv.Itoa(int(s.limit))
	}
	return ""
}

func (s *Statement) offsetExp() string {
	if s.offset > 0 {
		return "OFFSET " + strconv.Itoa(int(s.offset))
	}
	return ""
}

func (s *Statement) orderExp() string {
	if len(s.order) == 0 {
		return ""
	}
	orderBy := ""
	for _, o := range s.order {
		if orderBy != "" {
			orderBy += ", "
		}
		orderBy += o.SQL()
	}
	return "ORDER BY " + orderBy
}

func (j *TableJoinDefinition) String() string {
	if j.Type == "" {
		j.Type = InnerJoin
	}
	exp := string(j.Type) + " " + j.Table.String()
	onExp := ""
	for _, cond := range j.OnConditions {
		if onExp != "" {
			onExp = onExp + " AND "
		}
		onExp = onExp + cond
	}
	if onExp != "" {
		exp = exp + " ON " + onExp
	}
	return exp
}

func BuildPlaceholder(count int) string {
	return strings.Join(strings.Split(strings.Repeat("?", count), ""), ", ")
}

func BuildPlaceholderMap[T any](s []T) map[string]any {
	placeholder := make(map[string]any)
	for i, e := range s {
		placeholder["p"+strconv.Itoa(i)] = e
	}
	return placeholder
}

func maskFieldNameIfNeeded(name string) string {
	if name == "index" {
		return "`" + name + "`"
	}
	if name == "case" {
		return "`" + name + "`"
	}
	if strings.Contains(name, ".") {
		return "`" + name + "`"
	}
	return name
}

func FieldColumnsFromNames(names ...string) []ColumnExpression {
	var fieldDefs []ColumnExpression
	for _, name := range names {
		fieldDefs = append(fieldDefs, ColumnExpression{
			Name:   name,
			Source: Field,
		})
	}
	return fieldDefs
}

func ToParamSlice[ID comparable](ids []ID) (r []any) {
	for _, e := range ids {
		r = append(r, e)
	}
	return
}
