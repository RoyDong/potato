package potato

import (
    "net/http"
)

type Controller struct {
    Request *http.Request
    Params map[string]string
    RW http.ResponseWriter
}
