package orm

import (
    "database/sql"
    "fmt"
    "github.com/roydong/potato"
    "log"
)

var (
    D   *sql.DB
    L   *log.Logger

    C   = &Config{
        Type:   "mysql",
        Host:   "localhost",
        Port:   3306,
        User:   "root",
        Pass:   "",
        DBname: "",
    }
)

type Config struct {
    Type   string
    Host   string
    Port   int
    User   string
    Pass   string
    DBname string
}

func InitDefault() {
    if potato.L == nil {
        panic("orm: potato not init")
    }

    if c, ok := potato.C.Tree("sql"); ok {
        if v, ok := c.String("type"); ok {
            C.Type = v
        }
        if v, ok := c.String("host"); ok {
            C.Host = v
        }
        if v, ok := c.Int("port"); ok {
            C.Port = v
        }
        if v, ok := c.String("user"); ok {
            C.User = v
        }
        if v, ok := c.String("pass"); ok {
            C.Pass = v
        }
        if v, ok := c.String("dbname"); ok {
            C.DBname = v
        }
    } else {
        panic("orm: sql db config not found")
    }

    L = potato.L
    D = NewDB()
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
