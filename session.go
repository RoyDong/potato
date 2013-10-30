package potato

import (
    "io"
    "fmt"
    "time"
    "net/http"
    "crypto/rand"
    "encoding/hex"
    "crypto/md5"
)

var (
    SessionDir = "session/"
    SessionDuration = int64(10)
    SessionCookieName = "POTATO_SESSION_ID"
    sessions = make(map[string]*Session)
)

type Session struct {
    *Tree
    Id string
    LastActivity int64
}

func SessionStart() {
    go checkSessionExpire()
}

func NewSession(r *Request, p *Response) *Session {
    s := &Session{
        Tree: NewTree(make(map[interface{}]interface{})),
        Id: createSessionId(r),
        LastActivity: time.Now().Unix(),
    }

    sessions[s.Id] = s

    //set id to cookie
    p.SetCookie(&http.Cookie{
        Name: SessionCookieName,
        Value: s.Id,
    })

    return s
}

/**
 * InitSession gets current session by session id in cookie
 * if none creates a new session
 */
func InitSession(r *Request, p *Response) {
    if c := r.Cookie(SessionCookieName); c != nil {
        r.Session = sessions[c.Value]
    }

    if r.Session == nil {
        r.Session  = NewSession(r, p)
    } else {
        t := time.Now().Unix()

        //check session expire time
        if r.Session.LastActivity + SessionDuration < t {
            r.Session.Clear()
        }

        r.Session.LastActivity = t
    }
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

func checkSessionExpire() {
    for now := range time.Tick(time.Minute) {
        t := now.Unix()
        for k, s := range sessions {
            if s.LastActivity + SessionDuration < t {
                s.Clear()
                delete(sessions, k)
            }
        }
    }
}

