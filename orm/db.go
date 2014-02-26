package orm

import (
    "github.com/roydong/potato/lib"
    "database/sql"
    "log"
)

var (
    DefaultDB = "default"
)

var dbpool = make(map[string]*DB)

func Init(conf *lib.Tree) {
    for name, c := range conf.Branches() {
        db, e := newDB(c)
        if e != nil {
            log.Fatal("orm: ", e)
        }
        dbpool[name] = db
    }
}

type DB struct {
    db *sql.DB
}

func newDB(conf *lib.Tree) (*DB, error) {
    dsn, _ := conf.String("dsn")
    typ, _ := conf.String("type")
    db, e := sql.Open(typ, dsn)
    if e != nil {
        return nil, e
    }
    e = db.Ping()
    if e != nil {
        return nil, e
    }
    return &DB{db}, nil
}

func GetDB(name string) *DB {
    if name == "" {
        name = DefaultDB
    }
    db, has := dbpool[name]
    if !has {
        panic("orm: db not found")
    }
    return db
}

func NewStmt(name string) *Stmt {
    return GetDB(name).Stmt()
}

func NewTx(name string) (*Tx, error) {
    tx, e := GetDB(name).Begin()
    return tx, e
}

func SqlDB(name string) *sql.DB {
    return GetDB(name).db
}

func (d *DB) Stmt() *Stmt {
    s := &Stmt{}
    s.init()
    s.db = d.db
    return s
}

func (d *DB) Save(entity interface{}) bool {
    return save(entity, d.db, nil)
}

func (d *DB) Begin() (*Tx, error) {
    tx, e := d.db.Begin()
    if e != nil {
        log.Println("orm: ", e)
        return nil, e
    }
    return &Tx{tx}, nil
}

func (d *DB) SqlDB() *sql.DB {
    return d.db
}

func (d *DB) Close() {
    d.db.Close()
}

type Tx struct {
    tx *sql.Tx
}

func (tx *Tx) Stmt() *Stmt {
    s := &Stmt{}
    s.init()
    s.tx = tx.tx
    return s
}

func (tx *Tx) Save(entity interface{}) bool {
    return save(entity, nil, tx.tx)
}

func (tx *Tx) Commit() error {
    return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
    return tx.tx.Rollback()
}
