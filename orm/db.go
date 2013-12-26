package orm

import (
    "log"
    "database/sql"
)


var (
    D *sql.DB
    L *log.Logger
)

func Init(d *sql.DB, l *log.Logger) {
    L = l
    D = d
}