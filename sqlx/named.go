package sqlx

import (
	"database/sql"
	"errors"
	"strconv"
	"unicode"
)

type NamedQuerier interface {
	Select(dest any, query string, args ...any) error
	Exec(query string, args ...any) (sql.Result, error)
}

var allowedBindRunes = []*unicode.RangeTable{unicode.Letter, unicode.Digit}

const (
	UNKNOWN  = ' '
	QUESTION = '?'
	DOLLAR   = '$'
	NAMED    = ':'
	AT       = '@'
)

// compileNamedQuery compiles a query with named parameters into a plain query with database specific parameter markers and collects
// a list of names.
func compileNamedQuery(qs []byte, paramMarker rune) (query string, names []string, err error) {
	names = make([]string, 0, 10)
	plainQuery := make([]byte, 0, len(qs))

	inName := false
	last := len(qs) - 1
	currentVar := 1
	name := make([]byte, 0, 10)

	for i, b := range qs {
		if b == ':' {
			if inName {
				err = errors.New("unexpected `:` while reading named param at " + strconv.Itoa(i))
				return query, names, err
			}
			inName = true
			name = []byte{}
		} else if inName {
			isAllowedRune := unicode.IsOneOf(allowedBindRunes, rune(b)) || b == '_' || b == '.' || b == '[' || b == ']'
			if isAllowedRune {
				name = append(name, b)
			} else {
				inName = false
			}
			if i == last {
				inName = false
			}
			if !inName {
				names = append(names, string(name))
				switch paramMarker {
				// oracle only supports named type bind vars even for positional
				case NAMED:
					plainQuery = append(plainQuery, ':')
					plainQuery = append(plainQuery, name...)
				case QUESTION, UNKNOWN:
					plainQuery = append(plainQuery, '?')
				case DOLLAR:
					plainQuery = append(plainQuery, '$')
					for _, b := range strconv.Itoa(currentVar) {
						plainQuery = append(plainQuery, byte(b))
					}
					currentVar++
				case AT:
					plainQuery = append(plainQuery, '@', 'p')
					for _, b := range strconv.Itoa(currentVar) {
						plainQuery = append(plainQuery, byte(b))
					}
					currentVar++
				}
			}
			if !isAllowedRune {
				plainQuery = append(plainQuery, b)
			}
		} else {
			plainQuery = append(plainQuery, b)
		}
	}

	return string(plainQuery), names, err
}
