package potato

import (
    "log"
    "fmt"
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


type DB struct {
    *sql.DB
}

func InitDB() *DB {
    var db *sql.DB
    var e error
    dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s", DBConfig.User,
            DBConfig.Pass, DBConfig.Host, DBConfig.Port, DBConfig.DBname)
    if db, e = sql.Open(DBConfig.Type, dsn); e != nil {
        log.Fatal(e)
    }

    if e = db.Ping(); e != nil {
        log.Fatal(e)
    }

    return &DB{DB: db}
}

func (d *DB) Insert(stmt string) int64 {
    result, e := d.Exec(stmt)
    if e != nil {
        L.Println(e)
        return 0
    }

    id, e := result.LastInsertId()
    if e!= nil {
        L.Println(e)
        return 0
    }

    return id
}
