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

    Dir     = &appDir{
        Config:     "config/",
        Controller: "controller/",
        Model:      "model/",
        Template:   "template/",
        Log:        "log/",
    }

    C *Tree
    L *Logger
    R *Router
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
    file, e := os.OpenFile(Dir.Log + Env + ".log",
            os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
    if e != nil {
        log.Fatal("Error init log file:", e)
    }
    L = NewLogger(file)

    //router
    R = NewRouter()
    R.LoadRouteConfig(Dir.Config + "routes.yml")

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

    fmt.Println("work work")
    s := &http.Server{Handler: R}
    L.Println(s.Serve(l))
    l.Close()
}