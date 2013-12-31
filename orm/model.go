package orm

import (
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

func (m *Model) ColumnIndex(col string) int {
    for i, v := range m.Cols {
        if v == col {
            return i
        }
    }

    return -1
}
