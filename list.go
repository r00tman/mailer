package main

import (
	"strconv"

	"github.com/gdamore/tcell"
)

type ListItem interface {
	drawMessage(s tcell.Screen, y int)
	AsString() string
}

type List struct {
	list      []ListItem
	activeIdx int
	offset    int
	chord     string

	backCallback    func()
	forwardCallback func()
}

func (self *List) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()

	if len(self.list) == 0 {
		emitStrDef(s, 0, 0, "Updating\u2026")
		s.ShowCursor(0, 0)
	} else {
		for i, msg := range self.list[self.offset:] {
			if i >= h-1 {
				break
			}
			msg.drawMessage(s, i)
			if active && i+self.offset == self.activeIdx {
				s.ShowCursor(0, i)
			}
		}
	}
}

func (self *List) Update(s tcell.Screen, ev *tcell.EventKey) {
	_, h := s.Size()
	newChord := false
	inc := 1
	if len(self.chord) > 0 {
		ninc, err := strconv.Atoi(self.chord)
		if err == nil {
			inc = ninc
		}
	}
	switch ev.Rune() {
	case 'h':
		self.backCallback()
	case 'l':
		self.forwardCallback()
	case 'j':
		self.activeIdx += inc
	case 'k':
		self.activeIdx -= inc
	case 'g':
		if self.chord == "g" {
			self.activeIdx = 0
		} else {
			self.chord = "g"
			newChord = true
		}
	case 'G':
		self.activeIdx = len(self.list) - 1
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		self.chord = self.chord + string(ev.Rune())
		newChord = true
	case '}':
		self.activeIdx += inc * h / 2
	case '{':
		self.activeIdx -= inc * h / 2
	}
	switch ev.Key() {
	case tcell.KeyPgDn, tcell.KeyCtrlD:
		self.activeIdx += inc * h / 2
	case tcell.KeyPgUp, tcell.KeyCtrlU:
		self.activeIdx -= inc * h / 2
	case tcell.KeyUp:
		self.activeIdx -= inc
	case tcell.KeyDown:
		self.activeIdx += inc
	}
	if !newChord {
		self.chord = ""
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
