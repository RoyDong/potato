package db

import (
    "log"
    "fmt"
    "github.com/roydong/potato"
    "database/sql"
    _"github.com/go-sql-driver/mysql"
)


var (
    DBConfig = &dbConfig{
        Type: "mysql",
        Host: "localhost",
        Port: 3306,
        User: "root",
        Pass: "",
        DBname: "",
    }
)


type dbConfig struct {
    Type string
    Host string
    Port int
    User string
    Pass string
    DBname string
}

func SetConfig(c *potato.Tree) {
    if v, ok := c.String("type"); ok {
        DBConfig.Type = v
    }
    if v, ok := c.String("host"); ok {
        DBConfig.Host = v
    }
    if v, ok := c.Int("port"); ok {
        DBConfig.Port = v
    }
    if v, ok := c.String("user"); ok {
        DBConfig.User = v
    }
    if v, ok := c.String("pass"); ok {
        DBConfig.Pass = v
    }
    if v, ok := c.String("dbname"); ok {
        DBConfig.DBname = v
    }
}

type DB struct {
    *sql.DB
}

func NewDB() *DB {
    var db *sql.DB
    var e error
    dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s", DBConfig.User,
            DBConfig.Pass, DBConfig.Host, DBConfig.Port, DBConfig.DBname)
    if db, e = sql.Open(DBConfig.Type, dsn); e != nil {
        log.Fatal(e)
    }

    return &DB{DB: db}
}