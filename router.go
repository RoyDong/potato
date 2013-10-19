package potato

import (
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
    Prefix string `yaml:"prefix"`
    Regexp *regexp.Regexp
    Routes []*Route `yaml:"routes"`
}

type Router struct {
    routes []*PrefixedRoutes
    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    return &Router{controllers: make(map[string]reflect.Type)}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    route, params := rt.Route(r.URL.Path)
    if route == nil {
        log.Println(r.URL.Path, "not match")
        return
    }

    log.Println(r.URL.Path, route)
    if t, has := rt.controllers[route.Controller]; has {
        controller := reflect.New(t)
        if action := controller.MethodByName(route.Action); action.IsValid() {
            v := reflect.ValueOf(&Controller{
                Request: r,
                Params: params,
                RW: w,
            })
            controller.Elem().FieldByName("Controller").Set(v)
            if init := controller.MethodByName("Init"); init.IsValid() {
                init.Call(nil)
            }
            action.Call(nil)
        }
    }
}

func (rt *Router) InitConfig(filename string) {
    if e := LoadYaml(&rt.routes, filename); e != nil {
        log.Fatal(e)
    }

    for _,pr := range rt.routes {
        pr.Regexp = regexp.MustCompile("^" + pr.Prefix + "(.*)$")
        for _,r := range pr.Routes {
            r.Regexp = regexp.MustCompile("^" + r.Match + "$")
            log.Println(pr.Prefix, pr.Regexp, r)
        }
    }
}

func (rt *Router) Route(path string) (*Route, map[string]string) {
    path = strings.ToLower(path)
    for _,pr := range rt.routes {
        if m := pr.Regexp.FindStringSubmatch(path); len(m) == 2 {
            for _,r := range pr.Routes {
                if p := r.Regexp.FindStringSubmatch(m[1]); len(p) > 0 {
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

func (rt *Router) RegController(c interface{}) {
    elem := reflect.ValueOf(c).Elem()
    t := elem.Type()
    if elem.FieldByName("Controller").CanSet() {
        rt.controllers[t.Name()] = t
    } else {
        log.Println("Controller: " + t.Name() + " must embeded from *potato.Controller")
    }
}

func (rt *Router) RegControllers(cs []interface{}) {
    for _,c := range cs {
        rt.RegController(c)
    }
}
