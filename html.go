package potato

import (
    "os"
    "log"
    "strings"
    "html/template"
)


var MaxFileSize = int64(2 * 1024 * 1024)

type Template struct {
    *template.Template
}

type Html struct {
    templates map[string]*Template
}

/**
 * LoadTemplates loads all html files under dir
 */
func (h *Html) LoadTemplates(dir string) {
    h.LoadHtmlFiles(dir)
}

/**
 * LoadHtmlFiles loads all *.html files under the dir recursively
 */
func (h *Html) LoadHtmlFiles(dir string) {
    d, e := os.Open(dir)
    if e != nil {
        return
    }

    dinfo, e := d.Readdir(-1)
    if e != nil {
        return
    }

    for _,info := range dinfo {
        uri := dir + info.Name()
        if info.IsDir() {
            h.LoadHtmlFiles(uri + "/")
        } else if info.Size() <= MaxFileSize &&
                strings.HasSuffix(info.Name(), ".html") {
            if file, e := os.Open(uri); e == nil {
                str := make([]byte, info.Size())
                if _,e := file.Read(str); e == nil {
                    key := dir + strings.TrimRight(info.Name(), ".html")

                    log.Println(key, string(str))
                }
            }
        }
    }
}
