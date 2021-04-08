package main

import (
	"regexp"
	"strconv"

	"github.com/gdamore/tcell"
)

type ListItem interface {
	DrawMessage(s tcell.Screen, y int)
	String() string
}

type List struct {
	List      []ListItem
	ActiveIdx int
	Offset    int
    Updating  bool
	chord     string

	BackCallback    func()
	ForwardCallback func()
	ToggleReadCallback func()
}

func NewList() List {
	return List{[]ListItem{}, 0, 0, false, "", func() {}, func() {}, func() {}}
}

func (self *List) Clear() {
	self.List = []ListItem{}
	self.ActiveIdx = 0
	self.Offset = 0
}

func (self *List) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()

	if len(self.List) == 0 {
        if self.Updating {
            EmitStrDef(s, 0, 0, "Updating\u2026")
        } else {
            EmitStrDef(s, 0, 0, "Mailbox is empty")
        }
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

	chordHasInt := false
	intChordRe := regexp.MustCompile("^[0-9]+")
	intMatch, nonIntMatch := "", self.chord
	if loc := intChordRe.FindStringIndex(self.chord); loc != nil {
		intMatch = self.chord[loc[0]:loc[1]]
		nonIntMatch = self.chord[loc[1]:]
		ninc, err := strconv.Atoi(intMatch)
		if err == nil {
			inc = ninc
			chordHasInt = true
		} else {
			chordHasInt = false
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
	case 'r':
		self.ToggleReadCallback()
	case 'j':
		self.ActiveIdx += inc
	case 'k':
		self.ActiveIdx -= inc
	case 'g':
		if nonIntMatch == "g" {
			if chordHasInt {
				self.ActiveIdx = inc
			} else {
				self.ActiveIdx = 0
			}
		} else {
			self.chord = intMatch + "g"
			newChord = true
		}
	case 'G':
		if chordHasInt {
			self.ActiveIdx = inc
		} else {
			self.ActiveIdx = len(self.List) - 1
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if !chordHasInt {
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
	case tcell.KeyHome:
		self.ActiveIdx = 0
	case tcell.KeyEnd:
		self.ActiveIdx = len(self.List) - 1
	case tcell.KeyCtrlL:
		adjustView()
	}
	if !newChord {
		self.chord = ""
	}

	self.InvalidateRange(s)
}
