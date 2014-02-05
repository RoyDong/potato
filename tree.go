package potato

import (
    "strings"
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

func (t *Tree) LoadYaml(f string) error {
    var data map[interface{}]interface{}
    if e := LoadYaml(&data, f); e != nil {
        return e
    }
    t.load(t, data)
    return nil
}

func (t *Tree) load(p *Tree, data map[interface{}]interface{}) {
    if p.branches == nil {
        p.branches = make([]*Tree, 0)
    }
    for k, v := range data {
        tree := &Tree{name: k.(string)}
        p.branches = append(p.branches, tree)
        if d, ok := v.(map[interface{}]interface{}); ok {
            t.load(tree, d)
        } else {
            tree.value = v
        }
    }
}

func (t *Tree) find(key string) *Tree {
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
