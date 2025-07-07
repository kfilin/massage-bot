package appointment

import (
	"context"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, appt domain.Appointment) error
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]domain.Appointment, error)
}
