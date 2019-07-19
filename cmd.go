package main

import (
	"strings"

	"github.com/gdamore/tcell"
)

type CmdPrompt struct {
	cursor uint
	str    string
}

func (self *CmdPrompt) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()
	l := emitStrDef(s, 0, h-1, self.str)
	if active {
		s.ShowCursor(l, h-1)
	}
}

func (self *CmdPrompt) Update(ev *tcell.EventKey, q chan Event) (bool, bool) {
	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if sz := len(self.str); sz > 0 {
			self.str = self.str[:sz-1]
		}
		if len(self.str) == 0 {
			return false, false
		}
	case tcell.KeyEnter:
		if self.str == ":q" {
			return false, true
		} else if strings.HasPrefix(self.str, "/") || strings.HasPrefix(self.str, "?") {
			go func(f string, forward bool) {
				q <- SetFilterEvent{f, forward}
			}(self.str[1:], self.str[0] == '/')
		}
		self.str = ""
		return false, false
	default:
		self.str += string(ev.Rune())
	}
	return true, false
}
