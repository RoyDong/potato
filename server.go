package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "fmt"
    "net"
    "net/http"
    "os"
    "runtime"
    "strings"
    "syscall"
)

/**
 * events:
 *  before_init
 *  after_init
 *
 *  before_orm_init
 *  after_orm_init
 *
 *  request, request started, just before routing
 *
 *  before_action, after routing, just before running action
 *
 *  response, just before response
 */
var event = NewEvent()

func AddHandler(name string, handler EventHandler) {
    event.AddHandler(name, handler)
}

var route = &Route{}

func SetAction(action Action, patterns ...string) {
    for _, pattern := range patterns {
        route.Set(pattern, action)
    }
}

var NotfoundAction = func(r *Request, p *Response) error {
    p.WriteHeader(404)
    p.Write([]byte("page not found"))
    return nil
}

var ErrorAction = func(r *Request, p *Response) error {
    msg := "unknown"
    e, ok := r.Bag.Get("error").(error)
    if ok {
        msg = e.Error()
    }
    p.WriteHeader(500)
    p.Write([]byte("error: " + msg))
    Logger.Println(e)
    return nil
}

type handler struct {
    ws.Server
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    request := NewRequest(r)
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        if conn := h.Conn(w, r); conn != nil {
            request.WSConn = conn
            defer conn.Close()
        }
    }
    response := NewResponse(w)
    InitSession(request, response)
    event.Trigger("request", request, response)

    var rt *Route
    rt, request.params = route.Parse(r.URL.Path)
    event.Trigger("before_action", request, response)
    if rt == nil {
        NotfoundAction(request, response)
    } else {
        if e := rt.action(request, response); e != nil {
            request.Bag.Set("error", e, true)
            ErrorAction(request, response)
        }
    }
    event.Trigger("response", request, response)
}

func listener() net.Listener {
    var e error
    var lsr net.Listener
    if len(SockFile) > 0 {
        os.Remove(SockFile)
        lsr, e = net.Listen("unix", SockFile)
        if e == nil {
            os.Chmod(SockFile, os.ModePerm)
            return lsr
        }
    }
    lsr, e = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    if e != nil {
        Logger.Fatal(e)
    }
    return lsr
}

func fork() {
    darwin := runtime.GOOS == "darwin"
    if syscall.Getppid() == 1 {
        return
    }

    ret, ret2, err := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
    if err != 0 || ret2 < 0 {
        Logger.Fatal("orm: error forking process")
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
        Logger.Printf("orm: syscall.Setsid errno: %d", errno)
    }
    if sret < 0 {
        Logger.Fatal("orm: error setting sid")
    }

    println("work work")
    f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
    if e == nil {
        fd := int(f.Fd())
        syscall.Dup2(fd, int(os.Stdin.Fd()))
        syscall.Dup2(fd, int(os.Stdout.Fd()))
        syscall.Dup2(fd, int(os.Stderr.Fd()))
    }
}

func Serve() {
    tpl = NewTemplate(TplDir)
    tpl.loadTemplateFiles(tpl.dir)
    srv := &http.Server{Handler: &handler{ws.Server{}}}
    lsr := listener()
    defer lsr.Close()
    go sessionExpire()
    if Daemon {
        fork()
    } else {
        println("work work")
    }
    Logger.Println(srv.Serve(lsr))
}
