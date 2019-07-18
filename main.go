package main

import (
	"log"

	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

func main() {
	c := Email{}
	c.Connect()
	defer c.Logout()

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

	list := List{[]ListItem{}, 0, 0, "", func() {}, func() {}}
	viewer := List{[]ListItem{}, 0, 0, "", func() {}, func() {}}

	prompt := CmdPrompt{}
	isPromptActive := false
	isMailbox := true
	q := make(chan Event, 0)
	go func() {
		for {
			ev := s.PollEvent()
			q <- &TermEvent{t: ev}
		}
	}()

	go func() {
		c.Update(q)
	}()

	viewer.backCallback = func() {
		go func() {
			q <- &ViewMailboxEvent{}
		}()
	}
	list.forwardCallback = func() {
		go func() {
			q <- &ViewMessageEvent{(*imap.Message)(list.list[list.activeIdx].(*Message))}
		}()
	}
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
				} else if isMailbox {
					list.Update(s, ev)
				} else {
					viewer.Update(s, ev)
				}
			}
		case *NewMessageEvent:
			list.list = append([]ListItem{(*Message)(rev.m)}, list.list...)
		case *RefreshEvent:
		case *ViewMessageEvent:
			isMailbox = false
			viewer.list = []ListItem{}
			viewer.activeIdx = 0
			viewer.offset = 0
			out := make(chan string, 0)
			go func(msg *imap.Message) {
				c.ReadMail(msg, out)
				close(out)
			}(rev.m)
			go func() {
				l := []ListItem{}
				for m := range out {
					l = append(l, (Line)(m))
				}
				viewer.list = l
				q <- &RefreshEvent{}
			}()
		case *ViewMailboxEvent:
			isMailbox = true
		default:
			return
		}
		s.Clear()
		if isMailbox {
			list.Draw(s, !isPromptActive)
		} else {
			viewer.Draw(s, !isPromptActive)
		}
		prompt.Draw(s, isPromptActive)
		s.Show()
	}
}
