package unitsvc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/studiously/unitsvc/models"
)

var (
	ErrNotFound = errors.New("the requested resource could not be found, or the user is not allowed to view it")
)

type Middleware func(Service) Service

type Service interface {
	ListUnits(ctx context.Context, classID uuid.UUID) ([]uuid.UUID, error)
	GetUnit(ctx context.Context, unitID uuid.UUID) (*models.Unit, error)
	CreateUnit(ctx context.Context, classID uuid.UUID, title string) error
	RenameUnit(ctx context.Context, unitID uuid.UUID, title string) error
	DeleteUnit(ctx context.Context, unitID uuid.UUID) error
}
