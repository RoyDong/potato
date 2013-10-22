package potato

import (
    "os"
    "log"
    "fmt"
    "net/http"
)

var (
    AppName = "a potato application"
    Version = "0.0.1"
    Env     = "prod"

    Host    = "localhost"
    Port    = 80
    Timeout = 30

    Dir     = &AppDirStruct{
        Config:     "./config/",
        Controller: "./controller/",
        Model:      "./model/",
        Static:     "./static/",
        Log:        "./log/",
    }

    NotFoundRoute = &Route{
        Controller: "Error",
        Action: "NotFound",
    }

    ServerErrorRoute = &Route{
        Controller: "Error",
        Action: "ServerError",
    }

    R *Router
    S *http.Server

    Logger *log.Logger
)

type AppDirStruct struct {
    Config string
    Controller string
    Model string
    Static string
    Log string
}

func Init() {
    //initialize config
    var c map[interface{}]interface{}
    if e := LoadYaml(&c, Dir.Config + "config.yml"); e != nil {
        log.Fatal(e)
    }


    if name, ok := c["name"].(string); ok {
        AppName = name
    }

    if env, ok := c["env"].(string); ok {
        Env = env
    }

    if http, ok := c["http"].(map[interface{}]interface{}); ok {
        if host, ok := http["host"].(string); ok {
            Host = host
        }

        if port, ok := http["port"].(int); ok {
            Port = port
        }

        if t, ok := http["timeout"].(int); ok {
            Timeout = t
        }
    }

    if dir, ok := c["static_dir"].(string); ok {
        Dir.Static = dir
    }

    if dir, ok := c["log_dir"].(string); ok {
        Dir.Static = dir
    }

    if eh, ok := c["error_handler"].(map[interface{}]interface{}); ok {
        if nf, ok := eh["not_found"].(map[interface{}]interface{}); ok {
            if v, ok := nf["controller"].(string); ok {
                NotFoundRoute.Controller = v
            }

            if v, ok := nf["action"].(string); ok {
                NotFoundRoute.Action= v
            }
        }

        if se, ok := eh["server_error"].(map[interface{}]interface{}); ok {
            if v, ok := se["controller"].(string); ok {
                ServerErrorRoute.Controller = v
            }

            if v, ok := se["action"].(string); ok {
                ServerErrorRoute.Action = v
            }
        }
    }

    //initialize logger
    file, e := os.OpenFile(Dir.Log + Env + ".log",
            os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
    if e != nil {
        log.Fatal("Error init log file:", e)
    }

    Logger = log.New(file, "", log.LstdFlags)

    //initialize router and load routes config file
    R = NewRouter()
    R.InitConfig(Dir.Config + "routes.yml")

    //initialize server
    S = &http.Server{
        Addr: fmt.Sprintf(":%d", Port),
        Handler: R,
    }

    Logger.Println(fmt.Sprintf("Server ready: %s:%d", Host, Port))
}
