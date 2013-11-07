package potato

import (
    "os"
    "fmt"
    "bytes"
    "strings"
    "html/template"
)

var (
    //2m
    MaxFileSize = int64(2 * 1024 * 1024)
)

type Template struct {
    root *template.Template
    dir string
    funcs template.FuncMap
}

func NewTemplate(dir string) *Template {
    return &Template{
        root: template.New("/"),
        dir: dir,
    }
}

func (t *Template) Template(name string) *template.Template {
    return t.root.Lookup(name)
}

func (t *Template) Include(args ...interface{}) template.HTML {
    name := args[0].(string)
    if tpl := t.root.Lookup(name); tpl != nil {
        var data interface{}
        if len(args) >= 2 {
            data = args[1]
        }

        buffer := new(bytes.Buffer)
        tpl.Execute(buffer, data)
        return template.HTML(buffer.Bytes())
    }

    panic(name + " template not found")
}

func (t *Template) Defined(name string) bool {
    return t.root.Lookup(name) != nil
}

func (t *Template) Funcs(funcs map[string]interface{}) {
    t.funcs = template.FuncMap{
        "include": t.Include,
        "defined": t.Defined,
    }

    for k, f := range funcs {
        t.funcs[k] = f
    }

    t.root.Funcs(t.funcs)
    t.loadTemplateFiles(t.dir)
}

/**
 * loadTemplateFiles loads all *.html files under the dir recursively
 */
func (t *Template) loadTemplateFiles(dir string) {
    d, e := os.Open(dir)
    if e != nil { return }
    defer d.Close()

    dinfo, e := d.Readdir(-1)
    if e != nil { return }

    for _,info := range dinfo {
        uri := dir + info.Name()
        if info.IsDir() {
            t.loadTemplateFiles(uri + "/")

        //check file
        } else if info.Size() <= MaxFileSize &&
                strings.HasSuffix(info.Name(), ".html") {

            //load file
            if f, e := os.Open(uri); e == nil {
                txt := make([]byte, info.Size())
                if _,e := f.Read(txt); e == nil {

                    //init template
                    key := strings.TrimPrefix(
                            strings.TrimSuffix(uri, ".html"), t.dir)
                    template.Must(t.root.New(key).Parse(string(txt)))
                }

                f.Close()
            }
        }
    }
}

type Html struct {
    css, js []string
    title string
    Content template.HTML
    Data interface{}
}

func NewHtml() *Html {
    return &Html{
        css: make([]string, 0),
        js: make([]string, 0),
    }
}

func (h *Html) Title(title string) string {
    if len(title) == 0 {
        return h.title
    }
    h.title = title
    return ""
}

func (h *Html) CSS(uri string) template.HTML {
    if len(uri) == 0 {
        return h.cssHtml()
    }

    h.css = append(h.css, uri)
    return template.HTML("")
}

func (h *Html) cssHtml() template.HTML {
    format := `<link type="text/css" rel="stylesheet" href="%s"/>`
    tags := make([]string, len(h.css))
    for _,uri := range h.css {
        tags = append(tags, fmt.Sprintf(format, uri))
    }
    return template.HTML(strings.Join(tags, "\n"))
}

func (h *Html) JS(uri string) template.HTML {
    if len(uri) == 0 {
        return h.jsHtml()
    }

    h.js = append(h.js, uri)
    return template.HTML("")
}

func (h *Html) jsHtml() template.HTML {
    format := `<script type="text/javascript" src="%s" ></script>`
    tags := make([]string, len(h.js))
    for _,uri := range h.js {
        tags = append(tags, fmt.Sprintf(format, uri))
    }
    return template.HTML(strings.Join(tags, "\n"))
}
