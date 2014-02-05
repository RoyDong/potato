package potato

import (
    "encoding/json"
    "net/http"
)

var DefaultLayout = "layout"

type Response struct {
    http.ResponseWriter
    Layout string
}

func NewResponse(w http.ResponseWriter) *Response {
    return &Response{w, DefaultLayout}
}

func (r *Response) SetCookie(c *http.Cookie) {
    http.SetCookie(r, c)
}

func (r *Response) Redirect(request *Request, url string, code int) {
    http.Redirect(r, request.Request, url, code)
}

func (r *Response) RenderText(t string) {
    r.Write([]byte(t))
}

var tpl = NewTemplate()

func TemplateFuncs(funcs map[string]interface{}) {
    tpl.AddFuncs(funcs)
}

func (r *Response) Render(name string, data interface{}) {
    if t := tpl.Template(r.Layout); t != nil {
        html := NewHtml()
        html.Data = data
        html.Content = tpl.Include(name, html)
        t.Execute(r, html)
    } else {
        panic("potato:" + r.Layout + " template not found")
    }
}

func (r *Response) RenderPartial(name string, data interface{}) {
    if t := tpl.Template(name); t != nil {
        t.Execute(r, data)
    } else {
        panic("potato: " + name + " template not found")
    }
}

func (r *Response) RenderJson(v interface{}) {
    json, e := json.Marshal(v)
    if e != nil {
        panic("potato:" + e.Error())
    }

    r.Header().Set("Content-Type", "application/json; charset=utf8")
    r.Write(json)
}
