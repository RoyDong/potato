package potato

import (
    "crypto/md5"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "time"
)

var (
    SessionDomain     string
    SessionDuration   = int64(60 * 60 * 24)
    SessionCookieName = "POTATO_SESSION_ID"
    sessions          = make(map[string]*Session)
)

type Session struct {
    Tree
    id        string
    UpdatedAt time.Time
}

func NewSession(r *Request, p *Response) *Session {
    s := &Session{
        Tree:      *NewTree(nil),
        id:        sessionId(r),
        UpdatedAt: time.Now(),
    }

    //set id in cookie
    sessions[s.id] = s
    p.SetCookie(&http.Cookie{
        Name:     SessionCookieName,
        Value:    s.id,
        Path:     "/",
        Domain:   SessionDomain,
        HttpOnly: true})

    return s
}

/**
 * InitSession gets current session by session id in cookie
 * if none creates a new session
 */
func InitSession(r *Request, p *Response) {
    if c := r.Cookie(SessionCookieName); c != nil {
        var has bool
        if r.Session, has = sessions[c.Value]; has {
            sec := time.Now().Unix()
            if has && r.Session.UpdatedAt.Unix()+SessionDuration < sec {
                delete(sessions, r.Session.id)
                r.Session.Clear()
                r.Session = nil
            }
        }
    }
    if r.Session == nil {
        r.Session = NewSession(r, p)
    }
}

func sessionId(r *Request) string {
    rnd := make([]byte, 24)
    if _, e := io.ReadFull(rand.Reader, rnd); e != nil {
        panic("potato: session id " + e.Error())
    }

    sig := fmt.Sprintf("%s%d%s", r.RemoteAddr, time.Now().UnixNano(), rnd)
    hash := md5.New()
    if _, e := hash.Write([]byte(sig)); e != nil {
        panic("potato: session id " + e.Error())
    }

    return hex.EncodeToString(hash.Sum(nil))
}

/**
 * sessionExpire checks sessons expiration per minute
 * and delete all expired sessions
 */
func sessionExpire() {
    for now := range time.Tick(time.Minute) {
        t := now.Unix()
        for k, s := range sessions {
            if s.UpdatedAt.Unix()+SessionDuration < t {
                s.Clear()
                delete(sessions, k)
            }
        }
    }
}
