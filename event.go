package main

import (
	"github.com/gdamore/tcell"
)

type Event interface{}

type TermEvent tcell.Event

type QuitEvent struct{}

type NewMessageEvent Message

type NewMailboxEvent Mailbox

type SetFilterEvent struct {
	F       string
	Forward bool
}

type RefreshEvent struct{}

type ViewAccountEvent struct{}

type ViewMailboxEvent Mailbox

type ToggleReadEvent Message

type ViewMessageEvent Message
