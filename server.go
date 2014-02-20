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
    resp := r.TextResponse(fmt.Sprintf("code: %d, message: %s", c, m))
    resp.WriteHeader(c)
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
    } else if resp = act(req); resp.code >= 400 {
        resp = ErrorAction(req, resp.code, resp.message)
    }

    event.Trigger("respond", req, resp)
    if resp.code >= 300 && resp.code < 400 {
        http.Redirect(w, r, resp.message, resp.code)
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
