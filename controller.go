package potato

import (
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

func NewController(rq *Request, rp *Response) *Controller {
    return &Controller{
        Request: rq,
        Response: rp,
        Layout: "layout",
        Template: "index",
        Title: Host,
    }
}

func (c *Controller) Render(name string, data interface{}) {
    if t := H.Template(c.Layout); t != nil {
        c.Template = name
        c.Data = data
        t.Execute(c.Response, c)
    } else {
        panic("layout template not found")
    }
}

func (c *Controller) RenderJson(v interface{}) {
    json,_ := json.Marshal(v)
    c.Response.Write(json)
}
