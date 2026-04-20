package telegram

import (
	"testing"

	"gopkg.in/telebot.v3"
)

// middlewareCtx is a minimal telebot.Context stub for middleware tests.
type middlewareCtx struct {
	telebot.Context
	sender  *telebot.User
	sentMsg string
}

func (m *middlewareCtx) Sender() *telebot.User { return m.sender }
func (m *middlewareCtx) Send(what interface{}, opts ...interface{}) error {
	if s, ok := what.(string); ok {
		m.sentMsg = s
	}
	return nil
}
func (m *middlewareCtx) Respond(resp ...*telebot.CallbackResponse) error { return nil }
func (m *middlewareCtx) Callback() *telebot.Callback                     { return nil }

func TestAdminOnly_NonAdminBlocked(t *testing.T) {
	// Temporarily override the global map so tests are deterministic
	original := adminUserIDs
	adminUserIDs = map[int64]bool{111: true}
	defer func() { adminUserIDs = original }()

	nextCalled := false
	next := telebot.HandlerFunc(func(c telebot.Context) error {
		nextCalled = true
		return nil
	})

	ctx := &middlewareCtx{sender: &telebot.User{ID: 999}} // not in admin list
	handler := AdminOnly(next)
	_ = handler(ctx)

	if nextCalled {
		t.Error("Expected next NOT to be called for non-admin")
	}
	if ctx.sentMsg == "" {
		t.Error("Expected permission-denied message to be sent")
	}
}

func TestAdminOnly_AdminAllowed(t *testing.T) {
	original := adminUserIDs
	adminUserIDs = map[int64]bool{111: true}
	defer func() { adminUserIDs = original }()

	nextCalled := false
	next := telebot.HandlerFunc(func(c telebot.Context) error {
		nextCalled = true
		return nil
	})

	ctx := &middlewareCtx{sender: &telebot.User{ID: 111}} // in admin list
	handler := AdminOnly(next)
	_ = handler(ctx)

	if !nextCalled {
		t.Error("Expected next to be called for admin")
	}
}
