package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
)

type List struct {
	list      []*imap.Message
	activeIdx int
	offset    int
	chord     string
}

func isUnseen(flags []string) bool {
	for _, x := range flags {
		if x == imap.SeenFlag {
			return false
		}
	}
	return true
}

func formatDate(date time.Time) string {
	now := time.Now().Truncate(time.Hour)
	// Mimic GMail behavior
	now = now.Add(time.Hour * (time.Duration)(now.Hour()/6*6-now.Hour()))

	template := "01/02/06"
	if date.After(now.Add(-12 * time.Hour)) {
		template = "15:04"
	} else if date.Year() >= now.Year() {
		template = "Jan 02"
	}
	return date.Format(template)
}

func drawMessage(s tcell.Screen, msg *imap.Message, y int) {
	w, _ := s.Size()

	sender := msg.Envelope.Sender
	sender_str := sender[0].PersonalName
	if len(sender_str) == 0 {
		sender_str = sender[0].MailboxName + "@" + sender[0].HostName
	}
	sender_str = truncateFillRight(sender_str, 20)

	is_unseen := isUnseen(msg.Flags)
	unseen_str := " "
	if is_unseen {
		unseen_str = "*"
	}

	str := unseen_str + " " + sender_str + " " +
		msg.Envelope.Subject + " " + strings.Join(msg.Flags, " ")

	date_str := " " + formatDate(msg.Envelope.Date)

	str = truncateFillRight(str, w-len(date_str)) + date_str

	_ = emitStr(s, 0, y, tcell.StyleDefault.Bold(is_unseen), str)
}

func (self *List) Draw(s tcell.Screen, active bool) {
	_, h := s.Size()

	if len(self.list) == 0 {
		emitStrDef(s, 0, 0, "Updating\u2026")
	} else {
		list := []*imap.Message{}
		for i := range self.list {
			list = append(list, self.list[len(self.list)-1-i])
		}
		for i, msg := range list[self.offset:] {
			if i >= h-1 {
				break
			}
			drawMessage(s, msg, i)
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
	case 'l':
		go func(msg *imap.Message) {
			c := Email{}
			c.Connect()
			c.ReadMail(msg)
			c.Logout()
		}(self.list[len(self.list)-1-self.activeIdx])
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
