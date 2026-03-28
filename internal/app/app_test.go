package app_test

import (
	"testing"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
)

func TestNewApp(t *testing.T) {
	a, err := app.New(app.Config{DBPath: ":memory:"})
	if err != nil {
		t.Fatalf("app.New failed: %v", err)
	}
	if a.DB == nil {
		t.Error("DB must not be nil")
	}
	if a.Client == nil {
		t.Error("Client must not be nil")
	}
	if a.Scheduler == nil {
		t.Error("Scheduler must not be nil")
	}
}
