package orm

import (
    "fmt"
    "log"
    "regexp"
    "strings"
    "reflect"
    "database/sql"
)

type Scanner interface {
    Scan(args ...interface{}) error
}

var stmtWhereSpace = regexp.MustCompile(`\s+`)

type Stmt struct {
    cols []string
    alias map[string]string
    from, join, where, group, order string
    offset, limit int64
    placeholders []string
    values []interface{}
}

func (s *Stmt) Select(cols string) *Stmt {
    parts := strings.Split(cols, ",")
    s.cols = make([]string, 0, len(parts))
    for _,part := range parts {
        if col := strings.Split(part, "."); len(col) == 2 {
            t := strings.Trim(col[0], " `\n")
            c := strings.Trim(col[1], " `\n")
            if len(t) > 0 && len(c) > 0 {
                s.cols = append(s.cols, fmt.Sprintf("`%s`.`%s` _%s_%s", t, c, t, c))
            }
        }
    }

    return s
}

func (s *Stmt) From(name, alias string) *Stmt {
    s.from = fmt.Sprintf("`%s` %s", name, alias)
    s.alias[name] = alias

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

type Select struct {
    cols []string
}


type Rows struct {
    sql.Rows
}

var columnRegexp = regexp.MustCompile(`^_(\d+)_(\d+)`)

func (r *Rows) Scan(args ...interface{}) {
    for _,v := range args {
        elem := reflect.TypeOf(v).Elem()
    }

    for _,col := range r.Columns() {
        var alias, name string
        if info := columnRegexp.FindStringSubmatch(col); len(info) == 2 {
            alias = info[0]
            name = info[1]
        } else {
            name = col 
        }
    
        i := 
    }
}


type Model struct {
    table string
    cols, colsType []string
    entity reflect.Type
}

var tables = make(map[string]string)
var models = make(map[string]*Model)

func NewModel(table string, v interface{}) *Model {
    elem := reflect.TypeOf(v).Elem()
    cols := make([]string, 0, elem.NumField())
    colsType := make([]string, 0, elem.NumField())

    for i := 0; i < elem.NumField(); i++ {
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