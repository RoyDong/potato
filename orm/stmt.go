package orm

import (
    "fmt"
    "regexp"
    "strings"
)


var (
    placeholderRegexp = regexp.MustCompile(`:\w+`)
)

type Stmt struct {
    cols []string
    allSelected []string
    from, alias string
    orms map[string]string
    names map[string]string
    joins []string
    where, group, order string
    offset, limit int64
    placeholders []string
}

func NewStmt() *Stmt {
    return &Stmt {
        allSelected: make([]string, 0, 5),
        orms: make(map[string]string, 5),
        names: make(map[string]string, 5),
        joins: make([]string, 0, 5),
        placeholders: make([]string, 0, 5),
    }
}

func (s *Stmt) Select(cols string) *Stmt {
    parts := strings.Split(cols, ",")
    s.cols = make([]string, 0, 20)
    for _,part := range parts {
        part = strings.Trim(part, " ")
        if col := strings.Split(part, "."); len(col) == 2 {
            if len(col[0]) > 0 && len(col[1]) > 0 {
                if col[1] == "*" {
                    s.allSelected = append(s.allSelected, col[0])
                } else {
                    s.cols = append(s.cols, fmt.Sprintf(
                        "`%s`.`%s` _%s_%s", col[0], col[1], col[0], col[1]))
                }
            }
        }
    }

    return s
}

func (s *Stmt) From(name, alias string) *Stmt {
    s.from = name
    s.alias = alias
    s.names[s.alias] = s.from

    return s
}

func (s *Stmt) LeftJoin(name, alias, on string) *Stmt {
    return s.Join("LEFT", name, alias, on)
}


func (s *Stmt) InnerJoin(name, alias, on string) *Stmt {
    return s.Join("INNER", name, alias, on)
}

func (s *Stmt) RightJoin(name, alias, on string) *Stmt {
    return s.Join("RIGHT", name, alias, on)
}

func (s *Stmt) Join(t, name, alias, on string) *Stmt {
    if table, ok := tables[name]; ok {
        s.joins = append(s.joins, fmt.Sprintf(" %s JOIN `%s` %s ON %s", t, table, alias, on))
        s.orms[name] = alias
        s.names[alias] = name
    }

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
    s.order = fmt.Sprintf(" ORDER BY %s %s", col, order)

    return s
}


func (s *Stmt) Where(where string) *Stmt {
    s.where = " WHERE " + where

    return s
}

func (s *Stmt) String() string {
    for _,k := range s.allSelected {
        n := s.names[k]
        if len(n) == 0 {
            continue
        }

        m, ok := models[n]
        if !ok {
            continue
        }

        for _,c := range m.Cols {
            s.cols = append(s.cols, fmt.Sprintf(
                    "`%s`.`%s` _%s_%s", k, c, k, c))
        }
    }

    if len(s.cols) > 0 {
        stmt := fmt.Sprintf("SELECT %s FROM %s %s%s%s%s%s",
                strings.Join(s.cols, ","), tables[s.from], s.alias,
                strings.Join(s.joins, ""), s.where, s.group, s.order)

        if s.limit != 0 {
            stmt = fmt.Sprintf("%s LIMIT %d,%d", stmt, s.offset, s.limit)
        }

        stmt = placeholderRegexp.
                ReplaceAllStringFunc(stmt, func(sub string) string {

            s.placeholders = append(s.placeholders,
                    strings.TrimLeft(sub, ":"))
            return "?"
        })

        return stmt
    }

    return ""
}

func (s *Stmt) Query(params map[string]interface{}) (*Rows, error) {
    stmt := s.String()
    L.Println(stmt)
    values := make([]interface{}, 0, len(params))
    for _,p := range s.placeholders {
        if v, ok := params[p]; ok {
            values = append(values, v)
        }
    }

    if len(values) != len(s.placeholders) {
        panic("orm: missing stmt params")
    }

    rows, e := D.Query(stmt, values...)
    if e != nil {
        return nil, e
    }

    columns, e := rows.Columns()
    if e != nil {
        return nil, e
    }

    alias := make(map[string]string, len(s.orms) + 1)
    alias[s.from] = s.alias
    for n, a := range s.orms {
        alias[n] = a
    }

    return &Rows{rows, alias, columns}, nil
}
