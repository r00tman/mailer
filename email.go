package main

import (
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os/exec"

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

func dfs(m *message.Entity) {
	if mr := m.MultipartReader(); mr != nil {
		// This is a multipart message
		log.Println("This is a multipart message containing:")
		for {
			p, err := mr.NextPart()

			if err == io.EOF {
				break
			} else if err != nil {
				log.Println(err)
				continue
			}

			t, _, _ := p.Header.ContentType()
			log.Println("A part with type", t)

			dfs(p)
			// b, _ := ioutil.ReadAll(p.Body)
			// log.Println("A part with type", string(b))
		}
	} else {
		t, _, _ := m.Header.ContentType()
		log.Println("This is a non-multipart message with type", t)
		if t == "text/html" {
			c := exec.Command("w3m", "-dump", "-T", "text/html")
			c.Stdin = m.Body
			b, err := c.Output()
			if err == nil {
				log.Println(string(b))
			}

		} else if t == "text/plain" {
			b, _ := ioutil.ReadAll(m.Body)
			log.Println(string(b))
		}
	}
}

func (self *Email) ReadMail(msg *imap.Message) {
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

	log.Println("Last message:")
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
		log.Println("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}

	dec := new(mime.WordDecoder)
	header := m.Header
	log.Println("Date:", header.Get("Date"))
	log.Println("From:", header.Get("From"))
	log.Println("To:", header.Get("To"))
	val, err := dec.DecodeHeader(header.Get("Subject"))
	log.Println("Subject:", val, err)
	// log.Println("Subject:", header.Get("Subject"))

	dfs(m)
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
			[]imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchBody, imap.FetchBodyStructure},
			messages)
	}()

	for msg := range messages {
		q <- &MEvent{msg}
	}
	// q <- &MEvent{"Done"}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}
