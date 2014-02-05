package potato

import (
    "strings"
    "fmt"
    "sync"
)

type Tree struct {
    name string
    branches []*Tree
    value interface{}
    locker *sync.Mutex
}

func NewTree() *Tree {
    return &Tree{locker: &sync.Mutex{}}
}

/**
 * LoadYaml loads data written in yaml file
 * repl means whether to replace or keep the old value
 */
func (t *Tree) LoadYaml(file string, repl bool) error {
    var yml interface{}
    if e := LoadYaml(&yml, file); e != nil {
        return e
    }
    load2Tree(t, yml, repl)
    return nil
}

/**
 * LoadJson loads data written in yaml file
 * repl means whether to replace or keep the old value
 */
func (t *Tree) LoadJson(file string, repl bool) error {
    var json interface{}
    if e := LoadJson(&json, file); e != nil {
        return e
    }
    load2Tree(t, json, repl)
    return nil
}

func load2Tree(t *Tree, d interface{}, repl bool) {
    if v, ok := d.(map[interface{}]interface{}); ok {
        map2Tree(t, v, repl)
    } else if v, ok := d.([]interface{}); ok {
        arr2Tree(t, v, repl)
    } else if repl || t.value == nil {
        t.value = d
    }
}

func map2Tree(t *Tree, d map[interface{}]interface{}, repl bool) {
    if t.branches == nil {
        t.branches = make([]*Tree, 0)
    }
    for k, v := range d {
        value2Tree(t, k.(string), v, repl)
    }
}

func arr2Tree(t *Tree, d []interface{}, repl bool) {
    if t.branches == nil {
        t.branches = make([]*Tree, 0)
    }
    for i, v := range d {
        value2Tree(t, fmt.Sprintf("%d", i) , v, repl)
    }
}

func value2Tree(t *Tree, k string, v interface{}, repl bool) {
    var tree *Tree
    for _, b := range t.branches {
        if b.name == k {
            tree = b
            break
        }
    }
    if tree == nil {
        tree = &Tree{name: k}
    }
    t.branches = append(t.branches, tree)
    load2Tree(tree, v, repl)
}

func (t *Tree) find(key string) *Tree {
    if key == "" {
        return t
    }
    current := t
    nodes := strings.Split(
        strings.ToLower(strings.Trim(key, ".")), ".")
    for _, name := range nodes {
        found := false
        for _, tree := range current.branches {
            if name == tree.name {
                found = true
                current = tree
                break
            }
        }
        if !found {
            return nil
        }
    }
    return current
}

func (t *Tree) Get(key string) interface{} {
    tree := t.find(key)
    if tree == nil {
        return nil
    }
    return tree.value
}

func (t *Tree) prepare(key string) *Tree {
    current := t
    nodes := strings.Split(
        strings.ToLower(strings.Trim(key, ".")), ".")
    for _, name := range nodes {
        var found bool
        var tree *Tree
        for _, tree = range current.branches {
            if name == tree.name {
                found = true
                current = tree
                break
            }
        }
        if !found {
            tree = &Tree{name: name}
            if current.branches == nil {
                current.branches = make([]*Tree, 0)
            }
            current.branches = append(current.branches, tree)
        }
        current = tree
    }
    return current
}

func (t *Tree) Set(key string, val interface{}) {
    t.locker.Lock()
    defer t.locker.Unlock()
    tree := t.prepare(key)
    tree.value = val
}

func (t *Tree) Add(key string, val interface{}) bool {
    t.locker.Lock()
    defer t.locker.Unlock()
    if tree := t.prepare(key); tree.value == nil {
        tree.value = val
        return true
    }
    return false
}

func (t *Tree) Tree(key string) *Tree {
    tree := t.find(key)
    if tree == nil {
        return nil
    }
    tree.locker = t.locker
    return tree
}

func (t *Tree) Clear() {
    t.branches = nil
}

func (t *Tree) Int(key string) (int, bool) {
    if v := t.Get(key); v != nil {
        i, ok := v.(int)
        return i, ok
    }
    return 0, false
}

func (t *Tree) Int64(key string) (int64, bool) {
    if v := t.Get(key); v != nil {
        i, ok := v.(int64)
        return i, ok
    }
    return 0, false
}

func (t *Tree) Float64(key string) (float64, bool) {
    if v := t.Get(key); v != nil {
        f, ok := v.(float64)
        return f, ok
    }
    return 0, false
}

func (t *Tree) String(key string) (string, bool) {
    if v := t.Get(key); v != nil {
        s, ok := v.(string)
        return s, ok
    }
    return "", false
}
