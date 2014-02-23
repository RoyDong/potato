package potato

import (
    "github.com/roydong/potato/lib"
    "github.com/roydong/potato/orm"
    "log"
    "os"
    "strings"
    "runtime"
    "syscall"
)

var (
    AppName  string
    Version  string
    SockFile string
    Pwd      string
    Daemon   bool
    Conf     *lib.Tree
    Logger   = log.New(os.Stdout, "", log.LstdFlags)
    ConfDir  = "config/"
    TplDir   = "template/"
    LogDir   = "log/"
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
    if e := Conf.LoadYaml(confile, false); e != nil {
        log.Fatal("potato: ", e)
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

    if v, ok := Conf.String("default_dbname"); ok {
        orm.DefaultDBname = v
    }
    var dbfile string
    if v, ok := Conf.String("db_config"); ok {
        dbfile = v
    } else {
        dbfile = "database.yml"
    }
    orm.Init(ConfDir + dbfile, Logger)
}

func fork() {
    darwin := runtime.GOOS == "darwin"
    if syscall.Getppid() == 1 {
        return
    }

    ret, ret2, err := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
    if err != 0 || ret2 < 0 {
        Logger.Fatal("potato: error forking process")
    }
    if darwin && ret2 == 1 {
        ret = 0
    }
    if ret > 0 {
        os.Exit(0)
    }

    syscall.Umask(0)
    sret, errno := syscall.Setsid()
    if errno != nil {
        Logger.Printf("potato: syscall.Setsid errno: %d", errno)
    }
    if sret < 0 {
        Logger.Fatal("potato: error setting sid")
    }
}

func Init() {
    event.Trigger("before_init")
    initConfig()
    if Daemon {
        fork()
    }
    event.Trigger("after_init")
}
