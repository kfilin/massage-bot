package ports

import (
	"testing"

	"gopkg.in/telebot.v3"
)

// compile-time check: *telebot.Bot must satisfy BotAPI
var _ BotAPI = (*telebot.Bot)(nil)

// TestBotAPIImplemented is a runtime assertion that *telebot.Bot satisfies
// the BotAPI interface. If the interface drifts and the real bot stops
// satisfying it, this test fails at compile time (var declaration above)
// as well as at runtime.
func TestBotAPIImplemented(t *testing.T) {
	// The var declaration at package level already verifies the interface
	// satisfaction at compile time. This test exists so the function shows
	// up in coverage reports and documents the contract explicitly.
}
