package potato

import (
    "net/http"
    _"encoding/json"
)

type Response struct {
    http.ResponseWriter
    body []byte
}

func (r *Response) SetBody(b []byte) {
    r.body = b
}

func (r *Response) Send() {
    r.Write(r.body)
}

func (r *Response) SetCookie(c *http.Cookie) {
    http.SetCookie(r, c)
}
