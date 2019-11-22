package gocui

import (
	"log"
	"os"
)

func l(v ...interface{}) {
	f, _ := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.SetOutput(f)
	log.Println(v...)
	f.Close()
}
