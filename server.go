package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "github.com/roydong/potato/lib"
    "fmt"
    "net"
    "net/http"
    "os"
    "log"
    "sync"
    "syscall"
    "strings"
)

/*
events:
    before_init
    after_init
    run
    request  
    action
    respond
*/
var event = lib.NewEvent()

func AddHandler(name string, handler lib.EventHandler) {
    event.AddHandler(name, handler)
}

var route = &Route{}

func SetAction(action Action, patterns ...string) {
    for _, pattern := range patterns {
        route.Set(pattern, action)
    }
}

var ErrorAction = func(r *Request, c int, m string) *Response {
    resp := r.TextResponse(fmt.Sprintf("code: %d, message: %s", c, m))
    resp.SetStatus(c)
    return resp
}

type handler struct {
    ws.Server
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    req := newRequest(w, r)
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        if conn := h.Conn(w, r); conn != nil {
            req.ws = conn
            defer conn.Close()
        }
    }
    initSession(req)
    event.Trigger("request", req)
    if Env == "dev" {
        tpl.Load(TplDir)
    }

    var act Action
    act, req.params = route.Parse(r.URL.Path)
    event.Trigger("action", req)

    var resp *Response
    if act == nil {
        resp = ErrorAction(req, 404, "route not found")
    } else if resp = act(req); resp.code > 0 {
        resp = ErrorAction(req, resp.code, resp.message)
    }

    event.Trigger("respond", req, resp)
    if resp.status >= 300 && resp.status < 400 {
        http.Redirect(w, r, resp.message, resp.status)
    } else {
        w.Write(resp.body)
    }
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
        Logger.Fatal("potato:", e)
    }
    return lsr
}

func replaceLogio() {
    f, e := os.OpenFile(LogDir+Env+".log",
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if e != nil {
        Logger.Fatal("potato: log file", e)
    }
    Logger = log.New(f, "", log.LstdFlags)
    if f, e = os.OpenFile("/dev/null", os.O_RDWR, 0); e == nil {
        fd := int(f.Fd())
        syscall.Dup2(fd, int(os.Stdin.Fd()))
        syscall.Dup2(fd, int(os.Stdout.Fd()))
        syscall.Dup2(fd, int(os.Stderr.Fd()))
    }
}

var wg = &sync.WaitGroup{}

func serve() {
    defer wg.Done()
    srv := &http.Server{Handler: &handler{ws.Server{}}}
    lsr := listener()
    defer lsr.Close()
    Logger.Println(srv.Serve(lsr))
}

func Run() {
    tpl.Load(TplDir)
    go sessionExpire()
    wg.Add(1)
    go serve()
    println("work work")
    if Daemon {
        replaceLogio()
    }
    event.Trigger("run")
    wg.Wait()
}
