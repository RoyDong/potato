package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "github.com/roydong/potato/lib"
    "encoding/json"
    "net/http"
    "strconv"
    "bytes"
)

const (
    StatusBadRequest              = 400
    StatusPaymentRequired         = 402
    StatusForbidden               = 403
    StatusNotFound                = 404
    StatusUnsupportedMediaType    = 415
    StatusInternalServerError     = 500
    StatusServiceUnavailable      = 503
    StatusGatewayTimeout          = 504
    StatusHTTPVersionNotSupported = 505
)

var DefaultLayout = "layout"


var tpl = NewTemplate()

func TemplateFuncs(funcs map[string]interface{}) {
    tpl.AddFuncs(funcs)
}

type Request struct {
    *http.Request
    WSConn  *ws.Conn
    Session *Session
    Cookies []*http.Cookie
    Bag     *lib.Tree
    params  []string
    rw      http.ResponseWriter
}

func NewRequest(w http.ResponseWriter, r *http.Request) *Request {
    return &Request{
        Request: r,
        Cookies: r.Cookies(),
        Bag: lib.NewTree(),
        rw: w,
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

type Response struct {
    Request *Request
    Status   int
    cookies  []*http.Cookie
    body     []byte
    code     int
    message  string
    redirect string
    rw       http.ResponseWriter
}

func (r *Request) newResponse() *Response {
    return &Response{
        Request: r,
        rw: r.rw,
        Status: http.StatusOK,
        cookies: make([]*http.Cookie, 0),
    }
}

func (r *Request) TextResponse(txt string) *Response {
    p := r.newResponse()
    p.body = []byte(txt)
    return p
}

func (r *Request) HtmlResponse(name string, data interface{}) *Response {
    if t := tpl.Template(DefaultLayout); t != nil {
        p := r.newResponse()
        html := NewHtml()
        html.Data = data
        html.Content = tpl.Include(name, html)
        b := &bytes.Buffer{}
        t.Execute(b, html)
        p.body = b.Bytes()
        return p
    }
    panic("potato: " + DefaultLayout + " template not found")
}

func (r *Request) PartialResponse(name string, data interface{}) *Response {
    if t := tpl.Template(name); t != nil {
        p := r.newResponse()
        b := &bytes.Buffer{}
        t.Execute(b, data)
        p.body = b.Bytes()
        return p
    }
    panic("potato: " + name + " template not found")
}

func (r *Request) JsonResponse(data interface{}) *Response {
    json, e := json.Marshal(data)
    if e != nil {
        panic("potato: " + e.Error())
    }
    p := r.newResponse()
    p.Header().Set("Content-Type", "application/json;")
    p.body = json
    return p
}

func (r *Request) ErrorResponse(c int, m string) *Response {
    p := r.newResponse()
    p.code = c
    p.message = m
    return p
}

func (r *Request) RedirectResponse(url string, status int) *Response {
    p := r.newResponse()
    p.Status = status
    p.redirect = url
    return p
}

func (p *Response) Header() http.Header {
    return p.rw.Header()
}

func (p *Response) SetCookie(c *http.Cookie) *Response {
    http.SetCookie(p.rw, c)
    return p
}

func (p *Response) Body() []byte {
    return p.body
}

func (p *Response) flush() {
    if p.redirect == "" {
        p.rw.WriteHeader(p.Status)
        p.rw.Write(p.body)
    } else {
        http.Redirect(p.rw, p.Request.Request, p.redirect, p.Status)
    }
}
