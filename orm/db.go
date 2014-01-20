package orm

import (
    "database/sql"
    "fmt"
    "log"
)

var (
    D *sql.DB
    L *log.Logger
    C *Config
)

type Config struct {
    Type   string
    Host   string
    Port   int
    User   string
    Pass   string
    DBname string
}

func Init(c *Config, l *log.Logger) {
    L = l
    C = c
    D = NewDB()
}

func NewDB() *sql.DB {
    var db *sql.DB
    var e error

    dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s",
        C.User, C.Pass, C.Host, C.Port, C.DBname)

    if db, e = sql.Open(C.Type, dsn); e != nil {
        log.Fatal(e)
    }

    if e = db.Ping(); e != nil {
        log.Fatal(e)
    }

    return db
}
