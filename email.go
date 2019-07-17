package main

import (
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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
			[]imap.FetchItem{imap.FetchEnvelope},
			messages)
	}()

	for msg := range messages {
		q <- &MEvent{"* " + msg.Envelope.Subject}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}
