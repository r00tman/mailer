package main

import (
	"log"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

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

	list := List{[]string{}, 0, 0, ""}
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
		c := Email{}
		c.Update(q)
	}()
	for {
		rev := <-q
		switch rev := rev.(type) {
		case *TEvent:
			switch ev := rev.t.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if !isPromptActive {
					if ev.Rune() == ':' {
						isPromptActive = true
					}
					if ev.Rune() == 'q' {
						return
					}
				}
				if isPromptActive {
					ipa, quit := prompt.Update(ev)
					isPromptActive = ipa
					if quit {
						return
					}
				} else {
					list.Update(s, ev)
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
