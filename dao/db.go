package dao

import (
    "github.com/roydong/potato/lib"
    "database/sql"
    "sync"
    "log"
)

var (
    DefaultDB = "default"
    Conf *lib.Tree
)

var dbpool map[string]*DB

func Init(conf *lib.Tree) {
    Conf = conf
    dbpool = make(map[string]*DB, len(conf.Branches()))
}

type DB struct {
    db   *sql.DB
    name string
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

//终于碰到读写锁的一个应用场景了
var dbpoolLocker = &sync.RWMutex{}

func GetDB(name string) *DB {
    dbpoolLocker.RLock()
    defer dbpoolLocker.RUnlock()
    if name == "" {
        name = DefaultDB
    }
    db := dbpool[name]
    if db == nil {
        dbpoolLocker.Lock()
        if conf := Conf.Tree(name); conf != nil {
            var e error
            if db, e = newDB(conf); e != nil {
                dbpool[name] = db
            }
        }
        dbpoolLocker.Unlock()
    }
    return db
}


func (db *DB) Query(stmt string, vals ...interface{}) *Rows {
    colMap := parseSelect(stmt)
    rows, e = db.db.Query(stmt, vals...)
    return &Rows{rows, colMap}, e
}

func (db *DB) Exec(stmt string, vals ...interface{}) *sql.Result {

}

func (db *DB) Save() bool {

}

func (db *DB) Begin() (*Tx, error) {

}



