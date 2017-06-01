package unitsvc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/studiously/classsvc/classsvc"
	"github.com/studiously/introspector"
	"github.com/studiously/svcerror"
	"github.com/studiously/unitsvc/codes"
	"github.com/studiously/unitsvc/models"
)

var (
	ErrNotFound = svcerror.New(codes.NotFound, "the requested resource could not be found, or the user is not allowed to view it")
)

func New(db *sql.DB, cs classsvc.Service) Service {
	return &postgresService{
		db,
		cs,
	}
}

type postgresService struct {
	*sql.DB
	cs classsvc.Service
}

func (s *postgresService) ListUnits(ctx context.Context, classID uuid.UUID) ([]uuid.UUID, error) {
	classes, err := s.cs.ListClasses(ctx)
	if err != nil {
		return nil, err
	}
	for _, class := range classes {
		if class == classID {
			units, err := models.UnitsByClassID(s, classID)
			if err != nil {
				return nil, err
			}
			results := make([]uuid.UUID, len(units))
			for _, u := range units {
				results = append(results, u.ID)
			}
			return results, nil
		}
	}
	return nil, ErrNotFound
}

func (s *postgresService) CreateUnit(ctx context.Context, classID uuid.UUID, title string) error {
	classes, err := s.cs.ListClasses(ctx)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}
	for _, class := range classes {
		if class == classID {
			unit := &models.Unit{
				ID:      uuid.New(),
				ClassID: classID,
				Title:   title,
			}
			return unit.Save(s)
		}
	}
	return ErrNotFound
}

func (s *postgresService) RenameUnit(ctx context.Context, unitID uuid.UUID, title string) error {
	unit, err := models.UnitByID(s, unitID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}
	classes, err := s.cs.ListClasses(ctx)
	if err != nil {
		return err
	}
	for _, class := range classes {
		if class == unit.ClassID {
			unit.Title = title
			return unit.Save(s)
		}
	}
	return ErrNotFound
}

func (s *postgresService) DeleteUnit(ctx context.Context, unitID uuid.UUID) error {
	unit, err := models.UnitByID(s, unitID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}
	classes, err := s.cs.ListClasses(ctx)
	if err != nil {
		return err
	}
	for _, class := range classes {
		if class == unit.ClassID {
			return unit.Delete(s)
		}
	}
	return ErrNotFound
}

func (s *postgresService) GetUnit(ctx context.Context, unitID uuid.UUID) (*models.Unit, error) {
	unit, err := models.UnitByID(s, unitID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	classes, err := s.cs.ListClasses(ctx)
	if err != nil {
		return nil, err
	}
	for _, class := range classes {
		if class == unit.ClassID {
			return unit, nil
		}
	}
	return nil, ErrNotFound
}

func subj(ctx context.Context) uuid.UUID {
	return ctx.Value(introspector.SubjectContextKey).(uuid.UUID)
}
