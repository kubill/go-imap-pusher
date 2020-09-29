package main

import (
	"flag"
	"fmt"
	"log"

	idle "github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
)

var (
	host     string
	username string
	password string
	folder   string
)

func init() {

	flag.StringVar(&host, "h", "imap.gmail.com:993", "imap server host:port")
	flag.StringVar(&username, "u", "username", "your mail account")
	flag.StringVar(&password, "p", "password", "your mail password")
	flag.StringVar(&folder, "f", "INBOX", "select folder")
}

func main() {
	flag.Parse()
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(host, nil)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(username, password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select a mailbox
	if _, err := c.Select(folder, false); err != nil {
		log.Fatal(err)
	}

	idleClient := idle.NewClient(c)

	// Create a channel to receive mailbox updates
	updates := make(chan client.Update)
	c.Updates = updates

	// Start idling
	fmt.Println("Start listen...")
	done := make(chan error, 1)
	go func() {
		done <- idleClient.IdleWithFallback(nil, 0)
	}()

	// Listen for updates
	for {
		select {
		case update := <-updates:
			switch update.(type) {
			case *client.StatusUpdate:
				sts, _ := update.(*client.StatusUpdate)
				log.Println(sts.Status)
			case *client.MailboxUpdate:
				mbx, _ := update.(*client.MailboxUpdate)
				log.Println(mbx.Mailbox)
			case *client.ExpungeUpdate:
				exp, _ := update.(*client.ExpungeUpdate)
				log.Println(exp.SeqNum)
			case *client.MessageUpdate:
				msg, _ := update.(*client.MessageUpdate)
				log.Println(msg.Message)
			}
		case err := <-done:
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Not idling anymore")
			return
		}
	}
}
