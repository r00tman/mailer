package main

import (
	"strconv"

	"github.com/gdamore/tcell"
)

type ListItem interface {
	DrawMessage(s tcell.Screen, y int)
	AsString() string
}

type List struct {
	List      []ListItem
	ActiveIdx int
	Offset    int
	chord     string

	BackCallback    func()
	ForwardCallback func()
}

func (self *List) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()

	if len(self.List) == 0 {
		EmitStrDef(s, 0, 0, "Updating\u2026")
		s.ShowCursor(0, 0)
	} else {
		for i, msg := range self.List[self.Offset:] {
			if i >= h-1 {
				break
			}
			msg.DrawMessage(s, i)
			if active && i+self.Offset == self.ActiveIdx {
				s.ShowCursor(0, i)
			}
		}
	}
}

func (self *List) InvalidateRange(s tcell.Screen) {
	_, h := s.Size()
	if self.ActiveIdx >= len(self.List) {
		self.ActiveIdx = len(self.List) - 1
	}
	if self.ActiveIdx < 0 {
		self.ActiveIdx = 0
	}

	if self.ActiveIdx >= self.Offset+h-1 {
		self.Offset = self.ActiveIdx + 2 - h
	}
	if self.Offset >= len(self.List) {
		self.Offset = len(self.List) - 1
	}
	if self.ActiveIdx < self.Offset {
		self.Offset = self.ActiveIdx
	}
	if self.Offset < 0 {
		self.Offset = 0
	}
}

func (self *List) Update(s tcell.Screen, ev *tcell.EventKey) {
	_, h := s.Size()
	newChord := false
	inc := 1
	chordIsInt := true
	if len(self.chord) > 0 {
		ninc, err := strconv.Atoi(self.chord)
		if err == nil {
			inc = ninc
		} else {
			chordIsInt = false
		}
	}
	adjustView := func() {
		if self.chord == "^" {
			self.Offset = self.ActiveIdx + 1 - h/2
			self.chord = "^^"
			newChord = true
		} else if self.chord == "^^" {
			self.Offset = self.ActiveIdx - h + 1
			self.chord = ""
			newChord = false
		} else {
			self.Offset = self.ActiveIdx
			self.chord = "^"
			newChord = true
		}
	}
	switch ev.Rune() {
	case 'h':
		self.BackCallback()
	case 'l':
		self.ForwardCallback()
	case 'j':
		self.ActiveIdx += inc
	case 'k':
		self.ActiveIdx -= inc
	case 'g':
		if self.chord == "g" {
			self.ActiveIdx = 0
		} else {
			self.chord = "g"
			newChord = true
		}
	case 'G':
		self.ActiveIdx = len(self.List) - 1
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if !chordIsInt {
			self.chord = ""
		}
		self.chord = self.chord + string(ev.Rune())
		newChord = true
	case '}', ' ':
		self.ActiveIdx += inc * h / 2
	case '{', 'b':
		self.ActiveIdx -= inc * h / 2
	case ';':
		adjustView()
	}
	switch ev.Key() {
	case tcell.KeyLeft, tcell.KeyEsc:
		self.BackCallback()
	case tcell.KeyRight, tcell.KeyEnter:
		self.ForwardCallback()
	case tcell.KeyPgDn, tcell.KeyCtrlD:
		self.ActiveIdx += inc * h / 2
	case tcell.KeyPgUp, tcell.KeyCtrlU:
		self.ActiveIdx -= inc * h / 2
	case tcell.KeyUp:
		self.ActiveIdx -= inc
	case tcell.KeyDown:
		self.ActiveIdx += inc
	case tcell.KeyCtrlL:
		adjustView()
	}
	if !newChord {
		self.chord = ""
	}

	self.InvalidateRange(s)
}
