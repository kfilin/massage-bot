package handlers

// Session keys used by the booking flow to stash per-user state
// (selected service, date, time, name, etc.) between callback steps.
const (
	SessionKeyService              = "service"
	SessionKeyDate                 = "date"
	SessionKeyTime                 = "time"
	SessionKeyName                 = "name"
	SessionKeyAwaitingConfirmation = "awaiting_confirmation"
	SessionKeyCategory             = "category" // New for categorized menu
	SessionKeyIsAdminBlock         = "is_admin_block"
	SessionKeyIsAdminManual        = "is_admin_manual"
	SessionKeyAdminReplyingTo      = "admin_replying_to"
	SessionKeyPatientID            = "patient_id" // For manual booking
)
