package potato

import (
    ws "code.google.com/p/go.net/websocket"
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
    route, params := rt.route(r.URL.Path)
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
    rt.dispatch(route, request, response)
    rt.TriggerEvent("request_end", request, response)
}

func (rt *Router) route(path string) (*Route, map[string]string) {

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

    return rt.notfoundRoute, make(map[string]string)
}

func (rt *Router) dispatch(route *Route, r *Request, p *Response) {
    defer func() {
        if e := recover(); e != nil {
            if code, ok := e.(int); ok && code == CodeTerminate {
                return
            }

            r.Bag.Set("error", e, true)
            rt.run(rt.errorRoute, r, p)
            L.Println(e)
        }
    }()

    rt.run(route, r, p)
}

func (rt *Router) run(route *Route, r *Request, p *Response) {
    if t, has := rt.controllers[route.Controller]; has {
        c := NewController(t, r, p)
        rt.TriggerEvent("controller_start", c, r, p)
        if action := c.MethodByName(route.Action); action.IsValid() {

            //if controller has Init method, run it first
            if init := c.MethodByName("Init"); init.IsValid() {
                init.Call(nil)
            }

            rt.TriggerEvent("action_start", c, r, p)
            action.Call(nil)
            rt.TriggerEvent("action_end", c, r, p)
            return
        }
    }

    p.WriteHeader(404)
    p.Write([]byte("page not found"))
}
