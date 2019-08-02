package main

import (
	"strings"

	"github.com/gdamore/tcell"
)

type Line string

func (msg Line) DrawMessage(s tcell.Screen, y int) {
	_ = EmitStrDef(s, 0, y, strings.TrimRight(string(msg), "\r\n"))
}

func (msg Line) String() string {
	return string(msg)
}
