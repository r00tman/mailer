package main

import (
	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
)

type Event interface{}

type TermEvent tcell.Event

type NewMessageEvent imap.Message

type SetFilterEvent struct {
	F       string
	Forward bool
}

type RefreshEvent struct{}

type ViewMailboxEvent struct{}

type ViewMessageEvent imap.Message
