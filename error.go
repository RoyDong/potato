package potato

import (
    "fmt"
)


/*
those status codes are recomended to use for NewError function
*/
const (
    StatusBadRequest              = 400
    StatusPaymentRequired         = 402
    StatusForbidden               = 403
    StatusNotFound                = 404
    StatusUnsupportedMediaType    = 415
    StatusInternalServerError     = 500
    StatusServiceUnavailable      = 503
    StatusGatewayTimeout          = 504
    StatusHTTPVersionNotSupported = 505
)

type Error struct {
    code int
    message string
}

func NewError(c int, m string) *Error {
    return &Error{c, m}
}

func (e *Error) Code() int {
    return e.code
}

func (e *Error) Message() string {
    return e.message
}

func (e *Error) Error() string {
    return fmt.Sprintf("code: %d, message: %s", e.code, e.message)
}

func (e *Error) Json() string {
    return fmt.Sprintf("{code: %d, message: %s}", e.code, e.message)
}
