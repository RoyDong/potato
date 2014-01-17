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

type Route struct {
    Name       string   `yaml:"name"`
    Controller string   `yaml:"controller"`
    Action     string   `yaml:"action"`
    Pattern    string   `yaml:"pattern"`
    Keys       []string `yaml:"keys"`
    Regexp     *regexp.Regexp
}

/**
 * routes are grouped by their prefixes
 * when routing a url, first match the prefixes
 * then match the patterns of each route
 */
type PrefixedRoutes struct {
    Prefix string `yaml:"prefix"`
    Regexp *regexp.Regexp
    Routes []*Route `yaml:"routes"`
}

type Router struct {
    Event
    ws  ws.Server

    //all grouped routes
    routes []*PrefixedRoutes

    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    return &Router{
        Event:       Event{make(map[string][]EventHandler)},
        ws:          ws.Server{},
        controllers: make(map[string]reflect.Type),
    }
}

/**
 * Controllers register controllers on router
 */
func (rt *Router) SetControllers(cs map[string]interface{}) {
    for n, c := range cs {
        elem := reflect.Indirect(reflect.ValueOf(c))

        //Controller must embeded from potato.Controller
        if elem.FieldByName("Controller").CanSet() {
            rt.controllers[n] = elem.Type()
        }
    }
}

func (rt *Router) LoadRouteConfig(filename string) {
    if e := LoadYaml(&rt.routes, filename); e != nil {
        log.Fatal(e)
    }

    for _, pr := range rt.routes {

        //prepare regexps for prefixed routes
        pr.Regexp = regexp.MustCompile("^" + pr.Prefix + "(.*)$")
        for _, r := range pr.Routes {
            r.Regexp = regexp.MustCompile("^" + r.Pattern + "$")
        }
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
