package potato

import (
    ws "code.google.com/p/go.net/websocket"
    "fmt"
    "net"
    "net/http"
    "os"
)

func Serve() {
    var e error
    var lsn net.Listener

    if len(SockFile) > 0 {
        os.Remove(SockFile)
        lsn, e = net.Listen("unix", SockFile)
        if e != nil {
            L.Println("fail to open socket file", e)
        } else {
            os.Chmod(SockFile, os.ModePerm)
        }
    }

    if lsn == nil {
        lsn, e = net.Listen("tcp", fmt.Sprintf(":%d", Port))
    }

    if e != nil {
        L.Fatal(e)
    }

    fmt.Println("work work")
    s := &http.Server{Handler: &Router{ws.Server{}}}
    L.Println(s.Serve(lsn))
    lsn.Close()
}
