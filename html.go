package potato

import (
    "bytes"
    "fmt"
    "html/template"
    "os"
    "strings"
)

var (
    //2m
    MaxFileSize = int64(2 * 1024 * 1024)
    TemplateExt = ".html"
)

type Template struct {
    root  *template.Template
    dir   string
    funcs template.FuncMap
}

func NewTemplate(dir string) *Template {
    t := &Template{root: template.New("/"), dir: dir}
    t.funcs = template.FuncMap{
        "potato":  t.Potato,
        "include": t.Include,
        "defined": t.Defined,
        "html":    t.Html,
    }
    return t
}

func (t *Template) Template(name string) *template.Template {
    return t.root.Lookup(name)
}

func (t *Template) Include(args ...interface{}) template.HTML {
    name := args[0].(string)
    if tpl := t.root.Lookup(name); tpl != nil {
        buffer := new(bytes.Buffer)
        n := len(args)

        if n == 1 {
            tpl.Execute(buffer, nil)

            //only one argument
        } else if n == 2 {
            tpl.Execute(buffer, args[1])

            //set all data to a map
            //arguments must be listed as key, value, key, value,...
        } else if n > 2 {
            m := make(map[string]interface{}, n/2)
            for i := 1; i < n-1; i = i + 2 {
                m[args[i].(string)] = args[i+1]
            }

            tpl.Execute(buffer, m)
        }

        return template.HTML(buffer.Bytes())
    }

    panic(name + " template not found")
}

func (t *Template) Defined(name string) bool {
    return t.root.Lookup(name) != nil
}

func (t *Template) Html(str string) template.HTML {
    return template.HTML(str)
}

func (t *Template) Potato() template.HTML {
    return template.HTML(fmt.Sprintf(`<a href="https://github.com/roydong/potato">Potato framework %s</a>`, Version))
}

func (t *Template) AddFuncs(funcs map[string]interface{}) {
    for k, f := range funcs {
        t.funcs[k] = f
    }
    t.root.Funcs(t.funcs)
    t.loadTemplateFiles(t.dir)
}

/**
 * loadTemplateFiles loads all template files under the dir recursively
 */
func (t *Template) loadTemplateFiles(dir string) {
    d, e := os.Open(dir)
    if e != nil {
        return
    }
    defer d.Close()

    dinfo, e := d.Readdir(-1)
    if e != nil {
        return
    }

    for _, info := range dinfo {
        uri := dir + info.Name()
        if info.IsDir() {
            t.loadTemplateFiles(uri + "/")

            //check file
        } else if info.Size() <= MaxFileSize &&
            strings.HasSuffix(info.Name(), TemplateExt) {

            //load file
            if f, e := os.Open(uri); e == nil {
                txt := make([]byte, info.Size())
                if _, e := f.Read(txt); e == nil {

                    //init template
                    key := strings.TrimPrefix(
                        strings.TrimSuffix(uri, TemplateExt), t.dir)
                    template.Must(t.root.New(key).Parse(string(txt)))
                }

                f.Close()
            }
        }
    }
}

type Html struct {
    css, js   []string
    title     string
    fragments map[string]template.HTML
    Data      interface{}
    Content   template.HTML
}

func NewHtml() *Html {
    return &Html{
        css:       make([]string, 0),
        js:        make([]string, 0),
        fragments: make(map[string]template.HTML),
    }
}

func (h *Html) Title(title ...string) string {
    if len(title) == 0 {
        return h.title
    }

    h.title = title[0]
    return ""
}

func (h *Html) F(args ...string) template.HTML {
    if len(args) == 0 {
        return template.HTML("")
    }

    if len(args) == 1 {
        return h.fragments[args[0]]
    }

    h.fragments[args[0]] = template.HTML(args[1])
    return template.HTML("")
}

func (h *Html) CSS(urls ...string) template.HTML {
    if len(urls) == 0 {
        return h.cssHtml()
    }

    for _, url := range urls {
        h.css = append(h.css, url)
    }

    return template.HTML("")
}

func (h *Html) cssHtml() template.HTML {
    format := `<link type="text/css" rel="stylesheet" href="%s"/>`
    tags := make([]string, len(h.css))
    for _, uri := range h.css {
        tags = append(tags, fmt.Sprintf(format, uri))
    }
    return template.HTML(strings.Join(tags, "\n"))
}

func (h *Html) JS(urls ...string) template.HTML {
    if len(urls) == 0 {
        return h.jsHtml()
    }

    for _, url := range urls {
        h.js = append(h.js, url)
    }

    return template.HTML("")
}

func (h *Html) jsHtml() template.HTML {
    format := `<script charset="utf-8" src="%s"></script>`
    tags := make([]string, len(h.js))
    for _, uri := range h.js {
        tags = append(tags, fmt.Sprintf(format, uri))
    }
    return template.HTML(strings.Join(tags, "\n"))
}
