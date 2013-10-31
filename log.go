package potato


import (
    "os"
    "log"
)

type Logger struct {
    logger *log.Logger
}

func NewLogger(file *os.File) *Logger {
    return &Logger{
        log.New(file, "", log.LstdFlags),
    }
}

func (l *Logger) Println(v ...interface{}) {
    if Env == "dev" {
        log.Println(v...)
    }

    l.logger.Println(v...)
}

func (l *Logger) Fatal(v ...interface{}) {
    if Env == "dev" {
        log.Println(v...)
    }

    l.logger.Fatal(v...)
}
