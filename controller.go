package potato

import (
    "encoding/json"
)

type Controller struct {
    Request *Request
    Response *Response
}


func (c *Controller) Render(name string, v interface{}) {
    if t := H.Template(name); t != nil {
        t.Execute(c.Response, v)
    } else {
        panic(name + " template not found")
    }
}

func (c *Controller) RenderJson(v interface{}) {
    json,_ := json.Marshal(v)
    c.Response.SetBody(json)
}
