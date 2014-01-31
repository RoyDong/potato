package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "encoding/json"
    "net/http"
    "reflect"
)

type Controller struct {
    Request  *Request
    Response *Response
    Layout   string
}

func NewController(t reflect.Type, r *Request, p *Response) reflect.Value {
    c := reflect.New(t)
    c.Elem().FieldByName("Controller").
        Set(reflect.ValueOf(Controller{
        Request:  r,
        Response: p,
        Layout:   "layout"}))
    return c
}

func (c *Controller) Redirect(url string, code int) {
    http.Redirect(c.Response, c.Request.Request, url, code)
    panic(CodeTerminate)
}

func (c *Controller) RenderText(t string) {
    c.Response.Write([]byte(t))
    c.Response.Sent = true
}

func (c *Controller) Render(name string, data interface{}) {
    if t := T.Template(c.Layout); t != nil {
        html := NewHtml()
        html.Data = data
        html.Content = T.Include(name, html)
        t.Execute(c.Response, html)
        c.Response.Sent = true
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

    c.Response.Header().Set("Content-Type", "application/json; charset=utf8")
    c.Response.Write(json)
    c.Response.Sent = true
}

func (c *Controller) WSReceive() string {
    var txt string
    if e := ws.Message.Receive(c.Request.WSConn, &txt); e != nil {
        L.Println(e)
        return ""
    }

    return txt
}

func (c *Controller) WSSend(txt string) bool {
    if e := ws.Message.Send(c.Request.WSConn, txt); e != nil {
        L.Println(e)
        return false
    }

    return true
}

func (c *Controller) WSSendJson(v interface{}) bool {
    if e := ws.JSON.Send(c.Request.WSConn, v); e != nil {
        L.Println(e)
        return false
    }

    return true
}
