package orm

import (
    "fmt"
    "log"
    "time"
    "regexp"
    "strings"
    "reflect"
    "database/sql"
)


var (
    placeholderRegexp = regexp.MustCompile(`:\w+`)
)


type Stmt struct {
    cols []string
    from, alias string
    orms map[string]string
    joins []string
    where, group, order string
    offset, limit int64
    placeholders []string

    result map[string]map[int64]reflect.Value
    fields map[string]reflect.Value
    times map[string]reflect.Value
    index int64
}

func NewStmt() *Stmt {
    return &Stmt {
        orms: make(map[string]string, 5),
        joins: make([]string, 0, 5),
        placeholders: make([]string, 0, 5),
    }
}

func (s *Stmt) Select(cols string) *Stmt {
    parts := strings.Split(cols, ",")
    s.cols = make([]string, 0, len(parts))
    for _,part := range parts {
        part = strings.Trim(part, " ")
        if col := strings.Split(part, "."); len(col) == 2 {
            if len(col[0]) > 0 && len(col[1]) > 0 {
                s.cols = append(s.cols, fmt.Sprintf(
                        "`%s`.`%s` _%s_%s", col[0], col[1], col[0], col[1]))
            }
        }
    }

    return s
}

func (s *Stmt) From(name, alias string) *Stmt {
    s.from = name
    s.alias = alias

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
        s.joins = append(s.joins, fmt.Sprintf("%s JOIN `%s` %s ON %s", t, table, alias, on))
        s.orms[name] = alias
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
    s.order = fmt.Sprintf("ORDER BY %s %s", col, order)

    return s
}


func (s *Stmt) Where(where string) *Stmt {
    s.where = "WHERE " + where

    return s
}

func (s *Stmt) String() string {
    if len(s.cols) > 0 {
        stmt := fmt.Sprintf("SELECT %s FROM %s %s %s %s %s %s",
                strings.Join(s.cols, ","), tables[s.from], s.alias, strings.Join(s.joins, " "), s.where, s.group, s.order)

        if s.limit != 0 {
            stmt = fmt.Sprintf("%s LIMIT %d,%d", stmt, s.offset, s.limit)
        }

        stmt = placeholderRegexp.ReplaceAllStringFunc(stmt, func(sub string) string {
            s.placeholders = append(s.placeholders, strings.TrimLeft(sub, ":"))

            return "?"
        })

        return stmt
    }

    return ""
}

func (s *Stmt) Alias(name string) string {
    if name == s.from {
        return s.alias
    }

    return s.orms[name]
}

func (s *Stmt) Query(params map[string]interface{}) *Stmt {
    stmt := s.String()
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
        L.Println(e)
    }

    if e := s.scan(rows); e != nil {
        L.Println(e)
    }

    return s
}

func (s *Stmt) scan(rows *sql.Rows) error {
    columns, e := rows.Columns()
    if e != nil {
        return e
    }

    length := len(columns)
    s.result = make(map[string]map[int64]reflect.Value, len(s.orms) + 1)
    s.result[s.from] = make(map[int64]reflect.Value, length)
    for n := range s.orms {
        s.result[n] = make(map[int64]reflect.Value, length)
    }

    for rows.Next() {
        if e := s.scanRow(rows, columns); e != nil {
            return e
        }
    }

    return nil
}

func (s *Stmt) scanRow(rows *sql.Rows, columns []string) error {
    fields := make(map[string]reflect.Value, len(columns))
    times := make(map[string]reflect.Value, 10)
    ids := make(map[string]*int64, len(s.orms) + 1)
    values := make(map[string]reflect.Value, len(s.orms) + 1)
    values[s.from] = reflect.Indirect(reflect.New(models[s.from].Entity))
    for n := range s.orms {
        values[n] = reflect.Indirect(reflect.New(models[n].Entity))
    }

    for _,v := range values {
        for i := 0; i < v.NumField(); i++ {
            vt := v.Type()
            ft := vt.Field(i)
            if col := ft.Tag.Get("column"); len(col) > 0 {
                k := fmt.Sprintf("_%s_%s", s.Alias(vt.Name()), col)
                if col == "id" {
                    var id int64
                    ids[k] = &id
                } else if ft.Type.Name() == "Time" {
                    times[k] = v.Field(i)
                } else {
                    fields[k] = v.Field(i)
                }
            }
        }
    }

    index := make(map[string]int, len(times))
    row := make([]interface{}, 0, len(columns))
    for i, k := range columns {
        if f, ok := fields[k]; ok {
            row = append(row, f.Addr().Interface())
        } else if f, ok := ids[k]; ok {
            row = append(row, f)
        } else {
            var v int64
            row = append(row, &v)
            index[k] = i
        }
    }

    if e := rows.Scan(row...); e != nil {
        return e
    }

    for k, i := range index {
        v := row[i].(*int64)
        times[k].Set(reflect.ValueOf(time.Unix(0, *v)))
    }

    for n, res := range s.result {
        id := *(ids[fmt.Sprintf("_%s_id", s.Alias(n))])
        if _, ok := res[id]; !ok {
            res[id] = values[n]
        }
    }

    return nil
}
