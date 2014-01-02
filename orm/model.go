package orm

import (
    "fmt"
    "strings"
    "reflect"
)


var (
    tables = make(map[string]string, 20)
    models = make(map[string]*Model, 20)
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

func Insert(entity interface{}) bool {
    value := reflect.Indirect(reflect.ValueOf(entity))
    t := value.Type()
    table, ok := tables[t.Name()]
    if !ok {
        panic("orm: error entity var while using orm.Insert")
    }

    var pk reflect.Value
    var hasPk bool
    l := value.NumField()
    c := make([]string, 0, l)
    h := make([]string, 0, l)
    v := make([]interface{}, 0, l)
    for i := 0; i < l; i++ {
        n := t.Field(i).Tag.Get("column")
        val := value.Field(i)
        if n == "id" {
            pk = val
            hasPk = true
        }

        c = append(c, fmt.Sprintf("`%s`", n))
        h = append(h, "?")
        v = append(v, val.Interface())
    }

    if !hasPk {
        panic("orm: primary key not specified for entity")
    }

    stmt := fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)",
            table, strings.Join(c, ","), strings.Join(h, ","))

    result, e := D.Exec(stmt, v...)
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

func Update(entity interface{}) bool {
    value := reflect.Indirect(reflect.ValueOf(entity))
    t := value.Type()
    table, ok := tables[t.Name()]
    if !ok {
        panic("orm: error entity var while using orm.Update")
    }

    var pkv int64
    var pkn string
    l := value.NumField()
    c := make([]string, 0, l)
    v := make([]interface{}, 0, l)
    for i := 0; i < l; i++ {
        n := t.Field(i).Tag.Get("column")
        val := value.Field(i)
        if n == "id" {
            pkv = val.Int()
            pkn = n
        }

        c = append(c, fmt.Sprintf("`%s` = ?", n))
        v = append(v, val.Interface())
    }

    if pkv <= 0 {
        panic("orm: primary key not specified for entity")
    }

    stmt := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = %d",
            table, strings.Join(c, ","), pkn, pkv)

    _,e := D.Exec(stmt, v...)
    if e != nil {
        L.Println(e)
        return false
    }

    return true
}
