package limiter

import "context"

type Runner struct {
	limiters []Limiter
}

func NewRunner(limiters ...Limiter) *Runner {
	return &Runner{limiters: limiters}
}

func (lr *Runner) CanProcess(ctx context.Context, userID, resourceID string) bool {
	for _, limiter := range lr.limiters {
		if !limiter.CanProcess(ctx, userID, resourceID) {
			return false
		}
	}
	return true
}

func (lr *Runner) Process(ctx context.Context, userID, resourceID string) bool {
	if !lr.CanProcess(ctx, userID, resourceID) {
		return false
	}
	for _, limiter := range lr.limiters {
		if !limiter.Process(ctx, userID, resourceID) {
			return false
		}
	}
	return true
}

func (lr *Runner) Finish(ctx context.Context, userID, resourceID string) {
	for _, limiter := range lr.limiters {
		limiter.Finish(ctx, userID, resourceID)
	}
}
