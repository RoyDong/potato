package orm

import (
    "github.com/roydong/potato/lib"
    "database/sql"
    "os"
    "fmt"
    "log"
)

var (
    DB     *sql.DB
    Conf   *lib.Tree
    Logger = log.New(os.Stdout, "", log.LstdFlags)
)

func Init(conf *lib.Tree, logger *log.Logger) {
    Logger = logger
    Conf = conf
    DB = NewDB()
}

func NewDB() *sql.DB {
    dbc := Conf.Tree("db")
    if dbc == nil {
        Logger.Fatal("orm: db config not found")
    }

    user, _ := dbc.String("user")
    pass, _ := dbc.String("pass")
    host, _ := dbc.String("host")
    port, _ := dbc.Int("port")
    name, _ := dbc.String("dbname")
    maxc, _ := dbc.Int("max_conn")

    var db *sql.DB
    var e error
    dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s", user, pass, host, port, name)
    typ, _ := dbc.String("type")
    if db, e = sql.Open(typ, dsn); e != nil {
        log.Fatal("orm:", e)
    }

    if maxc > 0 {
        db.SetMaxOpenConns(maxc)
    }

    if e = db.Ping(); e != nil {
        log.Fatal("orm:", e)
    }

    return db
}
