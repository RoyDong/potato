package potato

import (
    "os"
    "log"
    "strings"
    "regexp"
    "reflect"
    "net/http"
)

type Route struct {
    Name string `yaml:"name"`
    Controller string `yaml:"controller"`
    Action string `yaml:"action"`
    Pattern string `yaml:"pattern"`
    Keys []string `yaml:"keys"`
    Regexp *regexp.Regexp
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

    //all grouped routes
    routes []*PrefixedRoutes

    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    return &Router{
        controllers: make(map[string]reflect.Type),
    }
}

/**
 * Controllers register controllers on router
 */
func (rt *Router) Controllers(cs []interface{}) {
    for _,c := range cs {
        elem := reflect.ValueOf(c).Elem()

        //Controller must embeded from *potato.Controller
        if elem.FieldByName("Controller").CanSet() {
            t := elem.Type()
            rt.controllers[t.Name()] = t
        }
    }
}

func (rt *Router) LoadConfig(filename string) {
    if e := LoadYaml(&rt.routes, filename); e != nil {
        log.Fatal(e)
    }

    for _,pr := range rt.routes {

        //prepare regexps for prefixed routes
        pr.Regexp = regexp.MustCompile("^" + pr.Prefix + "(.*)$")
        for _,r := range pr.Routes {
            r.Regexp = regexp.MustCompile("^" + r.Pattern + "$")
        }
    }
}


func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    L.Println(r.Method, r.Proto, r.Host, r.RequestURI, r.RemoteAddr)

    //static files, deny all dir requests
    file := Dir.Static + r.URL.Path[1:]
    if info, e := os.Stat(file); e == nil && !info.IsDir() {
        http.ServeFile(w, r, file)

    //dynamic requests
    } else {
        route, params := rt.route(r.URL.Path);
        request := NewRequest(r, params)
        response := &Response{w}
        InitSession(request, response)

        if route == nil {
            rt.handleError(&Error{http.StatusNotFound, "page not found"},
                    request, response)
        } else {
            rt.handle(route, request, response)
        }
    }
}

func (rt *Router) route(path string) (*Route, map[string]string) {

    //case insensitive
    //make sure the patterns in routes.yml is lower case too
    path = strings.ToLower(path)

    //check prefixes
    for _,pr := range rt.routes {
        if m := pr.Regexp.FindStringSubmatch(path); len(m) == 2 {

            //check routes on matched prefix
            for _,r := range pr.Routes {
                if p := r.Regexp.FindStringSubmatch(m[1]); len(p) > 0 {

                    //get params for matched route
                    params := make(map[string]string, len(p) - 1)
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

func (rt *Router) handle(route *Route, r *Request, p *Response) {

    //handle panics
    defer func () {
        if e := recover(); e != nil {
            rt.handleError(e, r, p)
        }
    }()

    if t, has := rt.controllers[route.Controller]; has {
        controller := rt.controller(t, r, p)

        //if action not found check the NotFound method
        action := controller.MethodByName(route.Action)
        if !action.IsValid() {
            if nf := controller.MethodByName(NotFoundRoute.Action); nf.IsValid() {
                action = nf
            } else {
                Panic(http.StatusNotFound, "page not found")
            }
        }

        //if controller has Init method, run it first
        if init := controller.MethodByName("Init"); init.IsValid() {
            init.Call(nil)
        }

        action.Call(nil)
    } else {
        Panic(http.StatusNotFound, "page not found")
    }
}

//initialize controller
func (rt *Router) controller(t reflect.Type, r *Request, p *Response) reflect.Value {
    controller := reflect.New(t)
    controller.Elem().FieldByName("Controller").
            Set(reflect.ValueOf(Controller{
                    Request: r,
                    Response: p,
                    Layout: "layout",
                    Template: "index",
                    Title: AppName}))

    return controller
}

func (rt *Router) handleError(e interface{}, r *Request, p *Response) {
    if err, ok := e.(*Error); ok {
        if t, has := rt.controllers[ServerErrorRoute.Controller]; has {
            controller := rt.controller(t, r, p)
            if action := controller.MethodByName(ServerErrorRoute.Action);
                    action.IsValid() {

                action.Call([]reflect.Value{reflect.ValueOf(err)})
                return
            }
        }

        p.Write([]byte(err.String()))
    }

    L.Println(e)
}
