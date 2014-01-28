package orm

import (
    "database/sql"
    "errors"
    "fmt"
    "reflect"
    "time"
)

type Rows struct {
    *sql.Rows
    alias   map[string]string
    columns []string
}

func (r *Rows) ScanEntity(entities ...interface{}) error {
    fields := make(map[string]reflect.Value, len(r.columns))
    times := make(map[string]reflect.Value, len(entities)*2)
    values := make([]reflect.Value, 0, len(entities))
    for _, entity := range entities {
        v := reflect.Indirect(reflect.ValueOf(entity))
        if v.Kind() != reflect.Ptr {
            return errors.New("orm: must pass a pointer to Scan")
        }

        v.Set(reflect.New(v.Type().Elem()))
        values = append(values, reflect.Indirect(v))
    }

    for _, val := range values {
        typ := val.Type()
        ali := r.alias[typ.Name()]
        if len(ali) == 0 {
            panic("orm: data not found for entity " + typ.Name())
        }

        for i := 0; i < val.NumField(); i++ {
            f := typ.Field(i)
            if col := f.Tag.Get(ColumnTagName); len(col) > 0 {
                k := fmt.Sprintf("_%s_%s", ali, col)
                if f.Type.Name() == "Time" {
                    times[k] = val.Field(i)
                } else {
                    fields[k] = val.Field(i)
                }
            }
        }
    }

    index := make(map[string]int, len(times))
    row := make([]interface{}, 0, len(r.columns))
    for i, k := range r.columns {
        if f, ok := fields[k]; ok {
            row = append(row, f.Addr().Interface())
        } else {
            var v int64
            row = append(row, &v)
            index[k] = i
        }
    }

    if e := r.Scan(row...); e != nil {
        return e
    }

    for k, i := range index {
        v := row[i].(*int64)
        times[k].Set(reflect.ValueOf(time.Unix(0, *v)))
    }

    return nil
}

func (r *Rows) ScanRow(entities ...interface{}) error {
    defer r.Close()
    if r.Next() {
        return r.ScanEntity(entities...)
    }

    return errors.New("orm: no result found")
}
