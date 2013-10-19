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
    Prefix string `yaml:"prefix"`
    Match string `yaml:"match"`
    Keys []string `yaml:"keys"`
    Regexp *regexp.Regexp
}

type Router struct {
    routes map[string][]*Route
    prefixes map[string]*regexp.Regexp
    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    return &Router{controllers: make(map[string]reflect.Type)}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    route, params := rt.Route(r.URL.Path)
    if route == nil {
        log.Println("not match")
        return
    }

    if t, has := rt.controllers[route.Controller]; has {
        controller := reflect.New(t)
        if action := controller.MethodByName(route.Action); action.IsValid() {
            v := reflect.ValueOf(&Controller{
                Request: r,
                Params: params,
                RW w,
            })
            controller.Elem().FieldByName("Controller").Set(v)
            if init := controller.MethodByName("Init"); init.IsValid() {
                init.Call(nil)
            }
            action.Call(nil)
        }
    } else {
        log.Println("page not found")
    }
}

func (rt *Router) InitConfig(filename string) {
    routes := make([]*Route, 0)
    if e := LoadYaml(&routes, filename); e != nil {
        log.Fatal(e)
    }

    rt.routes = make(map[string][]*Route)
    rt.prefixes = make(map[string]*regexp.Regexp)
    for _,r := range routes {
        if r.Prefix == "" {
            r.Prefix = "~"
        }

        if _,has := rt.prefixes[r.Prefix]; !has {
            rt.routes[r.Prefix] = make([]*Route, 0, 1)

            if r.Prefix == "~" {
                rt.prefixes[r.Prefix] = regexp.MustCompile("^" + r.Prefix + "(.*)$")
            }
        }

        rt.routes[r.Prefix] = append(rt.routes[r.Prefix], r)
        r.Regexp = regexp.MustCompile("^" + r.Match + "$")
    }
}

func (rt *Router) Route(path string) (*Route, map[string]string) {
    path = strings.ToLower(path)
    for prefix, regexp := range rt.prefixes {
        if parts := regexp.FindStringSubmatch(path); len(parts) == 2 {
            for _,r := range rt.routes[prefix] {
                if p := r.Regexp.FindStringSubmatch(parts[1]); len(p) > 0 {
                    params := make(map[string]string, len(p) - 1)
                    for i, v := range p[1:] {
                        params[r.Keys[i]] = v
                    }

                    return r, params
                }
            }
        }
    }

    for _,r := range rt.routes["~"] {
        if p := r.Regexp.FindStringSubmatch(path); len(p) > 0 {
            params := make(map[string]string, len(p) - 1)
            for i, v := range p[1:] {
                params[r.Keys[i]] = v
            }

            return r, params
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
