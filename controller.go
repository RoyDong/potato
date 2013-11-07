package potato

import (
    "net/http"
    "encoding/json"
)

type Controller struct {
    Request *Request
    Response *Response

    Title string
    Layout string
    Template string
    Data interface{}
}

func NewController(r *Request, p *Response) *Controller {
    return &Controller{
        Request: r,
        Response: p,
        Layout: "layout",
        Template: "index",
        Title: AppName,
    }
}

func (c *Controller) Redirect(url string) {
    http.Redirect(c.Response, c.Request.Request, url, http.StatusFound)
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

func (c *Controller) RenderJson(v interface{}) {
    json,_ := json.Marshal(v)
    c.Response.Write(json)
}


