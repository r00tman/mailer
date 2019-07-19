package main

import (
	"log"
	"strings"

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
			q <- TermEvent(ev)
		}
	}()

	go func() {
		c.Update(q)
	}()

	viewer.BackCallback = func() {
		go func() {
			q <- ViewMailboxEvent{}
		}()
	}
	list.ForwardCallback = func() {
		go func() {
			q <- ViewMessageEvent(imap.Message(list.List[list.ActiveIdx].(Message)))
		}()
	}
	filter := ""
	for {
		activeList := &viewer
		if isMailbox {
			activeList = &list
		}
		found := -1
		tryFind := func(start int, inc func(int) int) int {
			found := -1
			for i := start; i < len(activeList.List) && i >= 0; i = inc(i) {
				if strings.Contains(activeList.List[i].AsString(), filter) {
					activeList.ActiveIdx = i
					found = i
					break
				}
			}
			return found
		}
		inc := func(i int) int { return i + 1 }
		dec := func(i int) int { return i - 1 }

		rev := <-q
		switch rev := rev.(type) {
		case TermEvent:
			switch ev := rev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if !isPromptActive {
					switch ev.Rune() {
					case ':', '/', '?':
						prompt.str = ""
						isPromptActive = true
					case 'q':
						return
					case 'N':
						found := tryFind(dec(activeList.ActiveIdx), dec)

						if found < 0 {
							prompt.str = "Can't find '" + filter + "' starting from the end"
							found = tryFind(len(activeList.List)-1, dec)
							if found < 0 {
								prompt.str = "Can't find '" + filter + "'"
							}
						}
					case 'n':
						found := tryFind(inc(activeList.ActiveIdx), inc)

						if found < 0 {
							prompt.str = "Can't find '" + filter + "' starting from the beginning"
							found = tryFind(0, inc)
							if found < 0 {
								prompt.str = "Can't find '" + filter + "'"
							}
						}
					}
				}
				if isPromptActive {
					ipa, quit := prompt.Update(ev, q)
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
		case SetFilterEvent:
			filter = rev.F

			incFunc := dec
			if rev.Forward {
				incFunc = inc
			}
			found = tryFind(activeList.ActiveIdx, incFunc)
			if found < 0 {
				prompt.str = "Can't find '" + filter + "'"
			}
		case NewMessageEvent:
			list.List = append([]ListItem{Message(rev)}, list.List...)
		case RefreshEvent:
		case ViewMessageEvent:
			isMailbox = false
			viewer.List = []ListItem{}
			viewer.ActiveIdx = 0
			viewer.Offset = 0
			out := make(chan string, 0)
			go func(msg imap.Message) {
				c.ReadMail(msg, out)
				close(out)
			}(imap.Message(rev))
			go func() {
				l := []ListItem{}
				for m := range out {
					l = append(l, (Line)(m))
				}
				viewer.List = l
				q <- RefreshEvent{}
			}()
		case ViewMailboxEvent:
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
