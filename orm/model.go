package orm

import (
    "fmt"
    "strings"
    "reflect"
)


var (
    tables = make(map[string]string)
    models = make(map[string]*Model)
)

type Model struct {
    Table string
    Cols []string
    Entity reflect.Type
}

func NewModel(table string, v interface{}) *Model {
    t := reflect.Indirect(reflect.ValueOf(v)).Type()
    length := t.NumField()
    cols := make([]string, 0, length)

    for i := 0; i < length; i++ {
        if col := t.Field(i).Tag.Get("column"); len(col) > 0 {
            cols = append(cols, col)
        }
    }

    model := &Model{table, cols, t}
    tables[t.Name()] = table
    models[t.Name()] = model

    return model
}

func (m *Model) Save(entity interface{}) bool {
    val := reflect.Indirect(reflect.ValueOf(entity))
    typ := val.Type()
    name := typ.Name()
    table, ok := tables[name]
    if !ok {
        panic("orm: entity not supported")
    }

    var pk reflect.Value
    var pkv = int64(-1)
    n    := typ.NumField()
    cols := make([]string, 0, n)
    vals := make([]interface{}, 0, n)
    for i := 0; i < n; i++ {
        f := val.Field(i)
        col := typ.Field(i).Tag.Get("column")
        cols = append(cols, col)
        vals = append(vals, f.Interface())
        if col == "id" {
            pk = f
            pkv = f.Int()
        }
    }

    if pkv < 0 {
        panic("orm: primary key not specified in any entity field tag")
    }

    if pkv == 0 {
        cs := make([]string, 0, n)
        ph := make([]string, 0, n)
        for _,col := range cols {
            cs = append(cs, fmt.Sprintf("`%s`", col))
            ph = append(ph, "?")
        }

        stmt := fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)",
                table, strings.Join(cs, ","), strings.Join(ph, ","))
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
    for _,col := range cols {
        sets = append(sets, fmt.Sprintf("`%s` = ?", col))
    }

    stmt := fmt.Sprintf("UPDATE `%s` SET %s WHERE `id` = %d",
            table, strings.Join(sets, ","), pkv)

    _,e := D.Exec(stmt, vals...)
    if e != nil {
        L.Println(e)
        return false
    }

    return true
}
