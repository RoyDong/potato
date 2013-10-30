package potato

import (
    "strings"
)

type Tree struct {
    data map[interface{}]interface{}
}

func NewTree(data map[interface{}]interface{}) *Tree {
    return &Tree{data}
}

/**
 * find finds target data on the tree by provided nodes
 */
func (t *Tree) find(nodes []string) (map[interface{}]interface{}, bool) {
    var ok bool
    data := t.data
    for _,n := range nodes {
        if data, ok = data[n].(map[interface{}]interface{}); !ok {
            return nil, false
        }
    }

    return data, true
}

/**
 * Mount adds new value on the tree
 * f means force to replace old value if there is any
 */
func (t *Tree) Mount(path string, v interface{}, f bool) bool {
    var i int
    var n string
    nodes := strings.Split(path , ".")
    last := len(nodes) - 1
    data := t.data

    //get to the last existing node on the tree of the path
    for i, n = range nodes[:last] {
        if d, ok := data[n].(map[interface{}]interface{}); ok {
            data = d
        } else {
            break
        }
    }

    //the next node is a value and f is false
    if _,has := data[n]; has && !f {
        return false
    }

    //if the loop upove is ended by break
    //then create the rest nodes of the path
    if i < last - 1 {
        for _,n = range nodes[i:last] {
            d := make(map[interface{}]interface{}, 1)
            data[n] = d
            data = d
        }
    }

    if t, ok := v.(*Tree); ok {
        v = t.data
    }

    data[nodes[last]] = v
    return true
}

/**
 * Value returns the data found by path
 * path is a string with node names divided by dot(.)
 */
func (t *Tree) Value(path string) (interface{}, bool) {
    nodes := strings.Split(path , ".")
    l := len(nodes) - 1
    if data, ok := t.find(nodes[:l]); ok {
        v, has := data[nodes[l]]
        return v, has
    }

    return nil, false
}

/**
 * Sub returns a *Tree object stores the data found by path
 */
func (t *Tree) Sub(path string) (*Tree, bool) {
    if data, ok := t.find(strings.Split(path, ".")); ok {
        return &Tree{data}, true
    }

    return nil, false
}

func (t *Tree) Clear() {
    t.data = make(map[interface{}]interface{})
}

func (t *Tree) Load(data map[interface{}]interface{}) {
    t.data = data
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

