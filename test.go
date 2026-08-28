package main

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
)

func main() {
	f := tele.File{FileID: "123"}
	p := &tele.Photo{File: f}
	fmt.Printf("%T\n", p)
}
