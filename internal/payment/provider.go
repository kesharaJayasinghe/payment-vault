package payment

import (
	"errors"
	"math/rand"
	"time"
)

// Define the interface for payment providers
type Provider interface {
	Charge(amount float64, currency string) (string, error)
}

// Simulate a 3rd party API (eg. Stripe)
type MockProvider struct {}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// Simulate a network call (sometimes fails)
func (m *MockProvider) Charge(amount float64, currency string) (string, error) {
	// Simulate network latency
	time.Sleep(500 * time.Millisecond)

	// Simulate random failure
	if rand.Intn(10) == 0 {
		return "", errors.New("network_timeout_simulated")
	}

	// Simulate Declined card payment
	if rand.Intn(10) == 1 {
		return "", errors.New("card_declined")
	}

	// Successful charge (return a mock transaction ID)
	return "txn_" + generateRandomString(12), nil
}

func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}