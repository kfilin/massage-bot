package handlers

import (
	"testing"

	"gopkg.in/telebot.v3"
)

func TestAdminHandlers(t *testing.T) {
	adminID := "999999"

	tests := []struct {
		name          string
		handlerMethod func(h *BookingHandler, c telebot.Context) error
		userID        int64
		args          []string
		adminIDs      []string
		setupRepo     func(r *mockRepository)
		wantMsg       string
		wantErr       bool
	}{
		// HandleBan Tests
		{
			name:          "Ban User - Admin Success",
			handlerMethod: (*BookingHandler).HandleBan,
			userID:        999999,
			args:          []string{"123456"},
			adminIDs:      []string{adminID},
			setupRepo: func(r *mockRepository) {
				// No setup needed, BanUser just sets map
			},
			wantMsg: "заблокирован",
			wantErr: false,
		},
		{
			name:          "Ban User - Not Admin",
			handlerMethod: (*BookingHandler).HandleBan,
			userID:        123456,
			args:          []string{"111111"},
			adminIDs:      []string{adminID},
			wantMsg:       "Доступ запрещен",
			wantErr:       false,
		},
		{
			name:          "Ban User - Missing Args",
			handlerMethod: (*BookingHandler).HandleBan,
			userID:        999999,
			args:          []string{},
			adminIDs:      []string{adminID},
			wantMsg:       "Использование",
			wantErr:       false,
		},

		// HandleUnban Tests
		{
			name:          "Unban User - Admin Success",
			handlerMethod: (*BookingHandler).HandleUnban,
			userID:        999999,
			args:          []string{"123456"},
			adminIDs:      []string{adminID},
			setupRepo: func(r *mockRepository) {
				_ = r.BanUser("123456")
			},
			wantMsg: "разблокирован",
			wantErr: false,
		},

		// HandleStatus Tests
		{
			name:          "Status - Admin Success",
			handlerMethod: (*BookingHandler).HandleStatus,
			userID:        999999,
			adminIDs:      []string{adminID},
			setupRepo:     nil,
			wantMsg:       "Статус бота",
			wantErr:       false,
		},
		{
			name:          "Status - Not Admin",
			handlerMethod: (*BookingHandler).HandleStatus,
			userID:        123456,
			adminIDs:      []string{adminID},
			wantMsg:       "только администраторам",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			mockSession := newMockSessionStorage()
			mockApptService := &mockAppointmentService{}

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			handler := NewBookingHandler(
				mockApptService,
				mockSession,
				tt.adminIDs,
				"",
				nil,
				mockRepo,
				"",
				"",
			)

			ctx := &mockContext{
				sender: &telebot.User{ID: tt.userID},
				args:   tt.args,
			}

			err := tt.handlerMethod(handler, ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Handler error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantMsg != "" {
				s := ctx.sentMsg
				if !contains(s, tt.wantMsg) {
					t.Errorf("Expected msg containing %q, got %q", tt.wantMsg, s)
				}
			}
		})
	}
}

func TestHandleBlock(t *testing.T) {
	adminID := "999999"

	tests := []struct {
		name     string
		userID   int64
		adminIDs []string
		wantMsg  string
	}{
		{
			name:     "Admin Access Granted",
			userID:   999999,
			adminIDs: []string{adminID},
			wantMsg:  "Блокировка времени",
		},
		{
			name:     "User Access Denied",
			userID:   123456,
			adminIDs: []string{adminID},
			wantMsg:  "доступна только администраторам",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			mockSession := newMockSessionStorage()

			handler := NewBookingHandler(
				nil,
				mockSession,
				tt.adminIDs,
				"",
				nil,
				mockRepo,
				"",
				"",
			)

			ctx := &mockContext{
				sender: &telebot.User{ID: tt.userID},
			}

			_ = handler.HandleBlock(ctx)

			if !contains(ctx.sentMsg, tt.wantMsg) {
				t.Errorf("Expected msg containing %q, got %q", tt.wantMsg, ctx.sentMsg)
			}
		})
	}
}

func (m *mockRepository) DeleteAppointment(appointmentID string) error {
	return nil
}
