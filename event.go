package potato

type IEvent interface {
    AddHandler(name string, handler EventHandler)
    Trigger(name string, args ...interface{})
    ClearHandlers(name string)
    ClearAllHandlers()
}

type EventHandler func(args ...interface{})

type Event struct {
    events map[string][]EventHandler
}

func NewEvent() *Event {
    return &Event{make(map[string][]EventHandler)}
}

func (e *Event) AddHandler(name string, handler EventHandler) {
    handlers, has := e.events[name]
    if !has {
        handlers = make([]EventHandler, 0, 1)
    }
    e.events[name] = append(handlers, handler)
}

func (e *Event) ClearHandlers(name string) {
    delete(e.events, name)
}

func (e *Event) ClearAllHandlers() {
    e.events = make(map[string][]EventHandler)
}

func (e *Event) Trigger(name string, args ...interface{}) {
    if handlers, has := e.events[name]; has && len(handlers) > 0 {
        for _, handler := range handlers {
            handler(args...)
        }
    }
}
