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
    SessionDuration = int64(60 * 60 * 24)
    SessionCookieName = "POTATO_SESSION_ID"
    sessions = make(map[string]*Session)
)

type Session struct {
    Tree
    Id string
    LastActivity time.Time
}

func SessionStart() {
    go checkSessionExpiration()
}

func NewSession(r *Request, p *Response) *Session {
    s := &Session{
        Tree: Tree{make(map[interface{}]interface{})},
        Id: createSessionId(r),
        LastActivity: time.Now(),
    }

    sessions[s.Id] = s

    //set id in cookie
    p.SetCookie(&http.Cookie{
        Name: SessionCookieName,
        Value: s.Id,
        Path: "/",
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
        t := time.Now()

        //check session expire time
        if r.Session.LastActivity.Unix() + SessionDuration < t.Unix() {
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

/**
 * checkSessionExpiration checks sessons expiration per minute 
 * and delete all expired sessions
 */
func checkSessionExpiration() {
    for now := range time.Tick(time.Minute) {
        t := now.Unix()
        for k, s := range sessions {
            if s.LastActivity.Unix() + SessionDuration < t {
                s.Clear()
                delete(sessions, k)
            }
        }
    }
}

