package main

import (
	"log"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"net/mail"
)

// Feed Username and Password
const (
	Address = "imap.gmail.com:993"
	Username = "user@gmail.com"
	Password = "password"
)

type MessageDetail struct {
	from string `json:"from"`
	to string `json:"to"`
	cc string `json:"cc"`
	subject string `json:"subject"`
	bcc string `json:"bcc"`
}

func Getlast4Mails(resp http.ResponseWriter, req *http.Request){
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(Username, Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func () {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	eMessages := []string{}
	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []string{imap.EnvelopeMsgAttr}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		eMessages = append(eMessages, msg.Envelope.Subject)
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
	json.NewEncoder(resp).Encode(eMessages)
}

func FetchMails(resp http.ResponseWriter, req *http.Request){
	resp.Header().Set("Content-Type", "application/json")
	c, err := client.DialTLS(Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(Username, Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	compiledMessages := CompileMessages(c)
	log.Println("Done!")
	log.Println(compiledMessages)
	// TODO: Build Json response for array of structs
	json.NewEncoder(resp).Encode(compiledMessages)
}

func CompileMessages(c *client.Client) []MessageDetail {
		// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("------------")
	// Get the last message
	if mbox.Messages == 0 {
		log.Fatal("No message in mailbox")
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(mbox.Messages, mbox.Messages)
	// Get the whole message body
	attrs := []string{"BODY[]"}
	// Extract Last message
	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, attrs, messages)
	}()
	log.Println("========== Compile Messages ============")

	// TODO: Loop through messages
	// for msg := range messages {
	msg := <-messages

	r := msg.GetBody("BODY[]")
	if r == nil {
		log.Fatal("Server didn't returned message body")
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}
	messageDetails := make([]MessageDetail, 1)
	// }
	messageDetails[0] = FormatMessage(m)
	return messageDetails
}

func FormatMessage(msg *mail.Message) MessageDetail {
	log.Println("============== Format Message ==============")
	header := msg.Header
	formattedDetail := MessageDetail{ from: header.Get("From"),
																		to: header.Get("To"),
																		subject: header.Get("Subject"),
																		cc: header.Get("Cc"),
																		bcc: header.Get("Bcc") }
	return formattedDetail
}

func main() {
  router := mux.NewRouter()

  // Define routes
  router.HandleFunc("/last_mails", Getlast4Mails).Methods("GET")
  router.HandleFunc("/mails", FetchMails).Methods("GET")

	// Spinning up server
  log.Fatal(http.ListenAndServe(":8080", router))
}
