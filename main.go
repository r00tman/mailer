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

	prompt := CmdPrompt{}
	for {
		s.Clear()
		emitStr(s, 0, 0, tcell.StyleDefault, "hello world")

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if !prompt.Update(ev) {
				return
			}
		}
		prompt.Draw(s)
		s.Show()
	}
}
