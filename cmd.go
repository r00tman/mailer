package main

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) int {
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

type CmdPrompt struct {
	cursor uint
	str    string
}

func (self *CmdPrompt) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()
	l := emitStr(s, 0, h-1, tcell.StyleDefault, self.str)
	if active {
		s.ShowCursor(l, h-1)
	}
}

func (self *CmdPrompt) Update(ev *tcell.EventKey) (bool, bool) {
	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if sz := len(self.str); sz > 0 {
			self.str = self.str[:sz-1]
		}
	case tcell.KeyEnter:
		if self.str == ":q" {
			return false, true
		}
		self.str = ""
		return false, false
	default:
		self.str += string(ev.Rune())
	}
	return true, false
}
