package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "log"
    "net/http"
    "reflect"
    "regexp"
    "strings"
)

const (
    CodeTerminate = 0
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

    return nil, nil
}

func (rt *Router) dispatch(route *Route, r *Request, p *Response) {
    defer func() {
        if e := recover(); e != nil {
            if code, ok := e.(int); ok {
                if code == CodeTerminate {
                    return
                }
            }

            rt.TriggerEvent("error", e, r, p)
            L.Println(e)
        }
    }()

    if route != nil {
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
    }

    p.WriteHeader(404)
    p.Write([]byte("page not found"))
}
