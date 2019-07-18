package main

import (
	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
)

type Event interface {
}

type TermEvent struct {
	t tcell.Event
}

type NewMessageEvent struct {
	m *imap.Message
}

type ViewMessageEvent struct {
	m *imap.Message
}
