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

    alias map[string]string
    names map[string]string

    from    []string
    joins   []string
    updates []string
    orders  []string

    distinct, where, group, having string

    offset, limit int64

    placeholders []string

    stmt string
}

func NewStmt() *Stmt {
    return new(Stmt).init()
}

func (s *Stmt) init() *Stmt {
    s.allSelected  = make([]string, 0, 5)
    s.alias        = make(map[string]string, 4)
    s.names        = make(map[string]string, 4)
    s.joins        = make([]string, 0, 3)
    s.updates      = make([]string, 0, 5)
    s.orders       = make([]string, 0, 2)
    s.placeholders = make([]string, 0, 5)

    return s
}

func (s *Stmt) Clear() *Stmt {
    s.init()
    s.cols     = nil
    s.from     = nil
    s.distinct = ""
    s.where    = ""
    s.group    = ""
    s.having   = ""
    s.stmt     = ""
    s.offset   = 0
    s.limit    = 0

    return s
}

func (s *Stmt) Distinct(c string) *Stmt {
    s.distinct = "DISTINCT "
    return s.Select(c)
}

func (s *Stmt) Select(c string) *Stmt {
    parts := strings.Split(c, ",")
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

func (s *Stmt) From(name, alias string) *Stmt {
    s.from = []string{name, alias}
    if len(alias) > 0 {
        s.names[alias] = name
        s.alias[name] = alias
    }

    return s
}

func (s *Stmt) Count(name, alias string) *Stmt {
    s.From(name, alias)
    s.action = ActionCount

    return s
}

func (s *Stmt) Insert(name string) *Stmt {
    s.From(name, "")
    s.action = ActionInsert
    return s
}

func (s *Stmt) Update(name, alias string) *Stmt {
    s.From(name, alias)
    s.action = ActionUpdate

    return s
}

func (s *Stmt) Delete(name string) *Stmt {
    s.From(name, "")
    s.action = ActionDelete

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
        s.alias[name] = alias
        s.names[alias] = name

        return s
    }

    panic(fmt.Sprintf("orm: model %s not found", name))
}

func (s *Stmt) Set(u string) *Stmt {
    s.updates = append(s.updates, u)

    return s
}

