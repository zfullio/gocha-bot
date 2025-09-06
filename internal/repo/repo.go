package repo

import (
	"context"
	"errors"
	"time"

	"gocha/internal/entity"
)

type Repository interface {
	NewPet(ctx context.Context, p *entity.Pet, chatID int) error
	SavePet(ctx context.Context, p *entity.Pet, chatID int) error
	LoadPet(ctx context.Context, chatID int) (*entity.Pet, error)
	GetChats(ctx context.Context) ([]int, error)

	GetLastAlert(ctx context.Context, chatID int, alertType string) (time.Time, error)
	UpdateLastAlert(ctx context.Context, chatID int, alertType string, now time.Time) error
}

var ErrPetNotFound = errors.New("питомец не найден")
