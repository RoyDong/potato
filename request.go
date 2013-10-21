package potato


import (
    "strconv"
    "net/http"
)

type Request struct {
    *http.Request
    Route *Route
    params map[string]string
}

func (r *Request) GetInt(k string) (int64, bool) {
    if v, has := r.Get(k); has {
        if i, e := strconv.ParseInt(v, 10, 64); e == nil {
            return i, true
        }
    }

    return 0, false
}

func (r *Request) GetFloat(k string) (float64, bool) {
    if v, has := r.Get(k); has {
        if f, e := strconv.ParseFloat(v, 64); e == nil {
            return f, true
        }
    }

    return 0, false
}

func (r *Request) Get(k string) (string, bool) {
    if v, has := r.params[k]; has {
        return v, true
    }

    if v := r.FormValue(k); len(v) > 0 {
        return v, true
    }

    return "", false
}

