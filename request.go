package potato


import (
    "strings"
    "strconv"
    "net/http"
)

type Request struct {
    *http.Request
    params map[string]string
    RemoteIP string
    RemotePort int
    Session *Session
    Cookies []*http.Cookie
}

func NewRequest(r *http.Request, p map[string]string) *Request {
    rq := &Request{
        Request: r,
        params: p,
        Cookies: r.Cookies(),
    }

    rq.parseAddr()

    return rq
}

func (r *Request) SiteAddr() string {
    //TODO
    return "";
}

func (r *Request) IsAjax() bool {
    return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (r *Request) Int(k string) (int, bool) {
    if v, has := r.String(k); has {
        if i, e := strconv.ParseInt(v, 10, 0); e == nil {
            return int(i), true
        }
    }

    return 0, false
}

func (r *Request) Int64(k string) (int64, bool) {
    if v, has := r.String(k); has {
        if i, e := strconv.ParseInt(v, 10, 64); e == nil {
            return i, true
        }
    }

    return 0, false
}

func (r *Request) Float(k string) (float64, bool) {
    if v, has := r.String(k); has {
        if f, e := strconv.ParseFloat(v, 64); e == nil {
            return f, true
        }
    }

    return 0, false
}

func (r *Request) String(k string) (string, bool) {
    if v, has := r.params[k]; has {
        return v, true
    }

    if v := r.FormValue(k); len(v) > 0 {
        return v, true
    }

    return "", false
}

/**
 * get cookie by name
 */
func (r *Request) Cookie(name string) *http.Cookie {
    for _,c := range r.Cookies {
        if c.Name == name { return c }
    }

    return nil
}
