package orm

import (
    "github.com/roydong/potato/lib"
    "database/sql"
    "os"
    "fmt"
    "log"
)

var (
    Logger        = log.New(os.Stdout, "", log.LstdFlags)
    DefaultDBname = "default"
)

type DBConf struct {
    Type string
    User string
    Pass string
    Host string
    Name string
    Port int
    Maxc int
}

var dbpool = make(map[string]*DB)

func Init(file string, logger *log.Logger) {
    Logger = logger
    var conf map[string]*DBConf
    e := lib.LoadYaml(&conf, file)
    if e != nil {
        Logger.Fatal("orm: ", e)
    }
    for name, c := range conf {
        db, e := newDB(c)
        if e != nil {
            Logger.Fatal("orm: ", e)
        }
        dbpool[name] = db
    }
}

type DB struct {
    db   *sql.DB
    conf *DBConf
}

func newDB(conf *DBConf) (*DB, error) {
    var db *sql.DB
    var e error
    dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s",
        conf.User, conf.Pass, conf.Host, conf.Port, conf.Name)
    if db, e = sql.Open(conf.Type, dsn); e != nil {
        return nil, e
    }
    if e = db.Ping(); e != nil {
        return nil, e
    }
    if conf.Maxc > 0 {
        db.SetMaxOpenConns(conf.Maxc)
    }
    return &DB{db, conf}, nil
}

func GetDB(name string) *DB {
    if name == "" {
        name = DefaultDBname
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

func (d *DB) Conf() *DBConf {
    return d.conf
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
        Logger.Println("orm: ", e)
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
