package potato

import (
    "io"
    "fmt"
    "time"
    "crypto/rand"
    "encoding/hex"
    "crypto/md5"
)

var (
    SessionDuration = int64(10)
    SessionCookieName = "POTATO_SESSION_ID"
    sessions = make(map[string]*Session)
)


type Session struct {
    *Tree
    Id string
    LastActivity int64
}

func NewSession(r *Request) *Session {
    s := &Session{
        Tree: NewTree(make(map[string]interface{})),
        Id: createSessionId(r),
        LastActivity: time.Now().Unix(),
    }

    sessions[s.Id] = s
    return s
}


func createSessionId(r *Request) string {
    rnd := make([]byte, 24)
    if _,e := io.ReadFull(rand.Reader, rnd); e != nil {
        panic("could not get random chars while creating session id")
    }

    sig := fmt.Sprintf("%s%d%s", r.RemoteAddr, time.Now().UnixNano(), rnd)
    hash := md5.New()
    if _,e := hash.Write([]byte(sig)); e != nil {
        panic("could not hash string while creating session id")
    }

    return hex.EncodeToString(hash.Sum(nil))
}

