package potato

import (
    "os"
    "log"
    "fmt"
    "net"
    "net/http"
    "strings"
)

var (
    AppName = "a potato application"
    Version = "0.1.0"
    Env     = "prod"
    SockFile= ""
    Port    = 37221
    WSPort  = 81

    Dir     = &appDir{
        Config:     "config/",
        Controller: "controller/",
        Model:      "model/",
        Template:   "template/",
        Log:        "log/",
    }

    L *Logger
    R *Router
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
    fmt.Print("Starting...")
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

    if v, ok := config.String("sock_file"); ok {
        SockFile = v
    } else if v, ok := config.Int("port"); ok {
        Port = v
    }

    if v, ok := config.Int("wsport"); ok {
        WSPort = v
    }



    if dir, ok := config.String("log_dir"); ok {
        dir = strings.Trim(dir, "./")
        Dir.Log = dir + "/"
    }
    if dir, ok := config.String("session_dir"); ok {
        dir = strings.Trim(dir, "./")
        SessionDir = dir + "/"
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
    var e error
    var l net.Listener

    if len(SockFile) > 0 {
        os.Remove(SockFile)
        l, e = net.Listen("unix", SockFile)
        os.Chmod(SockFile, os.ModePerm)
    } else {
        l, e = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    }

    if e != nil {
        L.Fatal("failed to start listening", e)
    }

    hs := &http.Server{Handler: R}
    go hs.Serve(l)

    ws := &http.Server{Addr: fmt.Sprintf(":%d", WSPort), Handler: R}
    go ws.ListenAndServe()

    fmt.Println("done")
    //forever
    select{}
    l.Close()
}
