package unitsvc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/go-nats"
	"github.com/studiously/unitsvc/models"
)

const (
	SubjCreateUnit = "units.create"
	SubjRenameUnit = "units.rename"
	SubjDeleteUnit = "units.delete"
)

func MessagingMiddleware(nc *nats.Conn) Middleware {
	return func(next Service) Service {
		return messagingMiddleware{nc, next}
	}
}

type messagingMiddleware struct {
	nc   *nats.Conn
	next Service
}

func (mm messagingMiddleware) ListUnits(ctx context.Context, classID uuid.UUID) ([]uuid.UUID, error) {
	return mm.next.ListUnits(ctx, classID)
}

func (mm messagingMiddleware) GetUnit(ctx context.Context, unitID uuid.UUID) (*models.Unit, error) {
	return mm.next.GetUnit(ctx, unitID)
}

func (mm messagingMiddleware) CreateUnit(ctx context.Context, classID uuid.UUID, title string) (err error) {
	defer func() {
		if err == nil {
			cID, _ := classID.MarshalText()
			mm.nc.Publish(SubjCreateUnit, cID)
		}
	}()
	return mm.next.CreateUnit(ctx, classID, title)
}

func (mm messagingMiddleware) RenameUnit(ctx context.Context, unitID uuid.UUID, title string) (err error) {
	defer func() {
		if err == nil {
			uID, _ := unitID.MarshalText()
			mm.nc.Publish(SubjRenameUnit, uID)
		}
	}()
	return mm.next.RenameUnit(ctx, unitID, title)
}

func (mm messagingMiddleware) DeleteUnit(ctx context.Context, unitID uuid.UUID) (err error) {
	defer func() {
		if err == nil {
			// Don't care about errors because worst-case scenario, some data doesn't get deleted.
			// It's inaccessible and thus irrelevant.
			uID, _ := unitID.MarshalText()
			mm.nc.Publish(SubjDeleteUnit, uID)
		}
	}()
	return mm.DeleteUnit(ctx, unitID)
}
