package potato

import (
    "github.com/roydong/potato/lib"
    "log"
    "os"
    "strings"
)

var (
    AppName  string
    Version  string
    SockFile string
    Pwd      string
    Daemon   bool
    Conf     *lib.Tree
    ConfDir  = "config/"
    TplDir   = "template/"
    Env      = "prod"
    Port     = 37221
)

func initConfig() {
    confile := "config.yml"
    for i, arg := range os.Args {
        if arg == "-d" {
            Daemon = true
        } else if arg == "-c" && i+1 < len(os.Args) {
            confile = os.Args[i+1]
            if i := strings.LastIndex(confile, "/"); i >= 0 {
                Pwd = confile[:i+1]
            }
        }
    }

    Conf = lib.NewTree()
    if err := Conf.LoadYaml(confile, false); err != nil {
        log.Fatal("potato: ", err)
    }

    if v, ok := Conf.String("pwd"); ok {
        Pwd = v
    }
    if Pwd != "" {
        os.Chdir(Pwd)
    }

    if name, ok := Conf.String("name"); ok {
        AppName = name
    }

    if env, ok := Conf.String("env"); ok {
        Env = env
    }

    if v, ok := Conf.String("session_cookie_name"); ok {
        SessionCookieName = v
    }

    if v, ok := Conf.String("sock_file"); ok {
        SockFile = v
    }
    if v, ok := Conf.Int("port"); ok {
        Port = v
    }

    if v, ok := Conf.String("template_ext"); ok {
        TemplateExt = v
    }

    if dir, ok := Conf.String("template_dir"); ok {
        if dir[len(dir)-1] != '/' {
            dir = dir + "/"
        }
        TplDir = dir
    }
    if dir, ok := Conf.String("config_dir"); ok {
        if dir[len(dir)-1] != '/' {
            dir = dir + "/"
        }
        ConfDir = dir
    }

    if v, ok := Conf.String("default_db"); ok {
        DefaultDB = v
    }
}

func Init() {
    event.Trigger("before_init")
    initConfig()
    event.Trigger("after_init")
}
