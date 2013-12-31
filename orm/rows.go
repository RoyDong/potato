package orm

import (
    "fmt"
    "time"
    "reflect"
    "database/sql"
)

type Rows struct {
    *sql.Rows
    alias map[string]string
    columns []string
}

func (r *Rows) Scan(dest ...interface{}) error {
    fields := make(map[string]reflect.Value, len(r.columns))
    times := make(map[string]reflect.Value, len(dest) * 2)

    values := make([]reflect.Value, 0, len(dest))
    for _,d := range dest {
        v := reflect.Indirect(reflect.ValueOf(d))
        t := v.Type().Elem()
        v.Set(reflect.New(t))
        values = append(values, reflect.Indirect(v))
    }

    for _,v := range values {
        for i := 0; i < v.NumField(); i++ {
            vt := v.Type()
            ft := vt.Field(i)
            if col := ft.Tag.Get("column"); len(col) > 0 {
                k := fmt.Sprintf("_%s_%s", r.alias[vt.Name()], col)
                if ft.Type.Name() == "Time" {
                    times[k] = v.Field(i)
                } else {
                    fields[k] = v.Field(i)
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

    if e := r.Rows.Scan(row...); e != nil {
        return e
    }

    for k, i := range index {
        v := row[i].(*int64)
        times[k].Set(reflect.ValueOf(time.Unix(0, *v)))
    }

    return nil
}
