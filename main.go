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

	list := List{[]string{"1", "2", "3", "4"}, 0}
	prompt := CmdPrompt{}
	isPromptActive := false
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
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
		s.Clear()
		list.Draw(s, !isPromptActive)
		prompt.Draw(s, isPromptActive)
		s.Show()
	}
}
