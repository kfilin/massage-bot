package telegram

import (
	"strings"
	"testing"
)

// --- RouteCallback tests ---

func TestRouteCallback_CategoryPrefix(t *testing.T) {
	action, matched := RouteCallback("select_category|1")
	if !matched || action != CallbackPrefixCategory {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixCategory, action, matched)
	}
}

func TestRouteCallback_ServicePrefix(t *testing.T) {
	action, matched := RouteCallback("select_service|massage-60")
	if !matched || action != CallbackPrefixService {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixService, action, matched)
	}
}

func TestRouteCallback_DatePrefix(t *testing.T) {
	action, matched := RouteCallback("select_date|2026-07-01")
	if !matched || action != CallbackPrefixDate {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixDate, action, matched)
	}
}

func TestRouteCallback_NavigateMonthPrefix(t *testing.T) {
	action, matched := RouteCallback("navigate_month|2026-08")
	if !matched || action != CallbackPrefixNavigateMonth {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixNavigateMonth, action, matched)
	}
}

func TestRouteCallback_TimePrefix(t *testing.T) {
	action, matched := RouteCallback("select_time|14:30")
	if !matched || action != CallbackPrefixTime {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixTime, action, matched)
	}
}

func TestRouteCallback_CancelApptPrefix(t *testing.T) {
	action, matched := RouteCallback("cancel_appt|abc-123")
	if !matched || action != CallbackPrefixCancelAppt {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixCancelAppt, action, matched)
	}
}

func TestRouteCallback_ConfirmReminderPrefix(t *testing.T) {
	action, matched := RouteCallback("confirm_appt_reminder|appt-1")
	if !matched || action != CallbackPrefixConfirmReminder {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixConfirmReminder, action, matched)
	}
}

func TestRouteCallback_CancelReminderPrefix(t *testing.T) {
	action, matched := RouteCallback("cancel_appt_reminder|appt-1")
	if !matched || action != CallbackPrefixCancelReminder {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixCancelReminder, action, matched)
	}
}

func TestRouteCallback_AdminReplyPrefix(t *testing.T) {
	action, matched := RouteCallback("admin_reply|123456")
	if !matched || action != CallbackPrefixAdminReply {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackPrefixAdminReply, action, matched)
	}
}

func TestRouteCallback_ApproveDraftExact(t *testing.T) {
	action, matched := RouteCallback("approve_draft")
	if !matched || action != CallbackApproveDraft {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackApproveDraft, action, matched)
	}
}

func TestRouteCallback_DiscardDraftExact(t *testing.T) {
	action, matched := RouteCallback("discard_draft")
	if !matched || action != CallbackDiscardDraft {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackDiscardDraft, action, matched)
	}
}

// Exact-match callbacks (no suffix)

func TestRouteCallback_ConfirmBookingExact(t *testing.T) {
	action, matched := RouteCallback("confirm_booking")
	if !matched || action != CallbackConfirmBooking {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackConfirmBooking, action, matched)
	}
}

func TestRouteCallback_CancelBookingExact(t *testing.T) {
	action, matched := RouteCallback("cancel_booking")
	if !matched || action != CallbackCancelBooking {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackCancelBooking, action, matched)
	}
}

func TestRouteCallback_BackToServicesExact(t *testing.T) {
	action, matched := RouteCallback("back_to_services")
	if !matched || action != CallbackBackToServices {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackBackToServices, action, matched)
	}
}

func TestRouteCallback_BackToDateExact(t *testing.T) {
	action, matched := RouteCallback("back_to_date")
	if !matched || action != CallbackBackToDate {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackBackToDate, action, matched)
	}
}

func TestRouteCallback_IgnoreExact(t *testing.T) {
	action, matched := RouteCallback("ignore")
	if !matched || action != CallbackIgnore {
		t.Errorf("expected (%q, true), got (%q, %v)", CallbackIgnore, action, matched)
	}
}

// Edge cases

func TestRouteCallback_TrimsWhitespace(t *testing.T) {
	// Caller trims; ensure the function works on already-trimmed input.
	trimmed := strings.TrimSpace("  select_category|1  ")
	action, matched := RouteCallback(trimmed)
	if !matched || action != CallbackPrefixCategory {
		t.Errorf("expected trimmed data to match, got (%q, %v)", action, matched)
	}
}

func TestRouteCallback_UnknownPrefix(t *testing.T) {
	action, matched := RouteCallback("totally_unknown|xyz")
	if matched {
		t.Errorf("expected no match for unknown data, got (%q, %v)", action, matched)
	}
}

func TestRouteCallback_EmptyString(t *testing.T) {
	action, matched := RouteCallback("")
	if matched {
		t.Errorf("expected no match for empty string, got (%q, %v)", action, matched)
	}
}

