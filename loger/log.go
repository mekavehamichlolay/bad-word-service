package loger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type loger struct {
	wg    *sync.WaitGroup
	log   *log.Logger
	file  *os.File
	mutex *sync.Mutex
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
		fmt.Println("Failed to create log file")
		return nil
	}
	loger := loger{
		log:   &log.Logger{},
		mutex: &sync.Mutex{},
		wg:    wg,
		file:  file,
	}
	loger.log.SetOutput(file)
	loger.log.SetFlags(log.LstdFlags)
	return &loger
}

func (l *loger) Info(message string) {
	l.wg.Add(1)
	go func() {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.log.SetPrefix("INFO: ")
		_, file, line, _ := runtime.Caller(0)
		file = fileFormatter(file)
		l.log.Printf("%s:%d: %s\n", file, line, message)
		l.wg.Done()
	}()
}
func (l *loger) Warn(message string) {
	l.wg.Add(1)
	go func() {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.log.SetPrefix("WARN: ")
		_, file, line, _ := runtime.Caller(0)
		file = fileFormatter(file)
		l.log.Printf("%s:%d: %s\n", file, line, message)
		l.wg.Done()
	}()
}
func (l *loger) Err(message string) {
	l.wg.Add(1)
	go func() {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.log.SetPrefix("ERR: ")
		_, file, line, _ := runtime.Caller(0)
		file = fileFormatter(file)
		l.log.Printf("%s:%d: %s\n", file, line, message)
		l.wg.Done()
	}()
}
func (l *loger) Close() {
	l.file.Close()
}
func fileFormatter(file string) string {
	parts := strings.Split(file, "/")
	return parts[len(parts)-1]
}
