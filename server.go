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

    if Env == "dev" {
        tpl.Load(TplDir)
    }

    var action Action
    action, request.params = route.Parse(r.URL.Path)
    event.Trigger("before_action", request, response)
    if action == nil {
        NotfoundAction(request, response)
    } else {
        if e := action(request, response); e != nil {
            request.Bag.Set("error", e)
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
