package potato

import (
    "strings"
    "fmt"
    "sync"
)

type Tree struct {
    name string
    branches map[string]*Tree
    value interface{}
    locker *sync.Mutex
}

func NewTree() *Tree {
    return &Tree{locker: &sync.Mutex{}}
}

/*
LoadYaml loads data read from a yaml file,
repl means whether to replace or keep the old value
*/
func (t *Tree) LoadYaml(file string, repl bool) error {
    var yml interface{}
    if e := LoadYaml(&yml, file); e != nil {
        return e
    }
    data2Tree(t, yml, repl)
    return nil
}

/*
LoadJson loads data read from a json file,
repl means whether to replace or keep the old value
*/
func (t *Tree) LoadJson(file string, repl bool) error {
    var json interface{}
    if e := LoadJson(&json, file); e != nil {
        return e
    }
    data2Tree(t, json, repl)
    return nil
}

func data2Tree(t *Tree, d interface{}, repl bool) {
    if v, ok := d.(map[interface{}]interface{}); ok {
        map2Tree(t, v, repl)
    } else if v, ok := d.([]interface{}); ok {
        array2Tree(t, v, repl)
    } else if repl || t.value == nil {
        t.value = d
    }
}

func map2Tree(t *Tree, d map[interface{}]interface{}, repl bool) {
    if t.branches == nil {
        t.branches = make(map[string]*Tree)
    }
    for k, v := range d {
        var key string
        if v, ok := k.(string); ok {
            key = v
        } else if v, ok := k.(int); ok {
            key = fmt.Sprintf("%d", v)
        } else {
            continue
        }
        tree, has := t.branches[key]
        if !has {
            tree = &Tree{name: key}
            t.branches[key] = tree
        }
        data2Tree(tree, v, repl)
    }
}

func array2Tree(t *Tree, d []interface{}, repl bool) {
    if t.branches == nil {
        t.branches = make(map[string]*Tree)
    }
    for i, v := range d {
        key := fmt.Sprintf("%d", i)
        tree, has := t.branches[key]
        if !has {
            tree = &Tree{name: key}
            t.branches[key] = tree
        }
        data2Tree(tree, v, repl)
    }
}

func (t *Tree) find(key string) *Tree {
    if key == "" {
        return t
    }
    current := t
    nodes := strings.Split(
        strings.ToLower(strings.Trim(key, ".")), ".")
    for _, name := range nodes {
        var has bool
        if current, has = current.branches[name]; !has {
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
    if key == "" {
        return t
    }
    current := t
    nodes := strings.Split(
        strings.ToLower(strings.Trim(key, ".")), ".")
    for _, name := range nodes {
        var tree *Tree
        var has bool
        if current.branches == nil {
            current.branches = make(map[string]*Tree)
            has = false
        } else {
            tree, has = current.branches[name]
        }
        if !has {
            tree = &Tree{name: name}
            current.branches[name] = tree
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
    t.value = nil
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
        if i, ok := v.(int64); ok {
            return v, true
        }
        if i, ok := v.(int); ok {
            return int64(i), ok
        }
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
