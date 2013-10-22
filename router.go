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

//routes are grouped by their prefixes
//when routing a url, first match the prefixes
//then match the patterns of each route
type PrefixedRoutes struct {
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

func (rt *Router) Init(cs []interface{}) {
    for _,c := range cs {
        rt.RegController(c)
    }
}

func (rt *Router) InitConfig(filename string) {
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

    //static files, deny all dir requests
    file := Dir.Static + r.URL.Path[1:]
    if info, e := os.Stat(file); e == nil && !info.IsDir() {
        http.ServeFile(w, r, file)

    //dynamic requests
    } else {
        route, params := rt.Route(r.URL.Path);
        rq := &Request{r, params}
        rp := &Response{w}
        rt.RunAction(route, rq, rp)
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

    Logger.Println(path, "no route found")
    return NotFoundRoute, nil
}

func (rt *Router) RunAction(r *Route, rq *Request, rp *Response) {
    if t, has := rt.controllers[r.Controller]; has {

        //initialize controller
        controller := reflect.New(t)
        controller.Elem().FieldByName("Controller").
                Set(reflect.ValueOf(&Controller{rq, rp}))

        //if action not found check the NotFound method
        action := controller.MethodByName(r.Action)
        if !action.IsValid() && r.Action != NotFoundRoute.Action {
            Logger.Println(r.Action, "action not found")

            if nf := controller.MethodByName(NotFoundRoute.Action); nf.IsValid() {
                action = nf
            } else {
                goto NF
            }
        }

        //if controller has Init method, run it first
        if init := controller.MethodByName("Init"); init.IsValid() {
            init.Call(nil)
        }

        action.Call(nil)
        return
    }

    NF: http.NotFound(rp.ResponseWriter, rq.Request)
}


