package orm

import (
    "database/sql"
    "fmt"
    "log"
)

var (
    DB     *sql.DB
    Logger *log.Logger
    Conf   *Config
)

type Config struct {
    Type    string
    Host    string
    Port    int
    User    string
    Pass    string
    DBname  string
    MaxConn int
}

func Init(conf *Config, logger *log.Logger) {
    Logger = logger
    Conf = conf
    DB = NewDB()
}

func NewDB() *sql.DB {
    var db *sql.DB
    var e error

    dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s",
        Conf.User, Conf.Pass, Conf.Host,
        Conf.Port, Conf.DBname)

    if db, e = sql.Open(Conf.Type, dsn); e != nil {
        log.Fatal(e)
    }

    if Conf.MaxConn > 0 {
        db.SetMaxOpenConns(Conf.MaxConn)
    }

    if e = db.Ping(); e != nil {
        log.Fatal(e)
    }

    return db
}
