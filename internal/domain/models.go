package domain

import "time"

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
}

type Service struct {
	ID       int
	Name     string
	Duration time.Duration
	Price    float64
}

type Appointment struct {
	ID        string
	User      User
	Service   Service
	StartTime time.Time
	EndTime   time.Time
	Status    string // "confirmed", "cancelled"
}
