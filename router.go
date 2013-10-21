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
    Match string `yaml:"match"`
    Keys []string `yaml:"keys"`
    Regexp *regexp.Regexp
}

type PrefixedRoutes struct {

    //routes are grouped by their prefixes
    //when matching a route, first match its prefix then match the rest
    Prefix string `yaml:"prefix"`
    Regexp *regexp.Regexp
    Routes []*Route `yaml:"routes"`
}

type Router struct {

    //all the grouped routes
    routes []*PrefixedRoutes
    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    return &Router{
        controllers: make(map[string]reflect.Type),
    }
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    //static files, disallow all dir requests
    file := Dir.Static + r.URL.Path[1:]
    if info, e := os.Stat(file); e == nil && !info.IsDir() {
        http.ServeFile(w, r, file)
        return
    }

    //dynamic request
    route, params := rt.Route(r.URL.Path)
    if route == nil {
        Logger.Println(r.URL.Path, "not match")
        return
    }

    if t, has := rt.controllers[route.Controller]; has {
        controller := reflect.New(t)
        if action := controller.MethodByName(route.Action); action.IsValid() {
            request := &Request{Request: r, Route: route, params: params}
            v := reflect.ValueOf(&Controller{Request: request, RW: w})
            controller.Elem().FieldByName("Controller").Set(v)
            if init := controller.MethodByName("Init"); init.IsValid() {
                init.Call(nil)
            }
            action.Call(nil)
        } else {
            log.Println("action not found")
        }
    } else {
        log.Println("controller not found")
    }
}

func (rt *Router) RunAction(r *Request) {
}

func (rt *Router) InitConfig(filename string) {
    if e := LoadYaml(&rt.routes, filename); e != nil {
        log.Fatal(e)
    }

    for _,pr := range rt.routes {

        //prepare regexp for prefixes and routes
        pr.Regexp = regexp.MustCompile("^" + pr.Prefix + "(.*)$")
        for _,r := range pr.Routes {
            r.Regexp = regexp.MustCompile("^" + r.Match + "$")
        }
    }
}

func (rt *Router) Route(path string) (*Route, map[string]string) {

    //case insensitive
    //make sure the patterns in routes.yml is lower case too
    path = strings.ToLower(path)

    //check prefixes
    for _,pr := range rt.routes {
        if m := pr.Regexp.FindStringSubmatch(path); len(m) == 2 {

            //check routes on matched prefix
            for _,r := range pr.Routes {
                if p := r.Regexp.FindStringSubmatch(m[1]); len(p) > 0 {

                    //set params for matched route
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

/**
 * RegController register controller on router
 */
func (rt *Router) RegController(c interface{}) {
    elem := reflect.ValueOf(c).Elem()

    //Controller must embeded from *potato.Controller
    if elem.FieldByName("Controller").CanSet() {
        t := elem.Type()
        rt.controllers[t.Name()] = t
    }
}

func (rt *Router) RegControllers(cs []interface{}) {
    for _,c := range cs {
        rt.RegController(c)
    }
}
