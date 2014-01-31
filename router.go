package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "net/http"
    "reflect"
    "regexp"
    "strings"
)

const (
    CodeTerminate = 0
)

var (
    ErrorRouteName string
    NotfoundRouteName string
)

type Action func(r *Request, p *response)

type route struct {
    name   string
    routes []*route
    regexp *regexp.Regexp
    action Action
}

var route = &route{"", make([]*route, 0)}

func (r *route) route(path string) (*route, bool) {
    current := r
    nodes := string.Split(path)
    for _, name := range nodes {
        found := false
        for _, route := range current.routes {
            if node == route.name {
                current = route
                found = true
                break
            }
        }

        if !found {
            return nil, false
        }
    }

    return current, true
}

func (r *route) set(path string, action Action) {
    current := r
    nodes := string.Split(path)
    for _, name := range nodes {
        var found bool
        var route *route
        for _, route = range current.routes {
            if name == route.name {
                current = route
                found = true
                break
            }
        }

        if !found {
            var r *regexp.Regexp
            if strings.Contains(name, "(") {
                r = regexp.MustCompile("^" + name + "$")
            }
            route = &route{name, make([]*route, 0), r}
            current.routes = append(next.routes, route)
        }

        current = route
    }

    current.action = action
}

func SetAction(pattern string, action Action) {
    route.set(pattern, action)
}

type Router struct {
    Event
    ws            ws.Server
}

func NewRouter() *Router {
    return &Router{
        Event{make(map[string][]EventHandler)},
        ws.Server{},
    }
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    route, params := rt.Route(r.URL.Path)
    request := NewRequest(r, params)
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        if conn := rt.ws.Conn(w, r); conn != nil {
            request.WSConn = conn
            defer conn.Close()
        }
    }

    response := &Response{ResponseWriter: w}
    InitSession(request, response)
    rt.TriggerEvent("request_start", request, response)

    if route == nil {
        rt.handleError("page not found",
            request, response)
    } else {
        rt.Dispatch(route, request, response)
    }

    rt.TriggerEvent("request_end", request, response)
}

func (rt *Router) Route(path string) (*Route, map[string]string) {

    //case insensitive
    //make sure the patterns in routes.yml is lower case too
    path = strings.ToLower(path)

    //check prefixes
    for _, pr := range rt.routes {
        if m := pr.Regexp.FindStringSubmatch(path); len(m) == 2 {

            //check routes on matched prefix
            for _, r := range pr.Routes {
                if p := r.Regexp.FindStringSubmatch(m[1]); len(p) > 0 {

                    //get params for matched route
                    params := make(map[string]string, len(p)-1)
                    for i, v := range p[1:] {
                        params[r.Keys[i]] = v
                    }

                    return r, params
                }
            }
        }
    }

    return nil, nil
}

func (rt *Router) Dispatch(route *Route, r *Request, p *Response) {

    //handle panic
    defer func() {
        if e := recover(); e != nil {
            rt.handleError(e, r, p)
        }
    }()

    if t, has := rt.controllers[route.Controller]; has {
        c := NewController(t, r, p)
        rt.TriggerEvent("controller_start", r, p, c)

        //if action not found check the NotFound method
        action := c.MethodByName(route.Action)
        if !action.IsValid() {
            panic("page not found")
        }

        //if controller has Init method, run it first
        if init := c.MethodByName("Init"); init.IsValid() {
            init.Call(nil)
        }

        rt.TriggerEvent("action_start", r, p, c)
        action.Call(nil)
        rt.TriggerEvent("action_end", r, p, c)
    } else {
        panic("page not found")
    }
}

func (rt *Router) handleError(e interface{}, r *Request, p *Response) {
    var message string
    if v, ok := e.(string); ok {
        message = v
    } else if v, ok := e.(error); ok {
        message = v.Error()
    }

    if message == "redirect" {
        return
    }

    rt.TriggerEvent("error", r, p, message)
    if !p.Sent {
        if r.IsAjax() {
            message = fmt.Sprintf(`{error:"%s"}`,
                strings.Replace(message, `"`, `\"`, -1))
        }
        p.Write([]byte(message))
        p.Sent = true
    }

    L.Println(e)
}
