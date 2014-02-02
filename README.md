example: https://github.com/RoyDong/notes

modify go.net/websocket/server.go add a function:

    func (s Server) Conn(w http.ResponseWriter, req *http.Request) *Conn {
        rwc, buf, err := w.(http.Hijacker).Hijack()
        if err != nil {
            return nil
        }
        conn, err := newServerConn(rwc, buf, req, &s.Config, s.Handshake)
        if err != nil {
            return nil
        }
        if conn == nil {
            return nil
        }
        return conn
    }



PACKAGE DOCUMENTATION

package potato
    import "."



CONSTANTS

const (
    Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

const (
    TerminateCode = 0
)


VARIABLES

var (
    AppName  string
    Version  string
    SockFile string
    Conf     *Tree
    Logger   *log.Logger
    ConfDir  = "config/"
    TplDir   = "template/"
    LogDir   = "log/"
    Env      = "prod"
    Port     = 37221
)

var (
    SessionDomain     string
    SessionDuration   = int64(60 * 60 * 24)
    SessionCookieName = "POTATO_SESSION_ID"
)

var (
    //2m
    MaxFileSize = int64(2 * 1024 * 1024)
    TemplateExt = ".html"
)

var DefaultLayout = "layout"

var ErrorAction = func(r *Request, p *Response) {
    msg, _ := r.Bag.String("error")
    p.WriteHeader(500)
    p.Write([]byte("we'v got some error " + msg))
}

var NotfoundAction = func(r *Request, p *Response) {
    p.WriteHeader(404)
    p.Write([]byte("page not found"))
}


FUNCTIONS

func AddEventHandler(name string, handler EventHandler)

func Init()

func InitSession(r *Request, p *Response)
    *

	* InitSession gets current session by session id in cookie
	* if none creates a new session

func LoadFile(filename string) ([]byte, error)

func LoadJson(v interface{}, filename string) error

func LoadYaml(v interface{}, filename string) error

func RandString(length int) string

func Serve()

func SetAction(action Action, patterns ...string)

func TemplateFuncs(funcs map[string]interface{})


TYPES

type Action func(r *Request, p *Response)



type Event struct {
    // contains filtered or unexported fields
}


func NewEvent() *Event


func (e *Event) AddEventHandler(name string, handler EventHandler)

func (e *Event) ClearAllEventHandlers()

func (e *Event) ClearEventHandlers(name string)

func (e *Event) TriggerEvent(name string, args ...interface{})


type EventHandler func(args ...interface{})



type Html struct {
    Data    interface{}
    Content template.HTML
    // contains filtered or unexported fields
}


func NewHtml() *Html


func (h *Html) CSS(urls ...string) template.HTML

func (h *Html) F(args ...string) template.HTML

func (h *Html) JS(urls ...string) template.HTML

func (h *Html) Title(title ...string) string


type IEvent interface {
    AddEventHandler(name string, handler EventHandler)
    TriggerEvent(name string, args ...interface{})
    ClearEventHandlers(name string)
    ClearAllEventHandlers()
}



type Request struct {
    *http.Request
    WSConn *ws.Conn

    Session *Session
    Cookies []*http.Cookie
    Bag     *Tree
    // contains filtered or unexported fields
}


func NewRequest(r *http.Request) *Request


func (r *Request) Cookie(name string) *http.Cookie

func (r *Request) Float(k string) (float64, bool)

func (r *Request) Int(k string) (int, bool)

func (r *Request) Int64(k string) (int64, bool)

func (r *Request) IsAjax() bool

func (r *Request) String(k string) (string, bool)

func (r *Request) WSReceive() string

func (r *Request) WSSend(txt string) bool

func (r *Request) WSSendJson(v interface{}) bool


type Response struct {
    http.ResponseWriter
    Layout string
}


func NewResponse(w http.ResponseWriter) *Response


func (r *Response) Redirect(request *Request, url string, code int)

func (r *Response) Render(name string, data interface{})

func (r *Response) RenderJson(v interface{})

func (r *Response) RenderPartial(name string, data interface{})

func (r *Response) RenderText(t string)

func (r *Response) SetCookie(c *http.Cookie)


type Route struct {
    // contains filtered or unexported fields
}


func (r *Route) Action() Action

func (r *Route) Name() string

func (r *Route) Parse(path string) (*Route, []string)

func (r *Route) Set(path string, action Action)


type Session struct {
    Tree
    Id        string
    UpdatedAt time.Time
}


func NewSession(r *Request, p *Response) *Session



type Template struct {
    // contains filtered or unexported fields
}


func NewTemplate(dir string) *Template


func (t *Template) AddFuncs(funcs map[string]interface{})

func (t *Template) Defined(name string) bool

func (t *Template) Html(str string) template.HTML

func (t *Template) Include(args ...interface{}) template.HTML

func (t *Template) Potato() template.HTML

func (t *Template) Template(name string) *template.Template


type Tree struct {
    // contains filtered or unexported fields
}


func NewTree(data map[interface{}]interface{}) *Tree


func (t *Tree) Clear()

func (t *Tree) Float64(path string) (float64, bool)

func (t *Tree) Get(path string) interface{}
    *

	* Value returns the data found by path
	* path is a string with node names divided by dot(.)

func (t *Tree) Int(path string) (int, bool)

func (t *Tree) Int64(path string) (int64, bool)

func (t *Tree) Set(path string, v interface{}, f bool) bool
    *

	* Set adds new value on the tree
	* f means force to replace old value if there is any

func (t *Tree) String(path string) (string, bool)

func (t *Tree) Tree(path string) (*Tree, bool)
    *

	* Sub returns a *Tree object stores the data found by path



SUBDIRECTORIES

	orm

