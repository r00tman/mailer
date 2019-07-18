package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os/exec"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
)

type Email struct {
	c *client.Client
}

func (self *Email) Connect() {
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	self.c = c
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Login(USER, PASSWD); err != nil {
		log.Fatal(err)
	}
}

func (self *Email) Logout() {
	self.c.Logout()
}

func dfs(m *message.Entity, out chan string) {
	if mr := m.MultipartReader(); mr != nil {
		// This is a multipart message
		out <- fmt.Sprintln("This is a multipart message containing:")
		for {
			p, err := mr.NextPart()

			if err == io.EOF {
				break
			} else if err != nil {
				out <- fmt.Sprintln(err)
				continue
			}

			t, _, _ := p.Header.ContentType()
			out <- fmt.Sprintln("A part with type", t)

			dfs(p, out)
			// b, _ := ioutil.ReadAll(p.Body)
			// log.Println("A part with type", string(b))
		}
	} else {
		t, _, _ := m.Header.ContentType()
		b := []byte{}
		out <- fmt.Sprintln("This is a non-multipart message with type", t)
		out <- fmt.Sprintln("------------------------------------------" + strings.Repeat("-", len(t)))
		if t == "text/html" {
			c := exec.Command("w3m", "-dump", "-T", "text/html")
			c.Stdin = m.Body
			newb, err := c.Output()
			if err == nil {
				b = newb
			}

		} else if t == "text/plain" {
			newb, err := ioutil.ReadAll(m.Body)
			if err == nil {
				b = newb
			}
		}
		for _, x := range strings.Split(string(b), "\n") {
			out <- x
		}
	}
}

func (self *Email) ReadMail(msg *imap.Message, out chan string) {
	c := self.c
	_, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	from := msg.SeqNum
	to := msg.SeqNum
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	section := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	out <- fmt.Sprintln("Last message:")
	amsg := <-messages

	// log.Print(msg.SeqNum, msg.Uid, amsg.SeqNum, amsg.Uid, mbox.Messages)
	r := amsg.GetBody(section)
	if r == nil {
		log.Fatal("Server didn't returned message body")
	}
	// log.Printf("%s", r)
	m, err := message.Read(r)
	if message.IsUnknownCharset(err) {
		// This error is not fatal
		out <- fmt.Sprintln("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}

	dec := new(mime.WordDecoder)
	header := m.Header
	out <- fmt.Sprintln("Date:", header.Get("Date"))
	out <- fmt.Sprintln("From:", header.Get("From"))
	out <- fmt.Sprintln("To:", header.Get("To"))
	val, err := dec.DecodeHeader(header.Get("Subject"))
	out <- fmt.Sprintln("Subject:", val, err)
	// out <- fmt.Sprintln("Subject:", header.Get("Subject"))

	dfs(m, out)
}

func (self *Email) Update(q chan Event) {
	c := self.c
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	from := mbox.Messages - 100
	if from < 0 {
		from = 0
	}
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(
			seqset,
			[]imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags},
			messages)
	}()

	for msg := range messages {
		q <- &NewMessageEvent{msg}
	}
	// q <- &MEvent{"Done"}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}
