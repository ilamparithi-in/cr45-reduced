package store

import (
	"context"
	"errors"

	"github.com/CR45-NITT/cr45-reduced/backend/internal/domain"
)

var (
	ErrClassNotFound    = errors.New("class not found")
	ErrInvalidSlotIdx   = errors.New("invalid slot index")
	ErrInvalidLogin     = errors.New("invalid username or password")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrMissingSecretKey = errors.New("app secret is required")
)

type Repository interface {
	GetResolvedTimetable(ctx context.Context, classID string) (domain.Timetable, error)
	UpsertOverride(ctx context.Context, req domain.UpdateOverrideRequest) error
	DeleteSlot(ctx context.Context, req domain.DeleteSlotRequest) error
	GetPasswordHash(ctx context.Context, username string) (string, error)
}