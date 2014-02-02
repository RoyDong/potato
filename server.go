package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "strings"
    "fmt"
    "net"
    "net/http"
    "os"
)

const (
    TerminateCode = 0
)

var event = NewEvent()

func AddEventHandler(name string, handler EventHandler) {
    event.AddEventHandler(name, handler)
}

var route = &Route{}

func SetAction(action Action, patterns ...string) {
    for _, pattern := range patterns {
        route.Set(pattern, action)
    }
}

var NotfoundAction = func(r *Request, p *Response) {
    p.WriteHeader(404)
    p.Write([]byte("page not found"))
}

var ErrorAction = func(r *Request, p *Response) {
    msg, _ := r.Bag.String("error")
    p.WriteHeader(500)
    p.Write([]byte("we'v got some error " + msg))
}

type handler struct {
    ws.Server
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    request := NewRequest(r)
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        if conn := h.Conn(w, r); conn != nil {
            request.WSConn = conn
            defer conn.Close()
        }
    }

    response := NewResponse(w)
    InitSession(request, response)
    defer func() {
        if e := recover(); e != nil {
            if code, ok := e.(int); ok && code == TerminateCode {
                return
            }
            request.Bag.Set("error", e, true)
            ErrorAction(request, response)
        }
    }()

    var rt *Route
    event.TriggerEvent("request", request, response)
    event.TriggerEvent("before_route", request, response)
    rt, request.params = route.Parse(r.URL.Path)
    event.TriggerEvent("after_route", request, response)
    if rt == nil {
        NotfoundAction(request, response)
    } else {
        rt.action(request, response)
    }
    event.TriggerEvent("response", request, response)
}

var tpl *Template

func TemplateFuncs(funcs map[string]interface{}) {
    tpl.AddFuncs(funcs)
}

func Serve() {
    var e error
    var lsn net.Listener
    if len(SockFile) > 0 {
        os.Remove(SockFile)
        lsn, e = net.Listen("unix", SockFile)
        if e != nil {
            Logger.Println("fail to open socket file", e)
        } else {
            os.Chmod(SockFile, os.ModePerm)
        }
    }
    if lsn == nil {
        lsn, e = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    }
    if e != nil {
        Logger.Fatal(e)
    }
    defer lsn.Close()
    tpl.loadTemplateFiles(tpl.dir)
    go sessionExpire()
    Logger.Println("work work")
    server := &http.Server{Handler: &handler{ws.Server{}}}
    Logger.Println(server.Serve(lsn))
}
