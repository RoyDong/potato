package potato

import (
    "os"
    "strings"
    "html/template"
)


var (
    //2m
    MaxFileSize = int64(2 * 1024 * 1024)
)

type Html struct {
    template *template.Template
    root string
}

/**
 * LoadTemplates loads all html files under dir
 */
func (h *Html) LoadTemplates(dir string) {
    h.template = template.New("/")
    h.root = dir
    h.LoadHtmlFiles(h.root)
}

func (h *Html) Template(name string) *template.Template {
    return h.template.Lookup(name)
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

        //filt html files
        } else if info.Size() <= MaxFileSize &&
                strings.HasSuffix(info.Name(), ".html") {

            //load file
            if f, e := os.Open(uri); e == nil {
                str := make([]byte, info.Size())
                if _,e := f.Read(str); e == nil {

                    //init template
                    key := strings.TrimPrefix(strings.TrimSuffix(uri, ".html"), h.root)
                    t := h.template.New(key)
                    t.Parse(string(str))
                }
            }
        }
    }
}

