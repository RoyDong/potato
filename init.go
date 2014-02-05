package potato

import (
    "github.com/roydong/potato/orm"
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
    Conf     *Tree
    Logger   *log.Logger
    ConfDir  = "config/"
    TplDir   = "template/"
    LogDir   = "log/"
    Env      = "prod"
    Port     = 37221
)

func Init() {
    event.Trigger("before_init")
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

    //load config
    var data map[interface{}]interface{}
    if e := LoadYaml(&data, confile); e != nil {
        log.Fatal("potato: config file", e)
    }
    Conf = NewTree(data)

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

    if v, ok := Conf.String("default_layout"); ok {
        DefaultLayout = v
    }

    if v, ok := Conf.String("template_ext"); ok {
        TemplateExt = v
    }

    if dir, ok := Conf.String("log_dir"); ok {
        if dir[len(dir)-1] != '/' {
            dir = dir + "/"
        }
        LogDir = dir
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

    if v, ok := Conf.String("pwd"); ok {
        Pwd = v
    }

    if Pwd != "" {
        os.Chdir(Pwd)
    }

    //logger
    var logio *os.File
    if Daemon {
        var e error
        logio, e = os.OpenFile(LogDir+Env+".log",
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if e != nil {
            log.Fatal("potato: log file", e)
        }
    } else {
        logio = os.Stdout
    }
    Logger = log.New(logio, "", log.LstdFlags)
    initOrm()
    event.Trigger("after_init")
}

func initOrm() {
    event.Trigger("before_orm_init")
    if c, ok := Conf.Tree("sql"); ok {
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

        orm.Init(dbc, Logger)
    }
    event.Trigger("after_orm_init")
}
