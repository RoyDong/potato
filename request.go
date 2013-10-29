package potato


import (
    "time"
    "strconv"
    "net/http"
)

type Request struct {
    *http.Request
    params map[string]string
    Session *Session
    Cookies []*http.Cookie
}

func NewRequest(r *http.Request, p map[string]string) *Request {
    rq := &Request{
        Request: r,
        params: p,
        Cookies: r.Cookies(),
    }

    return rq
}

func (r *Request) GetInt(k string) (int64, bool) {
    if v, has := r.Get(k); has {
        if i, e := strconv.ParseInt(v, 10, 64); e == nil {
            return i, true
        }
    }

    return 0, false
}

func (r *Request) GetFloat(k string) (float64, bool) {
    if v, has := r.Get(k); has {
        if f, e := strconv.ParseFloat(v, 64); e == nil {
            return f, true
        }
    }

    return 0, false
}

func (r *Request) Get(k string) (string, bool) {
    if v, has := r.params[k]; has {
        return v, true
    }

    if v := r.FormValue(k); len(v) > 0 {
        return v, true
    }

    return "", false
}

func (r *Request) InitSession(rp *Response) {
    if c := r.Cookie(SessionCookieName); c != nil {
        r.Session = sessions[c.Value]
    }

    if r.Session == nil {
        r.Session  = NewSession(r)
        rp.SetCookie(&http.Cookie{
            Name: SessionCookieName,
            Value: r.Session.Id,
        })
    } else {
        t := time.Now().Unix()
        if r.Session.LastActivity + SessionDuration < t {
            r.Session.Clear()
        }
        r.Session.LastActivity = t
    }
}

func (r *Request) Cookie(name string) *http.Cookie {
    for _,c := range r.Cookies {
        if c.Name == name { return c }
    }

    return nil
}
