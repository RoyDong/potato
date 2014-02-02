package potato

import (
    "github.com/roydong/potato/orm"
    "log"
    "os"
    "strings"
)

var (
    AppName  = "a potato application"
    Version  = "0.1.0"
    Env      = "prod"
    SockFile = ""
    Port     = 37221

    Dir = &appDir{
        Config:   "config/",
        Template: "template/",
        Log:      "log/",
    }

    E   = NewEvent()
    T   = NewTemplate(Dir.Template)
    C   *Tree
    L   *log.Logger
)

type appDir struct {
    Config   string
    Template string
    Log      string
}

func Init() {
    E.TriggerEvent("frame_init_start")

    //load config
    var data map[interface{}]interface{}
    if e := LoadYaml(&data, Dir.Config+"config.yml"); e != nil {
        log.Fatal(e)
    }
    C = NewTree(data)

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

    if v, ok := C.String("default_layout"); ok {
        DefaultLayout = v
    }

    if v, ok := C.String("template_ext"); ok {
        TemplateExt = v
    }

    if dir, ok := C.String("log_dir"); ok {
        dir = strings.Trim(dir, "./")
        Dir.Log = dir + "/"
    }

    //logger
    var logio *os.File
    if Env == "dev" {
        logio = os.Stdout
    } else {
        var e error
        logio, e = os.OpenFile(Dir.Log+Env+".log",
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if e != nil {
            log.Fatal("Error init log file:", e)
        }
    }
    L = log.New(logio, "", log.LstdFlags)

    //init orm
    E.TriggerEvent("orm_init_start")
    initOrm()
    E.TriggerEvent("orm_init_done")

    //start session check
    go sessionExpire()

    E.TriggerEvent("frame_init_done")
}

func initOrm() {
    if c, ok := C.Tree("sql"); ok {
        dbc := &orm.Config{
            Type:   "mysql",
            Host:   "localhost",
            Port:   3306,
            User:   "root",
            Pass:   "",
            DBname: "",
        }

        if v, ok := c.String("type"); ok {
            dbc.Type = v
        }
        if v, ok := c.String("host"); ok {
            dbc.Host = v
        }
        if v, ok := c.Int("port"); ok {
            dbc.Port = v
        }
        if v, ok := c.String("user"); ok {
            dbc.User = v
        }
        if v, ok := c.String("pass"); ok {
            dbc.Pass = v
        }
        if v, ok := c.String("dbname"); ok {
            dbc.DBname = v
        }
        if v, ok := c.Int("max_conn"); ok {
            dbc.MaxConn = v
        }

        orm.Init(dbc, L)
    }
}
