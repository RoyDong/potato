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

type Html struct {
    root *template.Template
    dir string
    funcs template.FuncMap
}

func NewHtml(dir string) *Html {
    return &Html{
        root: template.New("/"),
        dir: dir,
    }
}

func (h *Html) Template(name string) *template.Template {
    return h.root.Lookup(name)
}

func (h *Html) Include(args ...interface{}) string {
    name := args[0].(string)
    if t := h.root.Lookup(name); t != nil {
        var data interface{}
        if len(args) >= 2 {
            data = args[1]
        }

        buffer := new(bytes.Buffer)
        t.Execute(buffer, data)
        return string(buffer.Bytes())
    }

    panic(name + " template not found")
}

func (h *Html) Defined(name string) bool {
    return h.root.Lookup(name) != nil
}

func (h *Html) Funcs(funcs map[string]interface{}) {
    h.funcs = template.FuncMap{
        "include": h.Include,
        "defined": h.Defined,
    }

    for k, f := range funcs {
        h.funcs[k] = f
    }

    h.root.Funcs(h.funcs)
    h.loadTemplateFiles(h.dir)
}

/**
 * loadTemplateFiles loads all *.html files under the dir recursively
 */
func (h *Html) loadTemplateFiles(dir string) {
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
                txt := make([]byte, info.Size())
                if _,e := f.Read(txt); e == nil {

                    //init template
                    key := strings.TrimPrefix(
                            strings.TrimSuffix(uri, ".html"), h.dir)
                    template.Must(h.root.New(key).Parse(string(txt)))
                }

                f.Close()
            }
        }
    }

    d.Close()
}
