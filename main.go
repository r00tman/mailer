package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

type Event interface {
}

type TEvent struct {
	t tcell.Event
}

type MEvent struct {
	m string
}

func main() {
	encoding.Register()

	s, e := tcell.NewScreen()

	if e != nil {
		log.Fatal(e)
	}
	if e := s.Init(); e != nil {
		log.Fatal(e)
	}
	defer s.Fini()
	s.EnableMouse()

	list := List{[]string{"1", "2", "3", "4"}, 0}
	prompt := CmdPrompt{}
	isPromptActive := false
	q := make(chan Event, 0)
	go func() {
		for {
			ev := s.PollEvent()
			q <- &TEvent{t: ev}
		}
	}()
	go func() {
		for {
			q <- &MEvent{m: "hi"}
			time.Sleep(1 * time.Second)
		}
	}()
	for {
		rev := <-q
		switch rev := rev.(type) {
		case *TEvent:
			switch ev := rev.t.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if ev.Rune() == ':' && !isPromptActive {
					isPromptActive = true
				}
				if isPromptActive {
					ipa, quit := prompt.Update(ev)
					isPromptActive = ipa
					if quit {
						return
					}
				} else {
					list.Update(ev)
				}
			}
		case *MEvent:
			list.list = append(list.list, rev.m)
		default:
			return
		}
		s.Clear()
		list.Draw(s, !isPromptActive)
		prompt.Draw(s, isPromptActive)
		s.Show()
	}
}
