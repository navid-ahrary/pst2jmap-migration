package model

import "time"

type Message struct {
	Folder     string
	Subject    string
	From       string
	To         []string
	Date       time.Time
	Flags      []string
	RawRFC822  []byte
}