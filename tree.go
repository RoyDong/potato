package potato

import (
    "strings"
)

type Tree struct {
    Data map[interface{}]interface{}
}

func (t *Tree) find(nodes []string) (map[interface{}]interface{}, bool) {
    var ok bool
    data := t.Data
    for _,n := range nodes {
        if data, ok = data[n].(map[interface{}]interface{}); !ok {
            return nil, false
        }
    }

    return data, true
}

func (t *Tree) Node(path string) (*Tree, bool) {
    if data, ok := t.find(strings.Split(path, ".")); ok {
        return &Tree{data}, true
    }

    return nil, false
}

func (t *Tree) Value(path string) (interface{}, bool) {
    nodes := strings.Split(path , ".")
    l := len(nodes) - 1
    if data, ok := t.find(nodes[:l]); ok {
        v, has := data[nodes[l]]
        return v, has
    }

    return nil, false
}

func (t *Tree) Int(path string) (int, bool) {
    if v, ok := t.Value(path); ok {
        i, ok := v.(int)
        return i, ok
    }

    return 0, false
}

func (t *Tree) Int64(path string) (int64, bool) {
    if v, ok := t.Value(path); ok {
        i, ok := v.(int64)
        return i, ok
    }

    return 0, false
}

func (t *Tree) Float64(path string) (float64, bool) {
    if v, ok := t.Value(path); ok {
        f, ok := v.(float64)
        return f, ok
    }

    return 0, false
}

func (t *Tree) String(path string) (string, bool) {
    if v, ok := t.Value(path); ok {
        s, ok := v.(string)
        return s, ok
    }

    return "", false
}
