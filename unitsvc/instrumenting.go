package unitsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/google/uuid"
	"github.com/studiously/unitsvc/models"
)

func InstrumentingMiddleware(
	requestCount metrics.Counter,
	requestLatency metrics.Histogram,
) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{requestCount, requestLatency, next}
	}
}

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

func (im instrumentingMiddleware) ListUnits(ctx context.Context, classID uuid.UUID) (units []uuid.UUID, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "ListUnits", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.ListUnits(ctx, classID)
}

func (im instrumentingMiddleware) GetUnit(ctx context.Context, unitID uuid.UUID) (unit *models.Unit, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetUnit", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.GetUnit(ctx, unitID)
}

func (im instrumentingMiddleware) CreateUnit(ctx context.Context, classID uuid.UUID, title string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "CreateUnit", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.CreateUnit(ctx, classID, title)
}

func (im instrumentingMiddleware) RenameUnit(ctx context.Context, unitID uuid.UUID, title string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "RenameUnit", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.RenameUnit(ctx, unitID, title)
}

func (im instrumentingMiddleware) DeleteUnit(ctx context.Context, unitID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "DeleteUnit", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.DeleteUnit(ctx, unitID)
}
