package potato

import (
    "os"
    "bytes"
    "regexp"
    "strings"
    "math/rand"
    "encoding/json"
    "launchpad.net/goyaml"
)

const (
    Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func RandString(length int) string {
    rnd := make([]byte, 0, length)
    max := len(Chars) - 1

    for i := 0; i < length; i++ {
        rnd = append(rnd, Chars[rand.Intn(max)])
    }

    return string(rnd)
}

func ParseHttpQuery(q string) map[string]string {
    parts := strings.Split(string(q), "&")
    params := make(map[string]string, len(parts))

    for _, v := range parts {
        p := strings.Split(v, "=")

        if len(p) == 2 {
            params[p[0]] = p[1]
        }
    }

    return params
}

func LoadJson(v interface{}, filename string) error {
    text, e := LoadFile(filename)
    if e != nil {
        return e
    }

    rows := bytes.Split(text, []byte("\n"))
    r := regexp.MustCompile(`^\s*[/#]+`)
    for i, row := range rows {
        if r.Match(row) {
            rows[i] = nil
        }
    }

    return json.Unmarshal(bytes.Join(rows, nil), v)
}

func LoadYaml(v interface{}, filename string) error {
    text, e := LoadFile(filename)
    if e != nil {
        return e
    }

    return goyaml.Unmarshal(text, v)
}

func LoadFile(filename string) ([]byte, error) {
    file, e := os.Open(filename)
    if e != nil {
        return nil, e
    }
    defer file.Close()

    fileInfo, e := file.Stat()
    if e != nil {
        return nil, e
    }

    text := make([]byte, fileInfo.Size())
    file.Read(text)
    return text, nil
}
