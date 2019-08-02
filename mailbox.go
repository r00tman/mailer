package main

import (
	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
)

type Mailbox imap.MailboxInfo

func (m Mailbox) DrawMessage(s tcell.Screen, y int) {
	_ = EmitStrDef(s, 0, y, m.Name)
}

func (m Mailbox) String() string {
	return m.Name
}
