package orm

import (
    "fmt"
    "reflect"
    "strings"
    "time"
)

var (
    tables = make(map[string]string)
    models = make(map[string]*Model)

    ColumnTagName = "column"
)

type Model struct {
    Table  string
    Cols   []string
    Entity reflect.Type
}

func NewModel(table string, v interface{}) *Model {
    t := reflect.Indirect(reflect.ValueOf(v)).Type()
    n := t.NumField()
    cols := make([]string, 0, n)
    for i := 0; i < n; i++ {
        if col := t.Field(i).Tag.Get(ColumnTagName); len(col) > 0 {
            cols = append(cols, col)
        }
    }

    model := &Model{table, cols, t}
    tables[t.Name()] = table
    models[t.Name()] = model
    return model
}

func (m *Model) Save(entity interface{}) bool {
    return Save(entity)
}

func Save(entity interface{}) bool {
    val := reflect.Indirect(reflect.ValueOf(entity))
    typ := val.Type()
    name := typ.Name()
    tbl := tables[name]
    if len(tbl) == 0 {
        panic("orm: entity not found")
    }

    var pk reflect.Value
    pkv := int64(-1)
    n := typ.NumField()
    cols := make([]string, 0, n)
    vals := make([]interface{}, 0, n)
    for i := 0; i < n; i++ {
        f := val.Field(i)
        col := typ.Field(i).Tag.Get(ColumnTagName)
        if len(col) > 0 {
            ifc := f.Interface()
            cols = append(cols, col)

            if t, ok := ifc.(time.Time); ok {
                vals = append(vals, t.UnixNano())
            } else {
                vals = append(vals, ifc)
            }

            if col == "id" {
                pk = f
                pkv = f.Int()
            }
        }
    }

    if pkv < 0 {
        panic("orm: primary key not specified in any field tag")
    }

    if pkv == 0 {
        cs := make([]string, 0, n)
        ph := make([]string, 0, n)
        for _, col := range cols {
            cs = append(cs, fmt.Sprintf("`%s`", col))
            ph = append(ph, "?")
        }

        stmt := fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)",
            tbl, strings.Join(cs, ","), strings.Join(ph, ","))
        result, e := D.Exec(stmt, vals...)
        if e != nil {
            L.Println(e)
            return false
        }

        n, e := result.LastInsertId()
        if e != nil {
            L.Println(e)
            return false
        }

        pk.SetInt(n)
        return true
    }

    sets := make([]string, 0, n)
    for _, col := range cols {
        sets = append(sets, fmt.Sprintf("`%s` = ?", col))
    }

    stmt := fmt.Sprintf("UPDATE `%s` SET %s WHERE `id` = %d",
        tbl, strings.Join(sets, ","), pkv)
    if _, e := D.Exec(stmt, vals...); e != nil {
        L.Println(e)
        return false
    }

    return true
}
