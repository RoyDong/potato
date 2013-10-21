package potato

import (
    "os"
    "log"
    "fmt"
    "net/http"
)

var (
    Env = "prod"

    Version = "0.0.1"

    Dir = &AppDirStruct{
        Config:     "./config/",
        Controller: "./controller/",
        Model:      "./model/",
        Static:     "./static/",
        Log:        "./log/",
    }

    R *Router
    S *http.Server

    Config *ConfigFile
    Logger *log.Logger
)

type AppDirStruct struct {
    Config string
    Controller string
    Model string
    Static string
    Log string
}

type ConfigFile struct {
    Name string `yaml:"name"`
    Env string `yaml:"env"`


    LogDir string `yaml:"log_dir"`
    StaticDir string `yaml:"static_dir"`
    AllowDir bool `yaml:"allow_dir"`

    Http struct {
        Host string `yaml:"host"`
        Port int `yaml:"port"`
        Timeout int `yaml:"timeout"`
    } `yaml:"http"`
}

func Init() {
    //load config
    if e := LoadYaml(&Config, Dir.Config + "config.yml"); e != nil {
        log.Fatal(e)
    }

    if len(Config.Env) > 0 {
        Env = Config.Env
    }
    if len(Config.StaticDir) > 0 {
        Dir.Static = Config.StaticDir
    }
    if len(Config.LogDir) > 0 {
        Dir.Log = Config.LogDir
    }

    //init logger
    file, e := os.OpenFile(Dir.Log + Env + ".log",
            os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
    if e != nil {
        log.Fatal("Error init log file:", e)
    }

    Logger = log.New(file, "", log.LstdFlags)

    //init router and load routes config file
    R = NewRouter()
    R.InitConfig(Dir.Config + "routes.yml")

    //create server
    S = &http.Server{
        Addr: fmt.Sprintf(":%d", Config.Http.Port),
        Handler: R,
    }

    Logger.Println("Server started")
}

/*
func Log(v ...interface{}) {
    if Env == "dev" {
        log.Println(v)
    }

    Logger.Println(v)
}

func Fatal(v ...interface{}) {
    if Env == "dev" {
        log.Println(v)
    }

    Logger.Fatal(v)
}
*/