// --- RouteTextMessage tests ---
// Each test builds a SessionView literal inline — no helper struct needed.

func TestRouteTextMessage_CommandFallback(t *testing.T) {
	got := RouteTextMessage("/create_appointment John", SessionView{})
	if got != TextActionManualAppointment {
		t.Errorf("expected TextActionManualAppointment, got %v", got)
	}
}

func TestRouteTextMessage_MainMenuButtons(t *testing.T) {
	cases := []struct {
		text   string
		expect TextAction
	}{
		{"🗓 Записаться", TextActionStart},
		{"📅 Мои записи", TextActionMyAppointments},
		{"📄 Мед-карта", TextActionMyRecords},
		{"📤 Загрузить документы", TextActionUpload},
	}
	for _, c := range cases {
		t.Run(c.text, func(t *testing.T) {
			got := RouteTextMessage(c.text, SessionView{})
			if got != c.expect {
				t.Errorf("text=%q: expected %v, got %v", c.text, c.expect, got)
			}
		})
	}
}

func TestRouteTextMessage_AdminReplyState(t *testing.T) {
	s := SessionView{AdminReplyingTo: "123456789"}
	got := RouteTextMessage("Hello patient", s)
	if got != TextActionAdminReply {
		t.Errorf("expected TextActionAdminReply, got %v", got)
	}
}

func TestRouteTextMessage_AwaitingConfirmation_Yes(t *testing.T) {
	s := SessionView{AwaitingConfirmation: true}
	for _, txt := range []string{"подтвердить", "Да", "Д", "yes", "y", "ok", "ок"} {
		t.Run(txt, func(t *testing.T) {
			got := RouteTextMessage(txt, s)
			if got != TextActionConfirmBooking {
				t.Errorf("text=%q: expected TextActionConfirmBooking, got %v", txt, got)
			}
		})
	}
}

func TestRouteTextMessage_AwaitingConfirmation_No(t *testing.T) {
	s := SessionView{AwaitingConfirmation: true}
	for _, txt := range []string{"отменить запись", "нет", "н", "no", "n", "отмена"} {
		t.Run(txt, func(t *testing.T) {
			got := RouteTextMessage(txt, s)
			if got != TextActionCancel {
				t.Errorf("text=%q: expected TextActionCancel, got %v", txt, got)
			}
		})
	}
}

func TestRouteTextMessage_AwaitingConfirmation_Invalid(t *testing.T) {
	s := SessionView{AwaitingConfirmation: true}
	got := RouteTextMessage("what??", s)
	if got != TextActionAskUseButtons {
		t.Errorf("expected TextActionAskUseButtons, got %v", got)
	}
}

func TestRouteTextMessage_SafetyFallbacks(t *testing.T) {
	cases := []struct {
		text   string
		expect TextAction
	}{
		{"Подтвердить", TextActionConfirmBooking},
		{"Отменить запись", TextActionCancel},
		{"Выбрать другую дату", TextActionRestartFlow},
		{"⬅️ Выбрать другую дату", TextActionRestartFlow},
	}
	for _, c := range cases {
		t.Run(c.text, func(t *testing.T) {
			got := RouteTextMessage(c.text, SessionView{})
			if got != c.expect {
				t.Errorf("text=%q: expected %v, got %v", c.text, c.expect, got)
			}
		})
	}
}

func TestRouteTextMessage_Default_NoServiceSet(t *testing.T) {
	got := RouteTextMessage("Some text", SessionView{})
	if got != TextActionAskSelectService {
		t.Errorf("expected TextActionAskSelectService, got %v", got)
	}
}

func TestRouteTextMessage_Default_ServiceSetNoName(t *testing.T) {
	got := RouteTextMessage("Some text", SessionView{HasService: true})
	if got != TextActionNameInput {
		t.Errorf("expected TextActionNameInput, got %v", got)
	}
}

func TestRouteTextMessage_Default_AllSessionSet_ForwardsToAdmins(t *testing.T) {
	got := RouteTextMessage("Some text", SessionView{HasService: true, HasName: true})
	if got != TextActionForwardToAdmins {
		t.Errorf("expected TextActionForwardToAdmins, got %v", got)
	}
}

// Priority: admin reply state beats confirmation state beats default flow.
func TestRouteTextMessage_PriorityOrder(t *testing.T) {
	s := SessionView{
		AdminReplyingTo:      "999",
		AwaitingConfirmation: true,
		HasService:           true,
		HasName:              true,
	}
	got := RouteTextMessage("anything", s)
	if got != TextActionAdminReply {
		t.Errorf("admin reply should win priority, got %v", got)
	}
}
