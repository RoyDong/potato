package potato

import (
    "fmt"
)

type Error struct {
    Code int
    Message string
}

func Panic(c int, m string) {
    panic(&Error{c, m})
}

func (e *Error) String() string {
    return fmt.Sprintf("%d %s", e.Code, e.Message)
}
