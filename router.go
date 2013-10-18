package potato

import (
    "log"
    "regexp"
    "net/http"
    "launchpad.net/goyaml"
)

type Route struct {
    Name string `yaml:"name"`
    Controller string `yaml:controller"`
    Action string `yaml:action"`
    Match string `yaml:match"`
    Keys []string `yaml:keys"`
    Regexp *regexp.Regexp
}

type Router struct {
    routes []*Route
    controllers map[string]reflect.Type
}

func NewRouter() *Router {
    r := new(Router)
    r.controllers = make(map[string]reflect.Type)
    return r
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    rt.Route(r.URL.Path)
}

func (rt *Router) InitFile(filename string) {
    rt.routes = make([]*Route, 0)
    LoadYaml(&rt.routes, filename)

    for _,r := range rt.routes {
        r.Regexp = regexp.MustCompile("(?i)" + r.Match)
    }
}

func (rt *Router) Init(yaml []byte) {
    e := goyaml.Unmarshal(yaml, &rt.routes)
    if e != nil {
        log.Fatal(e)
    }

    for _,r := range rt.routes {
        r.Regexp = regexp.MustCompile("(?i)" + r.Match)
    }
}

func (rt *Router) Route(path string) {
    for _,r := range rt.routes {
        if p := r.Regexp.FindStringSubmatch(path); len(p) > 0 {
            for _,v := range p {
                log.Println(path, v)
            }
        }
    }
}

func (rt *Router) RegController(c interface{}) {
    elem := reflect.ValueOf(c).Elem()
    t := elem.Type()
    if elem.FieldByName("Controller").CanSet() {
        rt.controllers[t.Name()] = t
    } else {
        Logger.Println("Controller: " + t.Name() + " must embeded from server.Message")
    }
}

func (rt *Router) RegControllers(cs []interface{}) {
    for _,c := range cs {
        rt.RegController(c)
    }
}
