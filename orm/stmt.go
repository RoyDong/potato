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

    columnTypes = []string{"bool", "int", "float", "string", "time"}

    tables = map[string]string{"User": "user"}

    models = make(map[string]*Model)
)


type Stmt struct {
    cols []string
    from, alias string
    orms map[string]string
    joins []string
    where, group, order string
    offset, limit int64
    placeholders []string
    values []interface{}
}

func NewStmt() *Stmt {
    return &Stmt {
        orms: make(map[string]string),
        joins: make([]string, 0),
        placeholders: make([]string, 0),
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

func (s *Stmt) Query(params map[string]interface{}) *Rows {
    stmt := s.String()
    values := make([]interface{}, 0, len(params))
    for _,p := range s.placeholders {
        if v, ok := params[p]; ok {
            values = append(values, v)
        }
    }

    if len(values) != len(s.placeholders) {
        panic("stmt query: missing params")
    }

    rows, e := D.Query(stmt, values...)
    if e != nil {
        L.Println(e)
        return nil
    }

    columns, e := rows.Columns()
    if e != nil {
        L.Println(e)
        return nil
    }

    r := &Rows{rows, columns, s.orms}
    r.alias[s.from] = s.alias
    return r
}

type Rows struct {
    *sql.Rows
    columns []string
    alias map[string]string
}


func (r *Rows) Scan(args ...interface{}) error {
    length := 0
    values := make([]reflect.Value, 0, len(args))
    for _,arg := range args {
        v := reflect.Indirect(reflect.ValueOf(arg))
        length = length + v.NumField()
        values = append(values, v)
    }

    fields := make(map[string]reflect.Value, length)
    times := make([]reflect.Value, 0, length)
    for _,v := range values {
        for i := 0; i < v.NumField(); i++ {
            t := v.Type()
            f := t.Field(i)
            col := fmt.Sprintf("_%s_%s", r.alias[t.Name()], f.Tag)

            if f.Type.Name() == "Time" {
                times = append(times, v.Field(i))
            } else {
                fields[col] = v.Field(i)
            }
        }
    }

    row := make([]interface{}, 0, len(r.columns))
    index := make([]int, 0, len(times))
    for i, col := range r.columns {
        if f, ok := fields[col]; ok {
            row = append(row, f.Addr().Interface())
        } else {
            var v int64
            row = append(row, &v)
            index = append(index, i)
        }
    }

    r.Rows.Next()
    if e := r.Rows.Scan(row...); e != nil {
        return e
    }

    for i, k := range index {
        v := row[k].(*int64)
        t := time.Unix(0, *v)
        times[i].Set(reflect.ValueOf(t))
    }

    return nil
}


type Model struct {
    table string
    cols, colsType []string
    entity reflect.Type
}

func NewModel(table string, v interface{}) *Model {
    elem := reflect.Indirect(reflect.ValueOf(v)).Type()
    length := elem.NumField()
    cols := make([]string, 0, length)
    colsType := make([]string, 0, length)

    for i := 0; i < length; i++ {
        if tag := elem.Field(i).Tag; len(tag) > 0 {
            cols = append(cols, string(tag))
        }
    }

    model := &Model{table, cols, colsType, elem}
    tables[elem.Name()] = table
    models[elem.Name()] = model

    return model
}

func (m *Model) ColumnIndex(col string) int {
    for i, v := range m.cols {
        if v == col {
            return i
        }
    }

    return -1
}
