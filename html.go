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
    template *template.Template
    root string
    funcs template.FuncMap
}

/**
 * LoadTemplates loads all html files under dir
 */
func (h *Html) LoadTemplates(dir string) {
    h.template = template.New("/")
    h.root = dir
    h.funcs = template.FuncMap{
        "include": h.Include,
    }
    h.LoadHtmlFiles(h.root)
}

func (h *Html) Template(name string) *template.Template {
    return h.template.Lookup(name)
}

func (h *Html) Include(name string, data interface{}) string {
    if t := h.template.Lookup(name); t != nil {
        buffer := new(bytes.Buffer)
        t.Execute(buffer, data)
        return string(buffer.Bytes())
    }

    panic(name + " template not found")
}

/**
 * LoadHtmlFiles loads all *.html files under the dir recursively
 */
func (h *Html) LoadHtmlFiles(dir string) {
    d, e := os.Open(dir)
    if e != nil { return }

    dinfo, e := d.Readdir(-1)
    if e != nil { return }

    for _,info := range dinfo {
        uri := dir + info.Name()
        if info.IsDir() {
            h.LoadHtmlFiles(uri + "/")

        //check file
        } else if info.Size() <= MaxFileSize &&
                strings.HasSuffix(info.Name(), ".html") {

            //load file
            if f, e := os.Open(uri); e == nil {
                str := make([]byte, info.Size())
                if _,e := f.Read(str); e == nil {

                    //init template
                    key := strings.TrimPrefix(strings.TrimSuffix(uri, ".html"), h.root)
                    h.template.New(key).Funcs(h.funcs).Parse(string(str))
                }
            }
        }
    }
}
