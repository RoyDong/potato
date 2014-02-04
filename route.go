package potato

import (
    "regexp"
    "strings"
)

type Action func(r *Request, p *Response) error

type Route struct {
    name   string
    routes []*Route
    regexp *regexp.Regexp
    action Action
}

func (r *Route) Parse(path string) (*Route, []string) {
    current := r
    params := []string{}
    nodes := strings.Split(
        strings.ToLower(strings.Trim(path, "/")), "/")
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

func (r *Route) Set(path string, action Action) {
    current := r
    nodes := strings.Split(
        strings.ToLower(strings.Trim(path, "/")), "/")
    for _, name := range nodes {
        var found bool
        var rt *Route
        for _, rt = range current.routes {
            if name == rt.name {
                current = rt
                found = true
                break
            }
        }
        if !found {
            rt = &Route{name: name}
            if strings.Contains(name, "(") {
                rt.regexp = regexp.MustCompile("^" + name + "$")
            }
            if current.routes == nil {
                current.routes = make([]*Route, 0)
            }
            current.routes = append(current.routes, rt)
        }
        current = rt
    }
    current.action = action
}

func (r *Route) Name() string {
    return r.name
}

func (r *Route) Action() Action {
    return r.action
}
