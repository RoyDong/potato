package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "github.com/roydong/potato/lib"
    "encoding/json"
    "html/template"
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


var tpl = NewTemplate()

func TemplateFuncs(funcs map[string]interface{}) {
    tpl.AddFuncs(funcs)
}

/*
consider it's the scope for an http request
*/
type Request struct {
    *http.Request
    Session *Session
    Cookies []*http.Cookie
    Bag     *lib.Tree
    params  []string
    ws      *ws.Conn
    rw      http.ResponseWriter
}

func newRequest(w http.ResponseWriter, r *http.Request) *Request {
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
    if r.ws == nil {
        panic("potato: normal request no websocket")
    }
    var txt string
    if e := ws.Message.Receive(r.ws, &txt); e != nil {
        Logger.Println(e)
        return ""
    }
    return txt
}

func (r *Request) WSSend(txt string) bool {
    if e := ws.Message.Send(r.ws, txt); e != nil {
        Logger.Println(e)
        return false
    }
    return true
}

func (r *Request) WSSendJson(v interface{}) bool {
    if e := ws.JSON.Send(r.ws, v); e != nil {
        Logger.Println(e)
        return false
    }
    return true
}

type Response struct {
    status  int
    code    int
    message string
    body    []byte
    rw      http.ResponseWriter
}

func (r *Request) newResponse() *Response {
    return &Response{rw: r.rw}
}

func (r *Request) TextResponse(txt string) *Response {
    p := r.newResponse()
    p.body = []byte(txt)
    return p
}

func (r *Request) HtmlResponse(name string, data interface{}) *Response {
    var t *template.Template
    buffer := &bytes.Buffer{}
    resp := r.newResponse()
    html := NewHtml()
    html.Data = data
    t = tpl.Template(name)
    if t == nil {
        panic("potato: " + name + " template not found")
    }
    t.Execute(buffer, html)

    //has layout
    if html.layout != "" {
        t = tpl.Template(html.layout)
        if t == nil {
            panic("potato: " + html.layout + " template not found")
        }
        html.Content = template.HTML(buffer.Bytes())
        buffer.Truncate(0)
        t.Execute(buffer, html)
    }
    resp.body = buffer.Bytes()
    return resp
}

func (r *Request) JsonResponse(data interface{}) *Response {
    json, e := json.Marshal(data)
    if e != nil {
        panic("potato: " + e.Error())
    }
    p := r.newResponse()
    p.rw.Header().Set("Content-Type", "application/json;")
    p.body = json
    return p
}

func (r *Request) ErrorResponse(code int, msg string) *Response {
    p := r.newResponse()
    p.code = code
    p.message = msg
    return p
}

func (r *Request) RedirectResponse(url string, status int) *Response {
    p := r.newResponse()
    p.status = status
    p.message = url
    return p
}

func (p *Response) Header() http.Header {
    return p.rw.Header()
}

func (p *Response) SetStatus(status int) {
    p.rw.WriteHeader(status)
}

func (p *Response) SetCookie(c *http.Cookie) {
    http.SetCookie(p.rw, c)
}

func (p *Response) Body() []byte {
    return p.body
}
