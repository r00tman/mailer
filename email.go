package main

import (
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/mattn/go-runewidth"
)

type Email struct {
	c *client.Client
}

func (self *Email) Update(q chan Event) {
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	self.c = c
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	if err := c.Login(USER, PASSWD); err != nil {
		log.Fatal(err)
	}

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
		sender := msg.Envelope.Sender
		sender_str := sender[0].PersonalName
		if len(sender_str) == 0 {
			sender_str = sender[0].MailboxName + "@" + sender[0].HostName
		}
		sender_str = runewidth.Truncate(sender_str, 20, "\u2026")
		sender_str = runewidth.FillRight(sender_str, 20)
		q <- &MEvent{"* " + sender_str + " " + msg.Envelope.Subject + " " + strings.Join(msg.Flags, " ")}
	}
	q <- &MEvent{"Done"}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}