func (s *Stmt) Where(where string) *Stmt {
    s.where = " WHERE " + where

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

func (s *Stmt) GroupBy(g string) *Stmt {
    s.group = " GROUP BY " + g

    return s
}

func (s *Stmt) Having(h string) *Stmt {
    s.having = " HAVING " + h

    return s
}

func (s *Stmt) order() string {
    if len(s.orders) > 0 {
        return fmt.Sprintf(" ORDER BY %s", strings.Join(s.orders, ","))
    }

    return ""
}

func (s *Stmt) Desc(col string) *Stmt {
    s.orders = append(s.orders, fmt.Sprintf("%s DESC", col))

    return s
}

func (s *Stmt) Asc(col string) *Stmt {
    s.orders = append(s.orders, fmt.Sprintf("%s ASC", col))

    return s
}

func (s *Stmt) replph(sub string) string {
    s.placeholders = append(s.placeholders, strings.TrimLeft(sub, ":"))
    return "?"
}

func (s *Stmt) countStmt(table string) {
    s.stmt = fmt.Sprintf("SELECT COUNT(*) num FROM `%s` %s%s%s%s%s",
            table, s.from[1], strings.Join(s.joins, ""),
            s.where, s.group, s.having)
}

func (s *Stmt) selectStmt(table string) {
    if len(s.from) != 2 || len(s.from[0]) == 0 {
        panic("orm: model name not specified while using Stmt.Update")
    }

    for _,a := range s.allSelected {
        n := s.names[a]
        if len(n) == 0 {
            continue
        }

        m, ok := models[n]
        if !ok {
            continue
        }

        for _,c := range m.Cols {
            s.cols = append(s.cols, fmt.Sprintf(
                    "`%s`.`%s` _%s_%s", a, c, a, c))
        }
    }

    if len(s.cols) == 0 {
        panic("orm: columns format error while using Stmt.Select")
    }

    s.stmt = fmt.Sprintf("SELECT %s%s FROM `%s` %s%s%s%s%s%s",
            s.distinct, strings.Join(s.cols, ","), table, s.from[1],
            strings.Join(s.joins, ""), s.where, s.group, s.having, s.order())

    if s.limit != 0 {
        s.stmt = fmt.Sprintf("%s LIMIT %d,%d", s.stmt, s.offset, s.limit)
    }
}

func (s *Stmt) updateStmt(table string) {
    s.stmt = fmt.Sprintf("UPDATE `%s` %s%s SET %s%s%s",
            table, s.from[1], strings.Join(s.joins, ""),
            strings.Join(s.updates, ","), s.where, s.order())

    if s.limit != 0 {
        s.stmt = fmt.Sprintf("%s LIMIT %d,%d", s.stmt, s.offset, s.limit)
    }
}

func (s *Stmt) deleteStmt(table string) {
    s.stmt = fmt.Sprintf(
            "DELETE FROM `%s` %s%s", table, s.where, s.order())

    if s.limit != 0 {
        s.stmt = fmt.Sprintf("%s LIMIT %d,%d", s.stmt, s.offset, s.limit)
    }
}

func (s *Stmt) String() string {
    if len(s.stmt) == 0 {
        if len(s.from) != 2 || len(s.from[0]) == 0 {
            panic("orm: model name not specified while using Stmt.Update")
        }

        table, ok := tables[s.from[0]]
        if !ok {
            panic(fmt.Sprintf("orm: model %s not found", s.from[0]))
        }

        switch s.action {
        case ActionCount:
            s.countStmt(table)
            break

        case ActionInsert:
            return "insert stmt will not generate before Stmt.Exec"

        case ActionSelect:
            s.selectStmt(table)
            break

        case ActionUpdate:
            s.updateStmt(table)
            break

        case ActionDelete:
            s.deleteStmt(table)
            break

        default:
            panic("orm: not supported stmt action")
        }

        s.stmt = placeholderRegexp.ReplaceAllStringFunc(s.stmt, s.replph)
    }

    return s.stmt
}

func (s *Stmt) values(params map[string]interface{}) []interface{} {
    values := make([]interface{}, 0, len(params))
    for _,p := range s.placeholders {
        if v, ok := params[p]; ok {
            values = append(values, v)
        } else {
            panic("orm: missing stmt param " + p)
        }
    }

    return values
}

func (s *Stmt) insert(params map[string]interface{}) (int64, error) {
    s.String()
    l := len(params)
    c := make([]string, 0, l)
    h := make([]string, 0, l)
    v := make([]interface{}, 0, l)
    for col, val := range params {
        c = append(c, fmt.Sprintf("`%s`", col))
        h = append(h, "?")
        v = append(v, val)
    }

    s.stmt = fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)",
            tables[s.from[0]], strings.Join(c, ","), strings.Join(h, ","))

    result, e := D.Exec(s.stmt, v...)
    if e != nil {
        return 0, e
    }

    n, e := result.LastInsertId()
    return n, e
}

func (s *Stmt) Exec(params map[string]interface{}) (int64, error) {
    var n int64
    if s.action == ActionCount {
        rows, e := D.Query(s.String(), s.values(params)...)
        if e != nil {
            return 0, e
        }

        if rows.Next() {
            e = rows.Scan(&n)
        }

        return n, e
    }

    var result sql.Result
    var e error
    if s.action == ActionInsert {
        n, e = s.insert(params)
        return n, e
    }

    if s.action == ActionUpdate {
        result, e = D.Exec(s.String(), s.values(params)...)
    }

    if s.action == ActionDelete {
        result, e = D.Exec(s.String(), s.values(params)...)
    }

    if e != nil {
        return 0, e
    }

    n, e = result.RowsAffected()
    return n, e
}

func (s *Stmt) Query(params map[string]interface{}) (*Rows, error) {
    rows, e := D.Query(s.String(), s.values(params)...)
    if e != nil {
        return nil, e
    }

    columns, e := rows.Columns()
    if e != nil {
        return nil, e
    }

    return &Rows{rows, s.alias, columns}, nil
}
