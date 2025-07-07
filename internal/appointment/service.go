package appointment

import (
	"context"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) BookAppointment(
	ctx context.Context,
	user domain.User,
	service domain.Service,
	startTime time.Time,
) error {
	appt := domain.Appointment{
		User:      user,
		Service:   service,
		StartTime: startTime,
		EndTime:   startTime.Add(service.Duration),
		Status:    "confirmed",
	}
	return s.repo.Create(ctx, appt)
}

func (s *Service) GetAvailableServices() []domain.Service {
	return []domain.Service{
		{ID: 1, Name: "Swedish Massage", Duration: 60 * time.Minute, Price: 2500},
		{ID: 2, Name: "Deep Tissue", Duration: 90 * time.Minute, Price: 3500},
	}
}
