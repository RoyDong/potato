package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "github.com/roydong/potato/lib"
    "fmt"
    "net"
    "net/http"
    "os"
    "strings"
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
    resp := TextResponse(fmt.Sprintf("code: %d, message: %s", c, m))
    resp.Status = c
    return resp
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
    initSession(request, w)
    if Env == "dev" {
        tpl.Load(TplDir)
    }

    var action Action
    event.Trigger("request", request)
    action, request.params = route.Parse(r.URL.Path)
    event.Trigger("before_action", request)

    var resp *Response
    if action == nil {
        resp = ErrorAction(request, 404, "page not found")
    } else if resp = action(request); resp.code > 0 {
        resp = ErrorAction(request, resp.code, resp.message)
    }

    event.Trigger("response", request, resp)
    resp.flush(w, request.Request)
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

func Serve() {
    tpl.Load(TplDir)
    go sessionExpire()
    srv := &http.Server{Handler: &handler{ws.Server{}}}
    lsr := listener()
    defer lsr.Close()
    event.Trigger("before_serve")
    defer event.Trigger("after_serve")
    Logger.Println(srv.Serve(lsr))
}
