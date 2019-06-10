package main

import (
	"github.com/lpicanco/i3-autodisplay/display"
)

func main() {
	display.Refresh()
	display.Refresh()
	display.ListenEvents()
}
