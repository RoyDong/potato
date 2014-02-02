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
    Id        string
    UpdatedAt time.Time
}

func NewSession(r *Request, p *Response) *Session {
    s := &Session{
        Tree:      *NewTree(nil),
        Id:        sessionId(r),
        UpdatedAt: time.Now(),
    }

    //set id in cookie
    sessions[s.Id] = s
    p.SetCookie(&http.Cookie{
        Name:     SessionCookieName,
        Value:    s.Id,
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
        r.Session = sessions[c.Value]
    }

    if r.Session == nil {
        r.Session = NewSession(r, p)
    } else {
        t := time.Now()

        //check session expiration
        if r.Session.UpdatedAt.Unix()+SessionDuration < t.Unix() {
            r.Session.Clear()
        }

        r.Session.UpdatedAt = t
    }
}

func sessionId(r *Request) string {
    rnd := make([]byte, 24)
    if _, e := io.ReadFull(rand.Reader, rnd); e != nil {
        panic("could not get random chars while creating session id")
    }

    sig := fmt.Sprintf("%s%d%s", r.RemoteAddr, time.Now().UnixNano(), rnd)
    hash := md5.New()
    if _, e := hash.Write([]byte(sig)); e != nil {
        panic("could not hash string while creating session id")
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
