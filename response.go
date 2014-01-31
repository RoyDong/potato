package potato

import (
    "encoding/json"
    "net/http"
)

type Response struct {
    http.ResponseWriter
    Layout string
}

func (r *Response) SetCookie(c *http.Cookie) {
    http.SetCookie(r, c)
}

func (r *Response) Redirect(request *Request, url string, code int) {
    http.Redirect(r, request.Request, url, code)
    panic(CodeTerminate)
}

func (r *Response) RenderText(t string) {
    r.Write([]byte(t))
}

func (r *Response) Render(name string, data interface{}) {
    if t := T.Template(r.Layout); t != nil {
        html := NewHtml()
        html.Data = data
        html.Content = T.Include(name, html)
        t.Execute(r, html)
    } else {
        panic(r.Layout + " template not found")
    }
}

func (r *Response) RenderPartial(name string, data interface{}) {
    if t := T.Template(name); t != nil {
        t.Execute(r, data)
    } else {
        panic(name + " template not found")
    }
}

func (r *Response) RenderJson(v interface{}) {
    json, e := json.Marshal(v)
    if e != nil {
        L.Println(e)
    }

    r.Header().Set("Content-Type", "application/json; charset=utf8")
    r.Write(json)
}
