package potato

import (
    "os"
    "bytes"
    "strings"
    "text/template"
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

func (h *Template) Template(name string) *template.Template {
    return h.root.Lookup(name)
}

func (h *Template) Include(name string, data interface{}) string {
    if t := h.root.Lookup(name); t != nil {
        buffer := new(bytes.Buffer)
        t.Execute(buffer, data)
        return string(buffer.Bytes())
    }

    panic(name + " template not found")
}

func (h *Template) Funcs(funcs map[string]interface{}) {
    h.funcs = template.FuncMap{
        "include": h.Include,
    }

    for k, f := range funcs {
        h.funcs[k] = f
    }

    h.loadTemplateFiles(h.dir)
}

/**
 * loadTemplateFiles loads all *.html files under the dir recursively
 */
func (h *Template) loadTemplateFiles(dir string) {
    d, e := os.Open(dir)
    if e != nil { return }

    dinfo, e := d.Readdir(-1)
    if e != nil { return }

    for _,info := range dinfo {
        uri := dir + info.Name()
        if info.IsDir() {
            h.loadTemplateFiles(uri + "/")

        //check file
        } else if info.Size() <= MaxFileSize &&
                strings.HasSuffix(info.Name(), ".html") {

            //load file
            if f, e := os.Open(uri); e == nil {
                str := make([]byte, info.Size())
                if _,e := f.Read(str); e == nil {

                    //init template
                    key := strings.TrimPrefix(strings.TrimSuffix(uri, ".html"), h.dir)
                    template.Must(h.root.New(key).Funcs(h.funcs).Parse(string(str)))
                }
            }
        }
    }
}

type Html struct {
    Js, Css []string
    Title, Content string
}


