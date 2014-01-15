package potato

type EventHandler func(args ...interface{})

type IEvent interface {
    AddEventHandler(name string, handler EventHandler)
    ClearEventHandlers(name string)
    TriggerEvent(name string, args ...interface{})
}

type Event struct {
    events map[string][]EventHandler
}

func NewEvent() *Event {
    return &Event{
        events: make(map[string][]EventHandler),
    }
}

func (e *Event) AddEventHandler(name string, handler EventHandler) {
    handlers, has := e.events[name]
    if !has {
        handlers = make([]EventHandler, 0, 1)
    }
    e.events[name] = append(handlers, handler)
}

func (e *Event) ClearEventHandlers(name string) {
    delete(e.events, name)
}

func (e *Event) ClearAllEventHandlers() {
    e.events = make(map[string][]EventHandler)
}

func (e *Event) TriggerEvent(name string, args ...interface{}) {
    if handlers, has := e.events[name]; has && len(handlers) > 0 {
        for _, handler := range handlers {
            handler(args...)
        }
    }
}
