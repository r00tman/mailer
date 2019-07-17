package main

import "github.com/gdamore/tcell"

type Event interface {
}

type TEvent struct {
	t tcell.Event
}

type MEvent struct {
	m string
}
