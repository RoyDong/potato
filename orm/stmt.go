package orm

import (
    "fmt"
    "time"
    "regexp"
    "strings"
    "reflect"
    "database/sql"
)


var (
    placeholderRegexp = regexp.MustCompile(`:\w+`)
    columnAliasRegexp = regexp.MustCompile(`^_([a-zA-Z]+)_(.*)$`)

    columnTypes = []string{"bool", "int64", "float64", "string", "time"}
    tables = map[string]string{"User": "user"}
    models = make(map[string]*Model)
)


type Stmt struct {
    cols []string
    alias map[string]string
    from, join, where, group, order string
    offset, limit int64
    placeholders []string
    values []interface{}
}

func NewStmt() *Stmt {
    return &Stmt {
        alias: make(map[string]string),
        placeholders: make([]string, 0),
    }
}


func (s *Stmt) Select(cols string) *Stmt {
    parts := strings.Split(cols, ",")
    s.cols = make([]string, 0, len(parts))
    for _,part := range parts {
        if col := strings.Split(part, "."); len(col) == 2 {
            t := strings.Trim(col[0], " `")
            c := strings.Trim(col[1], " `")
            if len(t) > 0 && len(c) > 0 {
                s.cols = append(s.cols, fmt.Sprintf("`%s`.`%s` _%s_%s", t, c, t, c))
            }
        }
    }

    return s
}


func (s *Stmt) From(name, alias string) *Stmt {
    if table, ok := tables[name]; ok {
        s.from = fmt.Sprintf("`%s` %s", table, alias)
        s.alias[name] = alias
    }

    return s
}


func (s *Stmt) LeftJoin(name, alias, on string) *Stmt {
    return s.Join("left", name, alias, on)
}


func (s *Stmt) InnerJoin(name, alias, on string) *Stmt {
    return s.Join("inner", name, alias, on)
}


func (s *Stmt) RightJoin(name, alias, on string) *Stmt {
    return s.Join("right", name, alias, on)
}

func (s *Stmt) Join(t, name, alias, on string) *Stmt {
    if table, ok := tables[name]; ok {
        s.join = fmt.Sprintf("%s JOIN `%s` %s ON %s", strings.ToUpper(t), table, alias, on)
        s.alias[name] = alias
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

func (s *Stmt) Values(v map[string]interface{}) *Stmt {
    for _,p := range s.placeholders {
        s.values = append(s.values, v[p])
    }

    return s
}

func (s *Stmt) String() string {
    if len(s.cols) > 0 {
        stmt := fmt.Sprintf("SELECT %s FROM %s %s %s %s %s",
                strings.Join(s.cols, ","), s.from, s.join, s.where, s.group, s.order)

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
    for _,n := range s.placeholders {
        if v, ok := params[n]; ok {
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

    return &Rows{rows, columns, s.alias}
}

type Rows struct {
    *sql.Rows
    columns []string
    alias map[string]string
}


func (r *Rows) Scan(args ...interface{}) error {
    row := make([]interface{}, 0, len(r.columns))
    for _= range r.columns {
        var v interface{}
        row = append(row, &v)
    }

    r.Rows.Next()
    if e := r.Rows.Scan(result...); e != nil {
        return e
    }

    length := 0
    values := make([]reflect.Value, 0, len(args))
    for _,arg := range args {
        v := reflect.Indirect(reflect.ValueOf(arg))
        length = length + v.NumField()
        values = append(values, v)
    }

    fields := make(map[string]reflect.Value, length)
    for _,v := range values {
        for i := 0; i < v.NumField(); i++ {
            t := v.Type()
            col := fmt.Sprintf("_%s_%s", r.alias[t.Name()], t.Field(i).Tag.Get("name"))
            row[col] = v.Field(i)
        }
    }

    for i, col := range r.columns {
        if raw := reflect.Indirect(reflect.ValueOf(row[i])).Interface(); raw != nil {
            v := reflect.ValueOf(raw)
            f := fields[col]

            switch f.Kind() {
            case reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
                f.SetInt(v.Int())
            case reflect.String:
                f.SetString(string(v.Bytes()))
            case reflect.Bool:
                f.SetBool(v.Int() > 0)
            case reflect.Struct:
                t := time.Unix(0, v.Int())
                f.Set(reflect.ValueOf(t))
            case reflect.Float32, reflect.Float64:
                f.SetFloat(v.Float())
            }
        }
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
        tag := elem.Field(i).Tag
        if name := tag.Get("name"); len(name) > 0 {
            cols = append(cols, name)
            colsType = append(colsType, tag.Get("type"))
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
