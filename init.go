package potato

import (
    "os"
    "log"
    "fmt"
    "net"
    "strings"
    "net/http"
    "database/sql"
    "github.com/roydong/potato/orm"
)

var (
    AppName  = "a potato application"
    Version  = "0.1.0"
    Env      = "prod"
    SockFile = ""
    Port     = 37221

    Dir      = &appDir{
        Config:     "config/",
        Controller: "controller/",
        Model:      "model/",
        Template:   "template/",
        Log:        "log/",
    }

    DBConfig = &dbConfig{
        Type: "mysql",
        Host: "localhost",
        Port: 3306,
        User: "root",
        Pass: "",
        DBname: "",
    }

    C *Tree
    L *log.Logger
    R *Router
    T *Template
    D *sql.DB
)

type appDir struct {
    Config string
    Controller string
    Model string
    Template string
    Log string
}

type dbConfig struct {
    Type string
    Host string
    Port int
    User string
    Pass string
    DBname string
}

func Init() {
    //initialize config
    C = new(Tree)
    if e := LoadYaml(&C.data, Dir.Config + "config.yml"); e != nil {
        log.Fatal(e)
    }

    if name, ok := C.String("name"); ok {
        AppName = name
    }

    if env, ok := C.String("env"); ok {
        Env = env
    }

    if v, ok := C.String("session_cookie_name"); ok {
        SessionCookieName = v
    }

    if v, ok := C.String("sock_file"); ok {
        SockFile = v
    }
    if v, ok := C.Int("port"); ok {
        Port = v
    }

    if dir, ok := C.String("log_dir"); ok {
        dir = strings.Trim(dir, "./")
        Dir.Log = dir + "/"
    }
    if dir, ok := C.String("session_dir"); ok {
        dir = strings.Trim(dir, "./")
        SessionDir = dir + "/"
    }

    //logger
    var logio *os.File
    if Env == "dev" {
        logio = os.Stdout
    } else {
        var e error
        logio, e = os.OpenFile(Dir.Log + Env + ".log",
                os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
        if e != nil {
            log.Fatal("Error init log file:", e)
        }
    }

    L = log.New(logio, "", log.LstdFlags)

    //router
    R = NewRouter()
    R.LoadRouteConfig(Dir.Config + "routes.yml")

    //template
    T = NewTemplate(Dir.Template)

    if v, ok := C.Tree("sql"); ok {
        InitDB(v)
        orm.Init(D, L)
    }

    SessionStart()
}

func InitDB(c *Tree) {
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

    D = NewDB()
}

func NewDB() *sql.DB {
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

    return db
}

func Serve() {
    var e error
    var lsn net.Listener

    if len(SockFile) > 0 {
        os.Remove(SockFile)
        lsn, e = net.Listen("unix", SockFile)
        if e != nil {
            L.Println("fail to open socket file", e)
        } else {
            os.Chmod(SockFile, os.ModePerm)
        }
    }

    if lsn == nil {
        lsn, e = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    }

    if e != nil {
        L.Fatal(e)
    }

    fmt.Println("work work")
    s := &http.Server{Handler: R}
    L.Println(s.Serve(lsn))
    lsn.Close()
}
