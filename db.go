package potato

import (
    "github.com/roydong/potato/lib"
    "database/sql"
    "strings"
    "reflect"
    "fmt"
    "errors"
    "sync"
)

var (
    DefaultDB = "default"

    ColumnTagName = "column"

    dbpool = make(map[string]*sql.DB)

    dbpoolLocker = &sync.RWMutex{}
)

func newDB(conf *lib.Tree) (*sql.DB, error) {
    dsn, _ := conf.String("dsn")
    typ, _ := conf.String("type")
    db, err := sql.Open(typ, dsn)
    if err != nil {
        return nil, err
    }
    err = db.Ping()
    if err != nil {
        return nil, err
    }
    return db, nil
}

func GetDB(name string) *sql.DB {
    dbpoolLocker.RLock()
    if name == "" {
        name = DefaultDB
    }
    db := dbpool[name]
    dbpoolLocker.RUnlock()
    if db == nil {
        dbpoolLocker.Lock()
        defer dbpoolLocker.Unlock()
        if conf := Conf.Tree("db." + name); conf != nil {
            var err error
            if db, err = newDB(conf); err != nil {
                dbpool[name] = db
            }
        }
    }
    if db == nil {
        panic("potato: db " + name + " not found")
    }
    return db
}

type sameNameField struct {
    fields []reflect.Value
    index int
}

func (s *sameNameField) val() interface{} {
    if s.index < len(s.fields) {
        v := s.fields[s.index]
        s.index++
        return v.Addr().Interface()
    }
    return nil
}

func Scan(r *sql.Rows, dest ...interface{}) error {
    cols, err := r.Columns()
    if err != nil {
        return err
    }

    daos := make([]reflect.Value, 0, len(dest))
    vals := make([]interface{}, 0)
    for _, val := range dest {
        v := reflect.ValueOf(val)
        if v.Kind() != reflect.Ptr {
            return errors.New("potato: must pass pointers to Scan")
        }
        vv := reflect.Indirect(v)
        if vv.Kind() != reflect.Ptr {
            vals = append(vals, val)
        } else {
            elem := vv.Type().Elem()
            vv.Set(reflect.New(elem))
            daos = append(daos, reflect.Indirect(vv))
        }
    }

    fields := make(map[string]*sameNameField, len(cols))
    for _, dao := range daos {
        typ := dao.Type()
        for i := 0; i < dao.NumField(); i++ {
            f := typ.Field(i)
            if col := f.Tag.Get(ColumnTagName); len(col) > 0 {
                snf, has := fields[col]
                if !has {
                    snf = &sameNameField{make([]reflect.Value, 0, 1), 0}
                    fields[col] = snf
                }
                snf.fields = append(snf.fields, dao.Field(i))
            }
        }
    }

    row := make([]interface{}, 0, len(cols))
    for _, k := range cols[:len(cols) - len(vals)] {
        if snf, has := fields[k]; has {
            row = append(row, snf.val())
        } else {
            row = append(row, nil)
        }
    }
    for _, val := range vals {
        row = append(row, val)
    }
    return r.Scan(row...)
}

func ScanRow(r *sql.Rows, dest ...interface{}) error {
    defer r.Close()
    if r.Next() {
        return Scan(r, dest...)
    }
    return errors.New("potato: no result")
}

type DBer interface {
    Exec(query string, args ...interface{}) (sql.Result, error)
    Prepare(query string) (*sql.Stmt, error)
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
}

func Save(dao interface{}, tbl string, db DBer) error {
    return save(dao, false, tbl, db)
}

func Insert(dao interface{}, tbl string, db DBer) error {
    return save(dao, true, tbl, db)
}

func save(dao interface{}, insert bool, tbl string, db DBer) error {
    val := reflect.Indirect(reflect.ValueOf(dao))
    typ := val.Type()
    cols := make([]string, 0, typ.NumField())
    vals := make([]interface{}, 0, typ.NumField())
    num := typ.NumField()
    pkv := int64(-1)
    var pk reflect.Value
    for i := 0; i < num; i++ {
        f := val.Field(i)
        col := typ.Field(i).Tag.Get(ColumnTagName)
        if len(col) > 0 {
            ifc := f.Interface()
            cols = append(cols, col)
            vals = append(vals, ifc)
            if col == "id" {
                pk = f
                pkv = f.Int()
            }
        }
    }

    if insert || pkv <= 0 {
        cs := make([]string, 0, num)
        ph := make([]string, 0, num)
        for _, col := range cols {
            cs = append(cs, fmt.Sprintf("`%s`", col))
            ph = append(ph, "?")
        }

        stmt := fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)",
            tbl, strings.Join(cs, ","), strings.Join(ph, ","))

        result, err := db.Exec(stmt, vals...)
        if err != nil {
            return err
        }

        id, err := result.LastInsertId()
        if err != nil {
            return err
        }

        pk.SetInt(id)
        return nil
    }

    if pkv <= 0 {
        panic("orm: primary key(id) not specified")
    }

    sets := make([]string, 0, num)
    for _, col := range cols {
        sets = append(sets, fmt.Sprintf("`%s` = ?", col))
    }

    stmt := fmt.Sprintf("UPDATE `%s` SET %s WHERE `id` = %d", tbl, strings.Join(sets, ","), pkv)
    _, err := db.Exec(stmt, vals...)
    return err
}
