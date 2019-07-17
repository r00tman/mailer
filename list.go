package main

import "github.com/gdamore/tcell"

type List struct {
	list      []string
	activeIdx int
	offset    int
}

func (self *List) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()

	list := []string{}
	if len(self.list) == 0 {
		list = append(list, "Updating...")
	} else {
		for i := range self.list {
			list = append(list, self.list[len(self.list)-1-i])
		}
	}
	for i, l := range list[self.offset:] {
		if i >= h-1 {
			break
		}
		_ = emitStr(s, 0, i, tcell.StyleDefault, l)
		if active && i+self.offset == self.activeIdx {
			s.ShowCursor(0, i)
		}
	}
}

func (self *List) Update(s tcell.Screen, ev *tcell.EventKey) {
	_, h := s.Size()
	switch ev.Rune() {
	case 'j':
		self.activeIdx += 1
	case 'k':
		self.activeIdx -= 1
	case 'G':
		self.activeIdx = len(self.list) - 1
	}
	switch ev.Key() {
	case tcell.KeyPgDn, tcell.KeyCtrlD:
		self.activeIdx += h / 2
	case tcell.KeyPgUp, tcell.KeyCtrlU:
		self.activeIdx -= h / 2
	case tcell.KeyUp:
		self.activeIdx -= 1
	case tcell.KeyDown:
		self.activeIdx += 1
	}

	if self.activeIdx >= len(self.list) {
		self.activeIdx = len(self.list) - 1
	}
	if self.activeIdx < 0 {
		self.activeIdx = 0
	}

	if self.activeIdx >= self.offset+h-1 {
		self.offset = self.activeIdx + 2 - h
	}
	if self.offset >= len(self.list) {
		self.offset = len(self.list) - 1
	}
	if self.activeIdx < self.offset {
		self.offset = self.activeIdx
	}
	if self.offset < 0 {
		self.offset = 0
	}
}
