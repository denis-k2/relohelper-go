package mocks

import (
	"sync"
)

type Email struct {
	Recipient string
	Subject   string
	PlainBody any
	HTMLBody  string
}

type MockMailer struct {
	mu    sync.Mutex
	Email Email
}

func NewMockMailer() *MockMailer {
	return &MockMailer{}
}

func (m *MockMailer) Send(recipient, templateFile string, data any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Email = Email{
		Recipient: recipient,
		Subject:   "Welcome to Relohelper!",
		PlainBody: data,
		HTMLBody:  templateFile,
	}

	return nil
}
