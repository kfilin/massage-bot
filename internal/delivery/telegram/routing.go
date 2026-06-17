package telegram

import (
	"strings"
)

// TextAction is the resolved routing decision for an incoming text message.
// The OnText handler in bot.go dispatches based on this value.
type TextAction int

const (
	TextActionUnknown TextAction = iota

	// Command fallback: user typed "/create_appointment ..." in OnText.
	TextActionManualAppointment

	// Main menu buttons.
	TextActionStart
	TextActionMyAppointments
	TextActionMyRecords
	TextActionUpload

	// Session-driven routes.
	TextActionAdminReply         // Admin is replying to a patient.
	TextActionConfirmBooking     // User confirmed a pending booking.
	TextActionCancel             // User cancelled / rejected a pending booking.
	TextActionAskUseButtons      // Awaiting confirmation but received unrelated text.
	TextActionRestartFlow        // "Выбрать другую дату" — wipe date and restart.
	TextActionAskSelectService   // No service in session — ask user to pick one.
	TextActionNameInput          // Service picked, awaiting name.
	TextActionForwardToAdmins    // All session fields populated — forward free-text to admins.
)

// SessionView is the minimal session snapshot RouteTextMessage needs to make
// a routing decision. Keeping it as a small struct (rather than the full
// SessionStorage interface) keeps the router pure and trivially testable.
type SessionView struct {
	AdminReplyingTo      string // Telegram ID of patient admin is replying to.
	AwaitingConfirmation bool   // True if booking flow is awaiting yes/no.
	HasService           bool   // True if a service is selected in session.
	HasName              bool   // True if a name has been entered in session.
}

// sessionString safely extracts a string value from a session map. Returns ""
// if missing or wrong type. Used by bot.go to build a SessionView from the
// raw session map without sprinkling type-assertion noise throughout.
func sessionString(session map[string]interface{}, key string) string {
	if v, ok := session[key].(string); ok {
		return v
	}
	return ""
}

// sessionBool safely extracts a bool value from a session map. Returns false
// if missing or wrong type.
func sessionBool(session map[string]interface{}, key string) bool {
	if v, ok := session[key].(bool); ok {
		return v
	}
	return false
}

// sessionHasKey reports whether the session map contains the given key at all
// (regardless of value). RouteTextMessage uses this to detect "service picked"
// vs "no service picked" without caring about the underlying value's type.
func sessionHasKey(session map[string]interface{}, key string) bool {
	_, ok := session[key]
	return ok
}

// RouteCallback inspects the data of an incoming OnCallback event and returns
// which prefix (or exact match) it belongs to. It returns matched=false for
// unknown data so the caller can emit a fallback message.
//
// Caller is expected to pass already-trimmed data (bot.go does
// strings.TrimSpace before invoking).
func RouteCallback(data string) (string, bool) {
	switch {
	case strings.HasPrefix(data, CallbackPrefixCategory):
		return CallbackPrefixCategory, true
	case strings.HasPrefix(data, CallbackPrefixService):
		return CallbackPrefixService, true
	case strings.HasPrefix(data, CallbackPrefixDate):
		return CallbackPrefixDate, true
	case strings.HasPrefix(data, CallbackPrefixNavigateMonth):
		return CallbackPrefixNavigateMonth, true
	case strings.HasPrefix(data, CallbackPrefixTime):
		return CallbackPrefixTime, true
	case data == CallbackBackToServices:
		return CallbackBackToServices, true
	case data == CallbackBackToDate:
		return CallbackBackToDate, true
	case data == CallbackConfirmBooking:
		return CallbackConfirmBooking, true
	case data == CallbackCancelBooking:
		return CallbackCancelBooking, true
	case strings.HasPrefix(data, CallbackPrefixCancelAppt):
		return CallbackPrefixCancelAppt, true
	case strings.HasPrefix(data, CallbackPrefixConfirmReminder):
		return CallbackPrefixConfirmReminder, true
	case strings.HasPrefix(data, CallbackPrefixCancelReminder):
		return CallbackPrefixCancelReminder, true
	case strings.HasPrefix(data, CallbackPrefixAdminReply):
		return CallbackPrefixAdminReply, true
	case data == CallbackApproveDraft:
		return CallbackApproveDraft, true
	case data == CallbackDiscardDraft:
		return CallbackDiscardDraft, true
	case data == CallbackIgnore:
		return CallbackIgnore, true
	}
	return "", false
}

// RouteTextMessage returns the routing decision for an incoming OnText event.
//
// Priority ladder (highest first):
//  1. /create_appointment command fallback
//  2. Main menu buttons
//  3. Admin reply state
//  4. Awaiting confirmation (yes/no/other)
//  5. Safety fallbacks (Подтвердить / Отменить запись / Выбрать другую дату)
//  6. Default name-input / forward-to-admins flow
//
// Caller is expected to pass already-trimmed text.
func RouteTextMessage(text string, s SessionView) TextAction {
	// Priority 1: command fallback.
	if strings.HasPrefix(text, "/create_appointment") {
		return TextActionManualAppointment
	}

	// Priority 2: main menu buttons (always available).
	switch text {
	case "🗓 Записаться":
		return TextActionStart
	case "📅 Мои записи":
		return TextActionMyAppointments
	case "📄 Мед-карта":
		return TextActionMyRecords
	case "📤 Загрузить документы":
		return TextActionUpload
	}

	// Priority 3: admin replying to patient.
	if s.AdminReplyingTo != "" {
		return TextActionAdminReply
	}

	// Priority 4: awaiting confirmation.
	if s.AwaitingConfirmation {
		switch strings.ToLower(text) {
		case "подтвердить", "да", "д", "yes", "y", "ok", "ок":
			return TextActionConfirmBooking
		case "отменить запись", "нет", "н", "no", "n", "отмена":
			return TextActionCancel
		default:
			return TextActionAskUseButtons
		}
	}

	// Priority 5: safety fallbacks.
	switch text {
	case "Подтвердить":
		return TextActionConfirmBooking
	case "Отменить запись":
		return TextActionCancel
	case "Выбрать другую дату", "⬅️ Выбрать другую дату":
		return TextActionRestartFlow
	}

	// Priority 6: default flow based on session completeness.
	if !s.HasService {
		return TextActionAskSelectService
	}
	if !s.HasName {
		return TextActionNameInput
	}
	return TextActionForwardToAdmins
}
