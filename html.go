package potato

import (
    "os"
    "strings"
    _"html/template"
)


type Html struct {
    templates map[string][]byte
    files map[string]os.FileInfo
}

/**
 * LoadTemplates loads all html files under dir
 */
func (h *Html) LoadTemplates(dir string) {
    h.files = make(map[string]os.FileInfo)
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
        if info.IsDir() {
            h.LoadHtmlFiles(dir + info.Name() + "/")
        } else if strings.HasSuffix(info.Name(), ".html") {
            key := dir + strings.TrimRight(info.Name(), ".html")
            h.files[key] = info
        }
    }
}
