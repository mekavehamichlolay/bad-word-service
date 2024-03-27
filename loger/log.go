package loger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type loger struct {
	wg   *sync.WaitGroup
	log  *log.Logger
	file *os.File
}
type Loger interface {
	Info(message string)
	Warn(message string)
	Err(message string)
	Close()
}

func NewLoger(path string, wg *sync.WaitGroup) Loger {
	file, err := os.Create(fmt.Sprintf("%d.%s", time.Now().Unix(), path))
	if err != nil {
		log.Fatal(err)
	}
	loger := loger{
		log:  &log.Logger{},
		wg:   wg,
		file: file,
	}
	loger.log.SetOutput(file)
	loger.log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return &loger
}

func (l *loger) Info(message string) {
	l.wg.Add(1)
	go func() {
		l.log.SetPrefix("INFO: ")
		l.log.Println(message)
		l.wg.Done()
	}()
}
func (l *loger) Warn(message string) {
	l.wg.Add(1)
	go func() {
		l.log.SetPrefix("WARN: ")
		l.log.Println(message)
		l.wg.Done()
	}()
}
func (l *loger) Err(message string) {
	l.wg.Add(1)
	go func() {
		l.log.SetPrefix("ERR: ")
		l.log.Println(message)
		l.wg.Done()
	}()
}
func (l *loger) Close() {
	l.file.Close()
}
