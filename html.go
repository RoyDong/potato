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
    d, e := os.Open(dir)
    if e != nil {
        log.Fatal("templates not found:", e)
    }

    infos, e := d.Readdir(-1)
    if e != nil {
        log.Fatal("reading templates error:", e)
    }

    for _,info := range infos {
        log.Println(info.Name(), info.IsDir())
    }
}

func (h *Html) LoadAllFiles(d os.File) {
    infos, e := d.Readdir(-1)
    if e != nil {
        log.Fatal("reading templates error:", e)
    }

    for _,info := range infos {
        if strings.HasSuffix(info.Name(), ".html") {

        }
    }
}
