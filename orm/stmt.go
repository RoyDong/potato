package orm

import (
    "fmt"
    "regexp"
    "strings"
    "database/sql"
)


const (
    ActionCount  = 0
    ActionSelect = 1
    ActionInsert = 2
    ActionUpdate = 3
    ActionDelete = 4
)

var (
    placeholderRegexp = regexp.MustCompile(`:\w+`)
)

type Stmt struct {
    action int
    cols []string
    allSelected []string
    from, alias string
    orms map[string]string
    names map[string]string
    joins []string
    updates []string
    orders []string
    where, group string
    offset, limit int64
    placeholders []string

    stmt string
}

func NewStmt() *Stmt {
    return new(Stmt).Clear()
}

func (s *Stmt) Clear() *Stmt {
    s.allSelected  = make([]string, 0, 5)
    s.orms         = make(map[string]string, 5)
    s.names        = make(map[string]string, 5)
    s.joins        = make([]string, 0, 5)
    s.updates      = make([]string, 0, 10)
    s.orders       = make([]string, 0, 5)
    s.placeholders = make([]string, 0, 5)

    s.cols   = nil
    s.from   = ""
    s.alias  = ""
    s.where  = ""
    s.group  = ""
    s.offset = 0
    s.limit  = 0
    s.stmt   = ""

    return s
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
                    s.cols = append(s.cols,fmt.Sprintf("`%s`.`%s` _%s_%s",
                            col[0], col[1], col[0], col[1]))
                }
            }
        }
    }

    s.action = ActionSelect
    return s
}

func (s *Stmt) parseTableAlias(table string) (string, string) {
    r := strings.SplitN(strings.Trim(table, " "), " ", 2)
    t := r[0]
    a := strings.Trim(r[1], " ")

    if len(t) == 0 || len(a) == 0 {
        panic("orm: table name and its alias not right")
    }

    return t, a
}

func (s *Stmt) Count(table string) *Stmt {
    s.from, s.alias = s.parseTableAlias(table)
    s.names[s.alias] = s.from
    s.action = ActionCount

    return s
}

func (s *Stmt) Delete(table string) *Stmt {
    s.from = table
    s.action = ActionDelete

    return s
}

func (s *Stmt) Update(table string) *Stmt {
    s.from, s.alias = s.parseTableAlias(table)
    s.names[s.alias] = s.from
    s.action = ActionUpdate

    return s
}

func (s *Stmt) Set(u string) *Stmt {
    s.updates = append(s.updates, u)

    return s
}

func (s *Stmt) From(table string) *Stmt {
    s.from, s.alias = s.parseTableAlias(table)
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
        s.joins = append(s.joins,
                fmt.Sprintf(" %s JOIN `%s` %s ON %s", t, table, alias, on))
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

func (s *Stmt) Desc(col string) *Stmt {
    s.orders = append(s.orders, fmt.Sprintf("%s DESC", col))

    return s
}

func (s *Stmt) Asc(col string) *Stmt {
    s.orders = append(s.orders, fmt.Sprintf("%s ASC", col))

    return s
}


func (s *Stmt) Where(where string) *Stmt {
    s.where = " WHERE " + where

    return s
}

func (s *Stmt) deleteStmt() string {
    s.stmt = fmt.Sprintf("DELETE FROM %s %s%s",
            tables[s.from], s.where, s.order())

    if s.limit != 0 {
        s.stmt = fmt.Sprintf("%s LIMIT %d,%d", s.stmt, s.offset, s.limit)
    }

    s.stmt = placeholderRegexp.ReplaceAllStringFunc(s.stmt, s.replph)

    return s.stmt
}

func (s *Stmt) updateStmt() string {
    s.stmt = fmt.Sprintf("UPDATE %s %s%s SET %s%s%s",
            tables[s.from], s.alias, strings.Join(s.joins, ""),
            strings.Join(s.updates, ","), s.where, s.order())

    if s.limit != 0 {
        s.stmt = fmt.Sprintf("%s LIMIT %d,%d", s.stmt, s.offset, s.limit)
    }

    s.stmt = placeholderRegexp.ReplaceAllStringFunc(s.stmt, s.replph)

    return s.stmt
}

func (s *Stmt) countStmt() string {
    s.stmt = fmt.Sprintf("SELECT COUNT(*) num FROM %s %s%s%s%s", tables[s.from],
            s.alias, strings.Join(s.joins, ""), s.where, s.group)
    s.stmt = placeholderRegexp.ReplaceAllStringFunc(s.stmt, s.replph)

    return s.stmt
}

func (s *Stmt) selectStmt() string {
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
        s.stmt = fmt.Sprintf("SELECT %s FROM %s %s%s%s%s%s",
                strings.Join(s.cols, ","), tables[s.from], s.alias,
                strings.Join(s.joins, ""), s.where, s.group, s.order())

        if s.limit != 0 {
            s.stmt = fmt.Sprintf("%s LIMIT %d,%d", s.stmt, s.offset, s.limit)
        }

        s.stmt = placeholderRegexp.ReplaceAllStringFunc(s.stmt, s.replph)
        return s.stmt
    }

    panic("orm: error format sql for potato orm")
}

func (s *Stmt) replph(sub string) string {
    s.placeholders = append(s.placeholders, strings.TrimLeft(sub, ":"))
    return "?"
}

func (s *Stmt) order() string {
    if len(s.orders) > 0 {
        return fmt.Sprintf(" ORDER BY %s", strings.Join(s.orders, ","))
    }

    return ""
}

func (s *Stmt) values(params map[string]interface{}) []interface{} {
    values := make([]interface{}, 0, len(params))
    for _,p := range s.placeholders {
        if v, ok := params[p]; ok {
            values = append(values, v)
        }
    }

    if len(values) != len(s.placeholders) {
        panic("orm: missing stmt params")
    }

    return values
}

func (s *Stmt) Exec(params map[string]interface{}) int64 {
    var n int64
    if s.action == ActionCount {
        rows, e := D.Query(s.countStmt(), s.values(params)...)
        if e != nil {
            L.Println(e)
            return 0
        }

        rows.Next()
        rows.Scan(&n)
        return n
    }

    var result sql.Result
    var e error
    if s.action == ActionUpdate {
        result, e = D.Exec(s.updateStmt(), s.values(params)...)
    }

    if s.action == ActionDelete {
        result, e = D.Exec(s.deleteStmt(), s.values(params)...)
    }

    if e != nil {
        L.Println(e)
        return 0
    }

    n, e = result.RowsAffected()
    if e!= nil {
        L.Println(e)
        return 0
    }

    return n
}

func (s *Stmt) Query(params map[string]interface{}) (*Rows, error) {
    rows, e := D.Query(s.selectStmt(), s.values(params)...)
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

func (s *Stmt) String() string {
    return s.stmt
}
