# Mail Collater
Fetch email messages using IMAP

## Install Go
For Mac:

		brew install golang

Refer to [Installation Guide](http://sourabhbajaj.com/mac-setup/Go/README.html)

## Install Dependencies

		go get "github.com/emersion/go-imap/client"
		go get "github.com/emersion/go-imap"
		go get "github.com/gorilla/mux"


## Spinning up Server
Run

		go install
		mail_collater


## Routes
Fetch mails:

		localhost:8080\mails
