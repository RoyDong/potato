package orm

import (
    "database/sql"
    "fmt"
    "strings"
    "time"
)

const (
    ActionCount  = 0
    ActionSelect = 1
    ActionInsert = 2
    ActionUpdate = 3
    ActionDelete = 4
)

type Stmt struct {
    action int

    cols        []string
    allSelected []string

    alias map[string]string
    names map[string]string

    from   []string
    joins  []string
    orders []string

    distinct, where, group, having string

    offset, limit int64
}

func NewStmt() *Stmt {
    return new(Stmt).init()
}

func (s *Stmt) init() *Stmt {
    s.allSelected = make([]string, 0)
    s.alias = make(map[string]string)
    s.names = make(map[string]string)
    s.joins = make([]string, 0)
    s.orders = make([]string, 0)

    return s
}

func (s *Stmt) Clear() *Stmt {
    s.init()
    s.cols = nil
    s.from = nil
    s.distinct = ""
    s.where = ""
    s.group = ""
    s.having = ""
    s.offset = 0
    s.limit = 0

    return s
}

func (s *Stmt) Distinct(c string) *Stmt {
    s.distinct = "DISTINCT "
    return s.Select(c)
}

func (s *Stmt) Select(c string) *Stmt {
    parts := strings.Split(c, ",")
    s.cols = make([]string, 0, len(parts))
    for _, part := range parts {
        part = strings.Trim(part, " ")
        if col := strings.Split(part, "."); len(col) == 2 {
            if len(col[0]) > 0 && len(col[1]) > 0 {
                if col[1] == "*" {
                    s.allSelected = append(s.allSelected, col[0])
                } else {
                    s.cols = append(s.cols, fmt.Sprintf("`%s`.`%s` _%s_%s",
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

func (s *Stmt) Insert(name string, cols ...string) *Stmt {
    s.cols = cols
    s.From(name, "")
    s.action = ActionInsert
    return s
}

func (s *Stmt) Update(name, alias string, cols ...string) *Stmt {
    s.cols = cols
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

func (s *Stmt) table() string {
    table, ok := tables[s.from[0]]
    if !ok {
        panic(fmt.Sprintf("orm: model %s not found", s.from[0]))
    }

    return table
}

func (s *Stmt) countStmt() string {
    return fmt.Sprintf("SELECT COUNT(*) num FROM `%s` %s%s%s%s%s",
        s.table(), s.from[1], strings.Join(s.joins, ""),
        s.where, s.group, s.having)
}

func (s *Stmt) selectStmt() string {
    for _, a := range s.allSelected {
        n := s.names[a]
        if len(n) == 0 {
            continue
        }

        m, ok := models[n]
        if !ok {
            continue
        }

        for _, c := range m.Cols {
            s.cols = append(s.cols, fmt.Sprintf(
                "`%s`.`%s` _%s_%s", a, c, a, c))
        }
    }

    if len(s.cols) == 0 {
        panic("orm: columns format error while using Stmt.Select")
    }

    stmt := fmt.Sprintf("SELECT %s%s FROM `%s` %s%s%s%s%s%s",
        s.distinct, strings.Join(s.cols, ","), s.table(), s.from[1],
        strings.Join(s.joins, ""), s.where, s.group, s.having, s.order())

    if s.limit != 0 {
        stmt = fmt.Sprintf("%s LIMIT %d,%d", stmt, s.offset, s.limit)
    }

    return stmt
}

func (s *Stmt) updateStmt() string {
    updates := make([]string, 0, len(s.cols))
    for _, col := range s.cols {
        updates = append(updates, fmt.Sprintf("`%s`.`%s` = ?", s.from[1], col))
    }

    stmt := fmt.Sprintf("UPDATE `%s` %s%s SET %s%s%s",
        s.table(), s.from[1], strings.Join(s.joins, ""),
        strings.Join(updates, ","), s.where, s.order())

    if s.limit != 0 {
        stmt = fmt.Sprintf("%s LIMIT %d,%d", stmt, s.offset, s.limit)
    }

    return stmt
}

func (s *Stmt) deleteStmt() string {
    stmt := fmt.Sprintf(
        "DELETE FROM `%s` %s%s", s.table(), s.where, s.order())

    if s.limit != 0 {
        stmt = fmt.Sprintf("%s LIMIT %d,%d", stmt, s.offset, s.limit)
    }

    return stmt
}

func (s *Stmt) insert(args ...interface{}) (int64, error) {
    n := len(args)
    if n != len(s.cols) {
        panic("orm: args not match column num while using Stmt.Insert")
    }

    c := make([]string, 0, n)
    h := make([]string, 0, n)
    v := make([]interface{}, 0, n)
    for i, col := range s.cols {
        c = append(c, fmt.Sprintf("`%s`", col))
        h = append(h, "?")

        val := args[i]
        if t, ok := val.(time.Time); ok {
            v = append(v, t.UnixNano())
        } else {
            v = append(v, val)
        }
    }

    stmt := fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)",
        s.table(), strings.Join(c, ","), strings.Join(h, ","))

    result, e := D.Exec(stmt, v...)
    if e != nil {
        return 0, e
    }

    id, e := result.LastInsertId()
    return id, e
}

func (s *Stmt) Exec(args ...interface{}) (int64, error) {
    for i, v := range args {
        if t, ok := v.(time.Time); ok {
            args[i] = t.UnixNano()
        }
    }

    var n int64
    if s.action == ActionCount {
        row := D.QueryRow(s.countStmt(), args...)
        e := row.Scan(&n)
        return n, e
    }

    var result sql.Result
    var e error
    if s.action == ActionInsert {
        n, e = s.insert(args...)
        return n, e
    }

    if s.action == ActionUpdate {
        result, e = D.Exec(s.updateStmt(), args...)
    }

    if s.action == ActionDelete {
        result, e = D.Exec(s.deleteStmt(), args...)
    }

    if e != nil {
        return 0, e
    }

    n, e = result.RowsAffected()
    return n, e
}

func (s *Stmt) Query(args ...interface{}) (*Rows, error) {
    rows, e := D.Query(s.selectStmt(), args...)
    if e != nil {
        return nil, e
    }

    columns, e := rows.Columns()
    if e != nil {
        return nil, e
    }

    return &Rows{rows, s.alias, columns}, nil
}
