package main

import (
	"log"

	"github.com/emersion/go-imap"
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

	list := List{[]ListItem{}, 0, 0, ""}
	prompt := CmdPrompt{}
	isPromptActive := false
	q := make(chan Event, 0)
	go func() {
		for {
			ev := s.PollEvent()
			q <- &TermEvent{t: ev}
		}
	}()
	c := Email{}
	c.Connect()
	defer c.Logout()

	go func() {
		c.Update(q)
	}()
	for {
		rev := <-q
		switch rev := rev.(type) {
		case *TermEvent:
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
					list.Update(s, ev, q)
				}
			}
		case *NewMessageEvent:
			list.list = append([]ListItem{(*Message)(rev.m)}, list.list...)
		case *ViewMessageEvent:
			out := make(chan string, 0)
			go func(msg *imap.Message) {
				c.ReadMail(msg, out)
				close(out)
			}(rev.m)
			go func() {
				list.list = []ListItem{}
				for m := range out {
					list.list = append(list.list, (Line)(m))
				}
			}()
		default:
			return
		}
		s.Clear()
		list.Draw(s, !isPromptActive)
		prompt.Draw(s, isPromptActive)
		s.Show()
	}
}
