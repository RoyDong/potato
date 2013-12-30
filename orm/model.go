package orm

import (
    "reflect"
)


var (
    tables = map[string]string{"User": "user"}

    models = make(map[string]*Model)
)

type Model struct {
    Table string
    Cols []string
    Entity reflect.Type
}

func NewModel(table string, v interface{}) *Model {
    elem := reflect.Indirect(reflect.ValueOf(v)).Type()
    length := elem.NumField()
    cols := make([]string, 0, length)

    for i := 0; i < length; i++ {
        if tag := elem.Field(i).Tag; len(tag) > 0 {
            cols = append(cols, string(tag))
        }
    }

    model := &Model{table, cols, elem}
    tables[elem.Name()] = table
    models[elem.Name()] = model

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
