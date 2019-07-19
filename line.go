package main

import (
	"strings"

	"github.com/gdamore/tcell"
)

type Line string

func (msg Line) drawMessage(s tcell.Screen, y int) {
	_ = emitStrDef(s, 0, y, strings.TrimRight(string(msg), "\r\n"))
}

func (msg Line) AsString() string {
	return string(msg)
}
