package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "net/http"
    "strconv"
)

type Request struct {
    *http.Request
    WSConn  *ws.Conn
    params  []string
    Session *Session
    Cookies []*http.Cookie
    Bag     *Tree
}

func NewRequest(r *http.Request) *Request {
    return &Request{
        Request: r,
        Cookies: r.Cookies(),
        Bag:     NewTree(),
    }
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
    if k[0] == '$' {
        n, e := strconv.ParseInt(k[1:], 10, 0)
        if e == nil && n > 0 && int(n) <= len(r.params) {
            return r.params[n-1], true
        }
    }
    if v := r.FormValue(k); len(v) > 0 {
        return v, true
    }
    return "", false
}

func (r *Request) Cookie(name string) *http.Cookie {
    for _, c := range r.Cookies {
        if c.Name == name {
            return c
        }
    }
    return nil
}

func (r *Request) WSReceive() string {
    if r.WSConn == nil {
        panic("potato: normal request no websocket")
    }
    var txt string
    if e := ws.Message.Receive(r.WSConn, &txt); e != nil {
        Logger.Println(e)
        return ""
    }
    return txt
}

func (r *Request) WSSend(txt string) bool {
    if e := ws.Message.Send(r.WSConn, txt); e != nil {
        Logger.Println(e)
        return false
    }
    return true
}

func (r *Request) WSSendJson(v interface{}) bool {
    if e := ws.JSON.Send(r.WSConn, v); e != nil {
        Logger.Println(e)
        return false
    }
    return true
}
