package main

import (
	"github.com/lonelycode/roomystan/scanner"
	"github.com/lonelycode/roomystan/tracker"
)

func main() {
	b := scanner.New()
	t := tracker.New([]string{"Pam", "room-assistant companion"})
	b.Scan(t.Update)
}