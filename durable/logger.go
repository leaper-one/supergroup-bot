package durable

import (
	"log"
	"runtime"
)

type Logger struct {
}

func (c *Logger) Println(v ...interface{}) {
	caller := 1
	for {
		if _, file, line, ok := runtime.Caller(caller); ok {
			log.Println(v, file, line)
			caller++
		} else {
			return
		}
	}
}
