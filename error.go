package potato

import (
    "fmt"
)

const (
    RedirectCode = 302
)

type Error struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func Panic(c int, m string) {
    panic(&Error{c, m})
}

func (e *Error) String() string {
    return fmt.Sprintf("%d %s", e.Code, e.Message)
}
