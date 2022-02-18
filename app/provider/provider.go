package provider

import (
	"net/http"
)

// ID provider enum
type ID string

type Provider struct {
	ID ID
	Client *http.Client
}

// enum of all provider ids
const (
	PIDMailgun  ID = "mailgun"
	PIDTelegram ID = "telegram"
)

type Interface interface {
	Send() error
	GetID() ID
}

func (s *Provider) GetID() ID {
	return s.ID
}