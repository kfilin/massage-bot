package keyboards

import (
	"testing"
	"time"
)

func TestNewDatePicker_NotNil(t *testing.T) {
	kb := NewDatePicker()
	if kb == nil {
		t.Fatal("NewDatePicker() returned nil")
	}
}

func TestNewDatePicker_HasMonthHeader(t *testing.T) {
	kb := NewDatePicker()
	// NewDatePicker calls kb.Reply(...) which populates ReplyKeyboard
	if len(kb.ReplyKeyboard) == 0 {
		t.Fatal("NewDatePicker() returned keyboard with no reply rows")
	}
	now := time.Now()
	monthYear := now.Format("January 2006")
	firstRow := kb.ReplyKeyboard[0]
	if len(firstRow) == 0 {
		t.Fatal("First reply row is empty")
	}
	found := false
	for _, btn := range firstRow {
		if btn.Text == monthYear {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected first row to contain month header %q, got %v", monthYear, firstRow)
	}
}

func TestNewDatePicker_HasWeekdayRow(t *testing.T) {
	kb := NewDatePicker()
	if len(kb.ReplyKeyboard) < 2 {
		t.Fatal("Expected at least 2 reply rows (header + weekdays)")
	}
	weekdayRow := kb.ReplyKeyboard[1]
	if len(weekdayRow) != 7 {
		t.Errorf("Expected 7 weekday buttons, got %d", len(weekdayRow))
	}
	expectedDays := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	for i, day := range expectedDays {
		if weekdayRow[i].Text != day {
			t.Errorf("Expected weekday[%d] = %q, got %q", i, day, weekdayRow[i].Text)
		}
	}
}
