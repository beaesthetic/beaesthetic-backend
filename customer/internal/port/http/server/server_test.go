package server

import "testing"

func TestNewRegistersCompatibleRoutes(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("New panicked while registering routes: %v", recovered)
		}
	}()

	_ = New(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil)
}
