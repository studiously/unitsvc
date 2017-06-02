package unitsvc

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/ory/hydra/oauth2"
	"github.com/studiously/introspector"
	"github.com/studiously/unitsvc/models"
)

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) ListUnits(ctx context.Context, classID uuid.UUID) (units []uuid.UUID, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"action", "ListUnits",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return mw.next.ListUnits(ctx, classID)
}

func (mw loggingMiddleware) GetUnit(ctx context.Context, unitID uuid.UUID) (unit *models.Unit, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"action", "GetUnit",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"unit", unitID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return mw.next.GetUnit(ctx, unitID)
}

func (mw loggingMiddleware) CreateUnit(ctx context.Context, classID uuid.UUID, title string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"action", "CreateUnit",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"title", title,
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return mw.next.CreateUnit(ctx, classID, title)
}

func (mw loggingMiddleware) RenameUnit(ctx context.Context, unitID uuid.UUID, title string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"action", "RenameUnit",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"unit", unitID.String(),
			"title", title,
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return mw.next.RenameUnit(ctx, unitID, title)
}

func (mw loggingMiddleware) DeleteUnit(ctx context.Context, unitID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"action", "DeleteUnit",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"unit", unitID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return mw.next.DeleteUnit(ctx, unitID)
}

func cli(ctx context.Context) string {
	return ctx.Value(introspector.OAuth2IntrospectionContextKey).(oauth2.Introspection).ClientID
}
