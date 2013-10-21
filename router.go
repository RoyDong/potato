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
    Prefix string `yaml:"prefix"`
    Regexp *regexp.Regexp
    Routes []*Route `yaml:"routes"`
}

type Router struct {
    routes []*PrefixedRoutes
    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    return &Router{
        controllers: make(map[string]reflect.Type),
    }
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    //static files
    file := Dir.Static + r.URL.Path[1:]
    if info, e := os.Stat(file); e == nil && (!info.IsDir() || Config.AllowDir) {
        http.ServeFile(w, r, file)
        return
    }

    //dynamic request
    route, params := rt.Route(r.URL.Path)
    if route == nil {
        Logger.Println(r.URL.Path, "not match", file)
        return
    }

    if t, has := rt.controllers[route.Controller]; has {
        controller := reflect.New(t)
        if action := controller.MethodByName(route.Action); action.IsValid() {
            v := reflect.ValueOf(&Controller{
                Request: &Request{Request: r, params: params},
                RW: w,
            })
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

func (rt *Router) InitConfig(filename string) {
    if e := LoadYaml(&rt.routes, filename); e != nil {
        log.Fatal(e)
    }

    for _,pr := range rt.routes {
        pr.Regexp = regexp.MustCompile("^" + pr.Prefix + "(.*)$")
        for _,r := range pr.Routes {
            r.Regexp = regexp.MustCompile("^" + r.Match + "$")
        }
    }
}

func (rt *Router) Route(path string) (*Route, map[string]string) {
    path = strings.ToLower(path)
    for _,pr := range rt.routes {
        if match := pr.Regexp.FindStringSubmatch(path); len(match) == 2 {
            for _,r := range pr.Routes {
                if parts := r.Regexp.FindStringSubmatch(match[1]); len(parts) > 0 {
                    params := make(map[string]string, len(parts) - 1)
                    for i, v := range parts[1:] {
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
