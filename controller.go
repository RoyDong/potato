package potato

import (
    "net/http"
    "encoding/json"
)

type Controller struct {
    Request *Request
    Response *Response
    Layout string
}

func NewController(r *Request, p *Response) *Controller {
    return &Controller{
        Request: r,
        Response: p,
        Layout: "layout",
    }
}

func (c *Controller) Redirect(url string) {
    http.Redirect(c.Response, c.Request.Request, url, http.StatusFound)
    Panic(RedirectCode, "redirect is not an error")
}

func (c *Controller) Render(name string, data interface{}) {
    if t := T.Template(c.Layout); t != nil {
        html := NewHtml()
        html.Data = data
        html.Content = T.Include(name, html)
        t.Execute(c.Response, html)
    } else {
        panic(c.Layout + " template not found")
    }
}

func (c *Controller) RenderPartial(name string, data interface{}) {
    if t := T.Template(name); t != nil {
        t.Execute(c.Response, data)
    } else {
        panic(name + " template not found")
    }
}

func (c *Controller) RenderJson(v interface{}) {
    json, e := json.Marshal(v)
    if e != nil {
        L.Println(e)
    }

    c.Response.Write(json)
}

