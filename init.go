package potato

import (
    "os"
    "log"
    "fmt"
    "strings"
    "net"
    "net/http"
)

var (
    AppName = "a potato application"
    Version = "0.1.0"
    Env     = "prod"
    Sock    = ""
    Port    = 37221

    Dir     = &appDir{
        Config:     "config/",
        Controller: "controller/",
        Model:      "model/",
        Template:   "template/",
        Log:        "log/",
    }

    NotFoundRoute = &Route{
        Controller: "Error",
        Action: "NotFound",
    }

    ServerErrorRoute = &Route{
        Controller: "Error",
        Action: "ServerError",
    }

    L *Logger
    R *Router
    S *http.Server
    D *DB
    T *Template
)

type appDir struct {
    Config string
    Controller string
    Model string
    Template string
    Log string
}

func Init() {
    fmt.Println("Starting...")
    //initialize config
    config := new(Tree)
    if e := LoadYaml(&config.data, Dir.Config + "config.yml"); e != nil {
        log.Fatal(e)
    }

    if name, ok := config.String("name"); ok {
        AppName = name
    }

    if env, ok := config.String("env"); ok {
        Env = env
    }

    if v, ok := config.String("session_cookie_name"); ok {
        SessionCookieName = v
    }

    if v, ok := config.String("sock"); ok {
        Sock = v 
    } else if v, ok := config.Int("port"); ok {
        Port = v
    }

    if dir, ok := config.String("log_dir"); ok {
        dir = strings.Trim(dir, "./")
        Dir.Log = dir + "/"
    }
    if dir, ok := config.String("session_dir"); ok {
        dir = strings.Trim(dir, "./")
        SessionDir = dir + "/"
    }

    //error handlers
    if v, ok := config.String("error_handler.not_found.controller"); ok {
        NotFoundRoute.Controller = v
    }
    if v, ok := config.String("error_handler.not_found.action"); ok {
        NotFoundRoute.Action = v
    }
    if v, ok := config.String("error_handler.server_error.controller"); ok {
        ServerErrorRoute.Controller = v
    }
    if v, ok := config.String("error_handler.server_error.action"); ok {
        ServerErrorRoute.Action = v
    }

    //db config
    if v, ok := config.String("sql.type"); ok {
        DBConfig.Type = v
    }
    if v, ok := config.String("sql.host"); ok {
        DBConfig.Host = v
    }
    if v, ok := config.Int("sql.port"); ok {
        DBConfig.Port = v
    }
    if v, ok := config.String("sql.user"); ok {
        DBConfig.User = v
    }
    if v, ok := config.String("sql.pass"); ok {
        DBConfig.Pass = v
    }
    if v, ok := config.String("sql.dbname"); ok {
        DBConfig.DBname = v
    }

    //logger
    file, e := os.OpenFile(Dir.Log + Env + ".log",
            os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
    if e != nil {
        log.Fatal("Error init log file:", e)
    }
    L = NewLogger(file)

    //router
    R = NewRouter()
    R.LoadRouteConfig(Dir.Config + "routes.yml")

    //db
    D = NewDB()

    //template
    T = NewTemplate(Dir.Template)

    SessionStart()
}

func Serve() {
    var l net.Listener
    var e error

    if len(Sock) > 0 {
        os.Remove(Sock)
        l, e = net.Listen("unix", Sock)
        os.Chmod(Sock, os.ModePerm)
    } else {
        l, e = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    }

    if e != nil {
        L.Fatal("failed to start listening", e)
    }

    S = &http.Server{Handler: R}
    L.Println(S.Serve(l))
    l.Close()
}

