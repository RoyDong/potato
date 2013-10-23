package potato

import (
    "os"
    "log"
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
    h.LoadDirFiles("template")
    log.Println(h.files)
}

func (h *Html) LoadDirFiles(name string) {
    dir, e := os.Open("./" + name)
    if e != nil {
        return
    }

    dinfo, e := dir.Readdir(-1)
    if e != nil {
        return
    }

    for _,info := range dinfo {
        if info.IsDir() {
            h.LoadDirFiles(name + "/" + info.Name())
        } else if strings.HasSuffix(info.Name(), ".html") {
            key := name + "/" + strings.TrimRight(info.Name(), ".html")
            h.files[key] = info
        }
    }
}
