package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "net/http"
    "regexp"
    "strings"
)

const (
    CodeTerminate = 0
)

type Action func(r *Request, p *Response)

type route struct {
    name   string
    routes []*route
    regexp *regexp.Regexp
    action Action
}

var rootRoute = &route{"", make([]*route, 0), nil, nil}

func (r *route) parse(path string) (*route, []string) {
    current := r
    params := []string{}
    nodes := strings.Split(strings.Trim(path, "/"), "/")
    for _, name := range nodes {
        found := false
        for _, route := range current.routes {
            if name == route.name {
                found = true
                current = route
                break
            } else if route.regexp != nil {
                subs := route.regexp.FindStringSubmatch(name)
                if len(subs) >= 2 {
                    found = true
                    params = append(params, subs[1:]...)
                    current = route
                    break
                }
            }
        }

        if !found {
            return nil, nil
        }
    }

    return current, params
}

func (r *route) set(path string, action Action) {
    current := r
    nodes := strings.Split(strings.Trim(path, "/"), "/")
    for _, name := range nodes {
        var found bool
        var rt *route
        for _, rt = range current.routes {
            if name == rt.name {
                current = rt
                found = true
                break
            }
        }

        if !found {
            var r *regexp.Regexp
            if strings.Contains(name, "(") {
                r = regexp.MustCompile("^" + name + "$")
            }
            rt = &route{name, make([]*route, 0), r, nil}
            current.routes = append(current.routes, rt)
        }

        current = rt
    }

    current.action = action
}

func SetAction(pattern string, action Action) {
    rootRoute.set(pattern, action)
}

func SetErrorAction(action Action) {
    errorAction = action
}

func SetNotfoundAction(action Action) {
    notfoundAction = action
}

type Router struct {
    ws ws.Server
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    route, params := rootRoute.parse(r.URL.Path)
    request := NewRequest(r, params)
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        if conn := rt.ws.Conn(w, r); conn != nil {
            request.WSConn = conn
            defer conn.Close()
        }
    }

    response := &Response{w, "layout"}
    InitSession(request, response)

    defer func() {
        if e := recover(); e != nil {
            if code, ok := e.(int); ok && code == CodeTerminate {
                return
            }
            request.Bag.Set("error", e, true)
            errorAction(request, response)
        }
    }()

    E.TriggerEvent("request_start", request, response)

    if route == nil {
        notfoundAction(request, response)
    } else {
        route.action(request, response)
    }

    E.TriggerEvent("request_done", request, response)
}

var notfoundAction = func(r *Request, p *Response) {
    p.WriteHeader(404)
    p.Write([]byte("page not found"))
}

var errorAction = func(r *Request, p *Response) {
    msg, _ := r.Bag.String("error")
    p.WriteHeader(500)
    p.Write([]byte("we'v got some error " + msg))
}
