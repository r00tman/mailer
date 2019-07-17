package main

import (
	"log"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mattn/go-runewidth"
)

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) int {
	x_or := x
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
	return x - x_or
}

type CmdPrompt struct {
	cursor uint
	str    string
	active bool
}

func (c *CmdPrompt) Draw(s tcell.Screen) {
	_, h := s.Size()
	l := emitStr(s, 0, h-1, tcell.StyleDefault, c.str)
	s.ShowCursor(l, h-1)
}

func (c *CmdPrompt) Update(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if !c.active {
		} else if sz := len(c.str); sz > 0 {
			c.str = c.str[:sz-1]
		}
	case tcell.KeyEnter:
		if c.str == ":q" {
			return false
		}
		c.str = ""
		c.active = false
	default:
		if c.active || ev.Rune() == ':' {
			c.str += string(ev.Rune())
			c.active = true
		}
	}
	return true
}

func main() {
	encoding.Register()

	s, e := tcell.NewScreen()
	prompt := CmdPrompt{}

	if e != nil {
		log.Fatal(e)
	}
	if e := s.Init(); e != nil {
		log.Fatal(e)
	}
	defer s.Fini()
	s.EnableMouse()
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
