package main

import (
	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
)

type Event interface {
}

type TEvent struct {
	t tcell.Event
}

type MEvent struct {
	m *imap.Message
}
