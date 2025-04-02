package mocks

import (
	"net/http"
	"sync"

	"github.com/denis-k2/relohelper-go/internal/data"
)

type MockUserModel struct {
	mu       sync.Mutex
	Users    map[string]*data.User
}

func NewMockUserModel() *MockUserModel {
	return &MockUserModel{
		Users: make(map[string]*data.User),
	}
}

func (m *MockUserModel) GetForToken(scope string, token string) (*data.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token == "XXXXXXXXXXXXXXXXXXXXXXXXXX" {
		return &data.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Activated: true,
		}, nil
	}
	return nil, nil
}

var Headers = http.Header{
	"Authorization": []string{"Bearer XXXXXXXXXXXXXXXXXXXXXXXXXX"},
}

func (m *MockUserModel) GetByEmail(email string) (*data.User, error) {
	return nil, nil
}

func (m *MockUserModel) Insert(user *data.User) error {
	return nil
}

func (m *MockUserModel) Update(user *data.User) error {
	return nil
}
