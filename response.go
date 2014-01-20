package potato

import (
    "net/http"
)

type Response struct {
    http.ResponseWriter
    Sent bool
}

func (r *Response) SetCookie(c *http.Cookie) {
    http.SetCookie(r, c)
}
