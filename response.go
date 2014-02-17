package potato

import (
    "encoding/json"
    "net/http"
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

type Response struct {
    Status   int
    Header   http.Header
    cookies  []*http.Cookie
    body     []byte
    code     int
    message  string
    redirect string
}

var tpl = NewTemplate()

func TemplateFuncs(funcs map[string]interface{}) {
    tpl.AddFuncs(funcs)
}

func NewResponse() *Response {
    return &Response{
        Status: http.StatusOK,
        Header: make(http.Header),
        cookies: make([]*http.Cookie, 0),
    }
}

func TextResponse(txt string) *Response {
    r := NewResponse()
    r.Write([]byte(txt))
    return r
}

func HtmlResponse(name string, data interface{}) *Response {
    if t := tpl.Template(DefaultLayout); t != nil {
        r := NewResponse()
        html := NewHtml()
        html.Data = data
        html.Content = tpl.Include(name, html)
        b := &bytes.Buffer{}
        t.Execute(b, html)
        r.body = b.Bytes()
        return r
    }
    panic("potato: " + DefaultLayout + " template not found")
}

func PartialResponse(name string, data interface{}) *Response {
    if t := tpl.Template(name); t != nil {
        r := NewResponse()
        b := &bytes.Buffer{}
        t.Execute(b, data)
        r.body = b.Bytes()
        return r
    }
    panic("potato: " + name + " template not found")
}

func JsonResponse(data interface{}) *Response {
    json, e := json.Marshal(data)
    if e != nil {
        panic("potato: " + e.Error())
    }
    r := NewResponse()
    r.Header.Set("Content-Type", "application/json; charset=utf8")
    r.body = json
    return r
}

func ErrorResponse(c int, m string) *Response {
    r := NewResponse()
    r.code = c
    r.message = m
    return r
}

func RedirectResponse(url string, status int) *Response {
    r := NewResponse()
    r.Status = status
    r.redirect = url
    return r
}

func (r *Response) SetCookie(c *http.Cookie) *Response {
    r.cookies = append(r.cookies, c)
    return r
}

func (r *Response) Write(p []byte) (int, error) {
    r.body = p
    return len(r.body), nil
}

func (r *Response) Body() []byte {
    return r.body
}

func (r *Response) flush(w http.ResponseWriter, rq *http.Request) {
    header := w.Header()
    *(&header) = r.Header
    for _, c := range r.cookies {
        http.SetCookie(w, c)
    }
    if r.redirect != "" {
        http.Redirect(w, rq, r.redirect, r.Status)
    } else {
        w.WriteHeader(r.Status)
        w.Write(r.body)
    }
}
