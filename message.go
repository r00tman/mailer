package main

import (
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
)

type Message imap.Message

func isUnseen(flags []string) bool {
	for _, x := range flags {
		if x == imap.SeenFlag {
			return false
		}
	}
	return true
}

func formatDate(date time.Time) string {
	date = date.Local()
	now := time.Now().Truncate(time.Hour)
	// Mimic GMail behavior
	now = now.Add(time.Hour * (time.Duration)(now.Hour()/12*12-now.Hour()))

	template := "01/02/06"
	if date.After(now.Add(-12 * time.Hour)) {
		template = "15:04"
	} else if date.Year() >= now.Year() {
		template = "Jan 02"
	}
	return date.Format(template)
}

func (msg Message) AsString() string {
	sender := msg.Envelope.Sender
	sender_str := sender[0].PersonalName
	if len(sender_str) == 0 {
		sender_str = sender[0].MailboxName + "@" + sender[0].HostName
	}

	is_unseen := isUnseen(msg.Flags)
	unseen_str := " "
	if is_unseen {
		unseen_str = "*"
	}

	date_str := formatDate(msg.Envelope.Date)

	str := unseen_str + " " + sender_str + " " +
		msg.Envelope.Subject + " " + strings.Join(msg.Flags, " ") + " " +
		date_str

	return str
}

func (msg Message) DrawMessage(s tcell.Screen, y int) {
	w, _ := s.Size()

	sender := msg.Envelope.Sender
	sender_str := sender[0].PersonalName
	if len(sender_str) == 0 {
		sender_str = sender[0].MailboxName + "@" + sender[0].HostName
	}
	sender_str = TruncateFillRight(sender_str, 20)

	is_unseen := isUnseen(msg.Flags)
	unseen_str := " "
	if is_unseen {
		unseen_str = "*"
	}

	str := unseen_str + " " + sender_str + " " +
		msg.Envelope.Subject + " " + strings.Join(msg.Flags, " ")

	date_str := " " + formatDate(msg.Envelope.Date)

	str = TruncateFillRight(str, w-len(date_str)) + date_str

	_ = EmitStr(s, 0, y, tcell.StyleDefault.Bold(is_unseen), str)
}
