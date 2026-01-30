package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}


func New() *Logger{
	return &Logger{
		Logger: log.New(
			os.Stdout,
			"",
			log.LstdFlags|log.Lshortfile,
		),
	}
}


func (l *Logger) Info(msg string, fields ...any){
	l.Println(append([]any{"[INFO]", msg}, fields...)...)
}

func (l *Logger) Error(msg string, fields ...any){
	l.Println(append([]any{"[ERROR]", msg}, fields...)...)
}