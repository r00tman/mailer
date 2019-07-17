package main

import "github.com/gdamore/tcell"

type List struct {
	list      []string
	activeIdx int
}

func (self *List) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()
	for i, l := range self.list {
		if i >= h-1 {
			break
		}
		_ = emitStr(s, 0, i, tcell.StyleDefault, l)
		if active && i == self.activeIdx {
			s.ShowCursor(0, i)
		}
	}
}

func (self *List) Update(ev *tcell.EventKey) {
	switch ev.Rune() {
	case 'j':
		self.activeIdx += 1
	case 'k':
		self.activeIdx -= 1
	}
	if self.activeIdx >= len(self.list) {
		self.activeIdx = len(self.list) - 1
	}
	if self.activeIdx < 0 {
		self.activeIdx = 0
	}
}
