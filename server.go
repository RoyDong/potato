package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "strings"
    "fmt"
    "net"
    "net/http"
    "os"
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
 *  after_action, after running action
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

var NotfoundAction = func(r *Request, p *Response) {
    p.WriteHeader(404)
    p.Write([]byte("page not found"))
}

var ErrorAction = func(r *Request, p *Response) {
    msg, has := r.Bag.String("error")
    if !has {
        msg = "unknown"
    }
    p.WriteHeader(500)
    p.Write([]byte("error: " + msg))
}

type Terminate string

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

    defer func() {
        event.Trigger("after_action", request, response)
        if e := recover(); e != nil {
            if _, ok := e.(Terminate); ok {
                return
            }
            request.Bag.Set("error", e, true)
            ErrorAction(request, response)
            Logger.Println(e)
        }
        event.Trigger("response", request, response)
    }()

    var rt *Route
    event.Trigger("request", request, response)
    rt, request.params = route.Parse(r.URL.Path)
    event.Trigger("before_action", request, response)
    if rt == nil {
        NotfoundAction(request, response)
    } else {
        rt.action(request, response)
    }
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
