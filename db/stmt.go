package db

import (
    "fmt"
    "strings"
    "regexp"
)

type Scanner interface{
    Scan(args ...interface{}) error
}

var stmtWhereSpace = regexp.MustCompile(`\s+`)

type Stmt struct {
    cols, from, join, where, group, order string
    offset, limit int64
    placeholders []string
    values []interface{}
}

func (s *Stmt) Select(cols string) *Stmt {
    s.cols = cols

    return s
}

func (s *Stmt) From(table, alias string) *Stmt {
    s.from = fmt.Sprintf("`%s` %s", table, alias)

    return s
}

func (s *Stmt) LeftJoin(table, alias, on string) *Stmt {
    s.join = fmt.Sprintf("LEFT JOIN `%s` %s ON %s", table, alias, on)

    return s
}

func (s *Stmt) InnerJoin(table, alias, on string) *Stmt {
    s.join = fmt.Sprintf("INNER JOIN `%s` %s ON %s", table, alias, on)

    return s
}

func (s *Stmt) RightJoin(table, alias, on string) *Stmt {
    s.join = fmt.Sprintf("RIGHT JOIN `%s` %s ON %s", table, alias, on)

    return s
}

func (s *Stmt) Offset(offset int64) *Stmt {
    s.offset = offset

    return s
}

func (s *Stmt) Limit(limit int64) *Stmt {
    s.limit = limit

    return s
}

func (s *Stmt) Group(group string) *Stmt {
    s.group = group

    return s
}

func (s *Stmt) Order(col, order string) *Stmt {
    s.order = fmt.Sprintf("ORDER BY %s %s", col, order)

    return s
}

func (s *Stmt) Where(where string) *Stmt {
    s.where = "WHERE " + strings.Trim(where, "()")

    return s
}

func (s *Stmt) And(exps ...string) string {
    return s.condition(exps, "AND")
}

func (s *Stmt) Or(exps ...string) string {
    return s.condition(exps, "OR")
}

func (s *Stmt) condition(exps []string, oper string) string {
    var parts = make([]string, 0, len(exps))
    for _,exp := range exps {
        if info := stmtWhereSpace.Split(strings.Trim(exp, " "), 4)
                len(info) == 3 {

            s.placeholders = append(s.placeholders, info[2])
            parts = append(parts, fmt.Sprintf("%s %s ?", info[0], info[1]))
        }
    }

    if len(parts) > 1 {
        return "(" + strings.Join(parts, " " + oper + " ") + ")"
    }

    return parts[0]
}

func (s *Stmt) Values(v map[string]interface{}) *Stmt {
    for _,p := range s.placeholders {
        s.values = append(s.values, v[p])
    }

    return s
}

func (s *Stmt) String() string {
    if len(s.cols) > 0 {
        stmt := fmt.Sprintf("SELECT %s FROM %s %s %s %s %s", s.cols, s.from, s.join, s.where, s.group, s.order)

        if s.limit != 0 {
            stmt = fmt.Sprintf("%s LIMIT %d,%d", stmt, s.offset, s.limit)
        }

        return stmt
    }

    return ""
}
