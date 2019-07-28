package main

import (
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

func main() {
	login, password, host := getOrCreateAccount()
	c := Email{}
	println("Connecting\u2026")
	c.Connect(login, password, host)
	defer func() {
		println("Logging out\u2026")
		c.Logout()
	}()

	encoding.Register()

	s, e := tcell.NewScreen()

	if e != nil {
		log.Fatal(e)
	}
	if e := s.Init(); e != nil {
		log.Fatal(e)
	}
	defer s.Fini()

	mailboxes := NewList()
	messages := NewList()
	viewer := NewList()

	prompt := CmdPrompt{}
	isPromptActive := false
	activeList := &messages
	activeMbox := "INBOX"
	q := make(chan Event, 0)
	go func() {
		for {
			ev := s.PollEvent()
			q <- TermEvent(ev)
		}
	}()

	go func() {
		c.Update(q, activeMbox)
	}()

	mailboxes.ForwardCallback = func() {
		go func() {
			mbox := Mailbox{Name: ""}
			if len(mailboxes.List) > 0 {
				mbox = mailboxes.List[mailboxes.ActiveIdx].(Mailbox)
			}
			q <- ViewMailboxEvent(mbox)
		}()
	}
	messages.ForwardCallback = func() {
		go func() {
			if len(messages.List) > 0 {
				message := messages.List[messages.ActiveIdx].(Message)
				q <- ViewMessageEvent(message)
			}
		}()
	}
	messages.BackCallback = func() {
		if len(mailboxes.List) == 0 {
			go func() {
				c.Mailboxes(q)
			}()
		}
		go func() {
			q <- ViewAccountEvent{}
		}()
	}
	viewer.BackCallback = func() {
		go func() {
			q <- ViewMailboxEvent(Mailbox{Name: ""})
		}()
	}
	filter := ""
	for {
		tryFind := func(start int, inc func(int) int) bool {
			f := strings.ToLower(filter)
			for i := start; i < len(activeList.List) && i >= 0; i = inc(i) {
				message := activeList.List[i].AsString()
				message = strings.ToLower(message)
				if strings.Contains(message, f) {
					activeList.ActiveIdx = i
					activeList.InvalidateRange(s)
					return true
				}
			}
			return false
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

						if !found {
							prompt.str = "Can't find '" + filter + "' starting from the end"
							found = tryFind(len(activeList.List)-1, dec)
							if !found {
								prompt.str = "Can't find '" + filter + "'"
							}
						}
					case 'n':
						found := tryFind(inc(activeList.ActiveIdx), inc)

						if !found {
							prompt.str = "Can't find '" + filter + "' starting from the beginning"
							found = tryFind(0, inc)
							if !found {
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
				} else {
					activeList.Update(s, ev)
				}
			}
		case SetFilterEvent:
			filter = rev.F

			incFunc := dec
			if rev.Forward {
				incFunc = inc
			}
			found := tryFind(activeList.ActiveIdx, incFunc)
			if !found {
				prompt.str = "Can't find '" + filter + "'"
			}
		case NewMessageEvent:
			messages.List = append([]ListItem{Message(rev)}, messages.List...)
		case NewMailboxEvent:
			mailboxes.List = append(mailboxes.List, Mailbox(rev))
		case RefreshEvent:
		case ViewMessageEvent:
			activeList = &viewer
			viewer.Clear()
			out := make(chan string, 0)
			go func(msg imap.Message) {
				c.ReadMail(msg, activeMbox, out)
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
			activeList = &messages
			if rev.Name != "" {
				activeMbox = rev.Name
				messages.Clear()
				go func() {
					c.Update(q, activeMbox)
				}()
			}
		case ViewAccountEvent:
			activeList = &mailboxes
		default:
			return
		}
		s.Clear()
		activeList.Draw(s, !isPromptActive)
		prompt.Draw(s, isPromptActive)
		s.Show()
	}
}
