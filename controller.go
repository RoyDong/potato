package potato

import (
    "net/http"
)

type Controller struct {
    Request *Request
    RW http.ResponseWriter
}
