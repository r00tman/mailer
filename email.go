package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os/exec"
	"strings"

	"golang.org/x/sync/semaphore"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
	"github.com/mattn/go-runewidth"
)

type Email struct {
	ctx context.Context
	c   *client.Client
	m   *semaphore.Weighted
}

func NewEmail() Email {
	return Email{
		context.TODO(),
		nil,
		semaphore.NewWeighted(1),
	}
}

func (self *Email) IsLocked() bool {
	acquired := self.m.TryAcquire(1)
	if acquired {
		self.m.Release(1)
	}
	return !acquired
}

func (self *Email) Connect(login, password, host string) {
	self.m.Acquire(self.ctx, 1)
	defer self.m.Release(1)
	c, err := client.DialTLS(host, nil)
	self.c = c
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Login(login, password); err != nil {
		log.Fatal(err)
	}
}

func (self *Email) Logout() {
	self.m.Acquire(self.ctx, 1)
	defer self.m.Release(1)
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
		b := ""
		out <- fmt.Sprintln("This is a non-multipart message with type", t)
		out <- fmt.Sprintln("------------------------------------------" + strings.Repeat("-", len(t)))
		if t == "text/html" {
			c := exec.Command("w3m", "-dump", "-T", "text/html")
			c.Stdin = m.Body
			newb, err := c.Output()
			if err == nil {
				b = string(newb)
			}

		} else if t == "text/plain" {
			newb, err := ioutil.ReadAll(m.Body)
			if err == nil {
				b = runewidth.Wrap(string(newb), 85)
			}
		}
		for _, x := range strings.Split(b, "\n") {
			out <- x
		}
	}
}

func (self *Email) ReadMail(msg imap.Message, mbox string, out chan string) {
	self.m.Acquire(self.ctx, 1)
	defer self.m.Release(1)
	c := self.c
	_, err := c.Select(mbox, false)
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

	dec := mime.WordDecoder{charset.Reader}
	header := m.Header
	getDecoded := func(h string) string {
		raw := header.Get(h)
		val, err := dec.DecodeHeader(raw)
		if err != nil {
			val = fmt.Sprint(err)
		}
		return val
	}
	out <- fmt.Sprintln("Date:", getDecoded("Date"))
	out <- fmt.Sprintln("From:", getDecoded("From"))
	out <- fmt.Sprintln("To:", getDecoded("To"))
	out <- fmt.Sprintln("Subject:", getDecoded("Subject"))

	dfs(m, out)
}

func (self *Email) SetReadFlag(msg imap.Message, value bool, out chan *imap.Message) {
	seq := new(imap.SeqSet)
	seq.AddNum(msg.Uid)
	var mode imap.StoreItem
	if value {
		mode = imap.AddFlags
	} else {
		mode = imap.RemoveFlags
	}
	err := self.c.UidStore(seq, mode, []interface{}{imap.SeenFlag}, out)
	if err != nil {
		log.Fatal(err)
	}
}

func (self *Email) Mailboxes(q chan Event) {
	self.m.Acquire(self.ctx, 1)
	defer self.m.Release(1)
	c := self.c
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		q <- NewMailboxEvent(*m)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}

func (self *Email) Update(q chan Event, mboxName string) {
	self.m.Acquire(self.ctx, 1)
	defer self.m.Release(1)
	c := self.c
	mbox, err := c.Select(mboxName, false)
	if err != nil {
		log.Fatal(err)
	}
	from := mbox.Messages - 100
	if mbox.Messages < 100 {
		from = 0
	}
	if mbox.Messages == 0 {
		q <- NewMessageEvent(imap.Message{})
		return
	}
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(
			seqset,
			[]imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid},
			messages)
	}()

	for msg := range messages {
		q <- NewMessageEvent(*msg)
	}
	// q <- &MEvent{"Done"}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	q <- NewMessageEvent(imap.Message{})
}
