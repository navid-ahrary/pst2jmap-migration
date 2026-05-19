package model

import "time"

type Message struct {
	Folder string

	Subject string

	FromName  string
	FromEmail string

	To  []string
	CC  []string
	BCC []string

	MessageID string

	Date time.Time

	Body string

	Headers string

	Attachments []Attachment
}

type Attachment struct {
	FileName    string
	ContentType string
	Data        []byte
}
