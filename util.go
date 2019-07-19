package main

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func EmitStr(s tcell.Screen, x, y int, style tcell.Style, str string) int {
	x_or := x
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
	return x - x_or
}

func EmitStrDef(s tcell.Screen, x, y int, str string) int {
	return EmitStr(s, x, y, tcell.StyleDefault, str)
}

func TruncateFillRight(x string, w int) string {
	x = runewidth.Truncate(x, w, "\u2026")
	x = runewidth.FillRight(x, w)
	return x
}
