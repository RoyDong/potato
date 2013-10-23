package potato

import (
    "net/http"
    _"encoding/json"
)

type Response struct {
    http.ResponseWriter
}
